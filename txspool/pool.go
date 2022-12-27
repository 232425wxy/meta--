package txspool

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/common/clist"
	"github.com/232425wxy/meta--/common/cmap"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/sha256"
	"github.com/232425wxy/meta--/proto/pbabci"
	"github.com/232425wxy/meta--/proxy"
	"github.com/232425wxy/meta--/types"
	"sync"
	"sync/atomic"
)

type TxsPool struct {
	cfg               *config.TxsPoolConfig
	height            int64
	txsBytes          int64
	notifiedAvailable bool
	txsAvailable      chan struct{}
	mu                sync.RWMutex
	txs               *clist.List
	txsMap            *cmap.CMap // 用于快速定位到存储在链表里的交易元素 hash(tx) -> *clist.Element
	proxyApp          *proxy.AppConnTxsPool
	metrics           *Metrics
}

func NewTxsPool(cfg *config.TxsPoolConfig, proxyApp *proxy.AppConnTxsPool, height int64) *TxsPool {
	pool := &TxsPool{
		cfg:               cfg,
		height:            height,
		txsBytes:          0,
		notifiedAvailable: false,
		txsAvailable:      make(chan struct{}, 1),
		mu:                sync.RWMutex{},
		txs:               clist.NewList(),
		txsMap:            cmap.NewCap(),
		proxyApp:          proxyApp,
		metrics:           TxsPoolMetrics(),
	}
	return pool
}

func (p *TxsPool) Lock() {
	p.mu.Lock()
}

func (p *TxsPool) Unlock() {
	p.mu.Unlock()
}

// Size ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Size 返回目前交易池里的交易个数。
func (p *TxsPool) Size() int {
	return p.txs.Size()
}

// TxsBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// TxsBytes 返回目前交易池里所有交易加一起的大小，单位是字节。
func (p *TxsPool) TxsBytes() int {
	return int(atomic.LoadInt64(&p.txsBytes))
}

// TxsHead ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// TxsHead 返回交易池里第一个交易。
func (p *TxsPool) TxsHead() *clist.Element {
	return p.txs.Head()
}

// WaitTxs ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// WaitTxs 返回一个被阻塞的通道，如果交易池里有交易数据，那么该通道将不再阻塞，这个方法用来将交易池里的
// 数据广播给其他节点。
func (p *TxsPool) WaitTxs() <-chan struct{} {
	return p.txs.WaitChan()
}

// CheckTx ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// CheckTx 首先本地交易池会检查池子是否已满，如果满了的话，就返回错误，如果没满，则将交易数据
// 交给代理应用去检查，例如在key-value数据库里，会检查该笔交易是否已在数据库里被存储，如果已经
// 被存储过，则检查不会被通过，否则就让它通过吧。
func (p *TxsPool) CheckTx(tx types.Tx, sender crypto.ID) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.txsMap.Has(txKey(tx)) {
		// 从其他节点那里收到了我们池子里已经存在的交易，记录一下是谁发来的这个交易数据，
		// 也就是说如果我从很多节点那里收到了这个交易数据，那么就代表我不需要再把这个交
		// 易数据发送给这些节点了。
		elem := p.txsMap.Get(txKey(tx))
		element := elem.(*clist.Element)
		ptx := element.Value.(*poolTx)
		ptx.senders.Set(string(sender), struct{}{})
		return errors.New("tx already in pool")
	}
	if p.isFull() {
		return errors.New("txs pool has been full")
	}
	if len(tx) > p.cfg.MaxTxBytes {
		return fmt.Errorf("single tx is to large: %d > %d", len(tx), p.cfg.MaxTxBytes)
	}
	res := p.proxyApp.CheckTx(pbabci.RequestCheckTx{Tx: tx})
	if !res.OK {
		return errors.New("check tx is not passed")
	}

	ptx := &poolTx{
		tx:      tx,
		height:  p.height,
		senders: cmap.NewCap(),
	}
	ptx.senders.Set(string(sender), struct{}{}) // 记录一下是谁发来的这个交易数据
	p.addTx(ptx)
	p.notifyTxsAvailable() // 通知其他模块说交易池里有交易数据了
	p.metrics.Size.Set(float64(p.Size()))
	return nil
}

// TxsAvailable ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// TxsAvailable 方法返回一个被阻塞的通道，但是只要往交易池里添加一条交易数据，通道的阻塞状态就会被解除，
// 那么共识模块就可以提取交易池里的交易数据打包成区块了，这里倒不是说交易池里只要有一条数据就打包一个区块，
// 因为共识模块有超时设置，所以这段超时时间会给交易池收集更多的交易数据，超时时间一到，再来打包区块，那么就
// 可以获得比较多的交易数据了。
func (p *TxsPool) TxsAvailable() <-chan struct{} {
	return p.txsAvailable
}

// ReapMaxBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ReapMaxBytes 从交易池里获取最多maxBytes大小的交易数据集合。
func (p *TxsPool) ReapMaxBytes(maxBytes int) types.Txs {
	p.mu.RLock()
	defer p.mu.RUnlock()
	txs := make([]types.Tx, 0)
	size := 0
	for elem := p.txs.Head(); elem != nil; elem = elem.Next() {
		ptx := elem.Value.(*poolTx)
		txs = append(txs, ptx.tx)
		size += len(ptx.tx)
		if size >= maxBytes {
			return txs
		}
	}
	return txs
}

// Update ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Update 方法在新区块被commit后调用，该方法的入参txs表示的是刚刚被commit的新区块里包含的交易数据，那么既然
// 这个区块都被commit了，那么就应该将交易池里已经被commit的交易txs删除掉。
func (p *TxsPool) Update(height int64, txs types.Txs) {
	p.height = height
	p.notifiedAvailable = false
	for _, tx := range txs {
		// 从交易池里删除掉已经被提交的交易数据
		if elem := p.txsMap.Get(txKey(tx)); elem != nil {
			p.removeTx(tx, elem.(*clist.Element))
		}
	}
	if p.Size() > 0 {
		p.notifyTxsAvailable()
	}
	p.metrics.Size.Set(float64(p.Size()))
}

func (p *TxsPool) addTx(ptx *poolTx) {
	elem := p.txs.Push(ptx)
	p.txsMap.Set(txKey(ptx.tx), elem)
	atomic.AddInt64(&p.txsBytes, int64(len(ptx.tx)))
	p.metrics.TxsSizeBytes.Observe(float64(len(ptx.tx)))
}

func (p *TxsPool) removeTx(tx types.Tx, elem *clist.Element) {
	p.txs.Remove(elem)
	p.txsMap.Delete(txKey(tx))
	elem.DetachPrev()
	atomic.AddInt64(&p.txsBytes, int64(-len(tx)))
}

// notifyTxsAvailable ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// notifyTxsAvailable 方法通知共识模块交易池里还有数据可以提取。
func (p *TxsPool) notifyTxsAvailable() {
	if p.Size() == 0 {
		panic("notified txs available but txs pool is empty")
	}
	if !p.notifiedAvailable {
		p.notifiedAvailable = true
		select {
		case p.txsAvailable <- struct{}{}:
		default:
		}
	}
}

// isFull ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// isFull 判断交易池是否已满。
func (p *TxsPool) isFull() bool {
	return p.Size() >= p.cfg.MaxSize
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量：存储在交易池里的交易

type poolTx struct {
	tx      types.Tx
	senders *cmap.CMap
	height  int64
}

// txKey ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// txKey 计算交易的sha256哈希值。
func txKey(tx types.Tx) string {
	h := sha256.Sum(tx)
	return hex.EncodeToString(h[:])
}
