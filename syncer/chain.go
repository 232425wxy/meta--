package syncer

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/types"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxPendingNum        = 500
	maxTotalRequesters   = 500
	maxPendingNumPerPeer = 20
	peerTimeout          = 15 * time.Second
)

type BlockRequest struct {
	Height int64
	PeerID crypto.ID
}

type requester struct {
	chain  *Blockchain
	peerID crypto.ID
	height int64 // 要请求的区块的高度
	block  *types.Block
	gotCh  chan struct{}
	redoCh chan crypto.ID
	mu     sync.Mutex
	done   chan struct{}
}

type peer struct {
	isTimeout  bool
	pendingNum int32
	height     int64
	chain      *Blockchain
	id         crypto.ID
	timeout    *time.Timer
}

type peerError struct {
	err    error
	peerID crypto.ID
}

type Blockchain struct {
	service.BaseService
	startTime     time.Time
	height        int64 // 此时区块链高度
	maxPeerHeight int64 // 节点中最高的区块高度
	peers         map[crypto.ID]*peer
	requestsCh    chan BlockRequest
	requesters    map[int64]*requester
	pendingNum    int32 // 表示从其他节点处要请求的区块数量，每从其他节点处获得一个区块，该字段减1，每增加一个请求区块的请求，该字段加1
	errorsCh      chan peerError
	mu            sync.Mutex
}

func NewBlockchain(start int64, requestsCh chan BlockRequest, errorsCh chan peerError) *Blockchain {
	return &Blockchain{
		BaseService:   *service.NewBaseService(nil, "Syncer"),
		height:        start,
		maxPeerHeight: 0,
		peers:         make(map[crypto.ID]*peer),
		requestsCh:    requestsCh,
		requesters:    make(map[int64]*requester),
		pendingNum:    0,
		errorsCh:      errorsCh,
		mu:            sync.Mutex{},
	}
}

func (bc *Blockchain) Start() error {
	go bc.requestRoutine()
	bc.startTime = time.Now()
	return bc.BaseService.Start()
}

func (bc *Blockchain) GetStatus() (height int64, pendingNum int32, requestersNum int) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.height, atomic.LoadInt32(&bc.pendingNum), len(bc.requesters)
}

func (bc *Blockchain) AddBlock(peerID crypto.ID, block *types.Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	r := bc.requesters[block.Header.Height]
	if r == nil {
		bc.Logger.Info("peer sent us a block that we didn't expect", "peer_id", peerID, "current_height", bc.height, "block_height", block.Header.Height)
		return
	}
	if r.setBlock(block, peerID) {
		atomic.AddInt32(&bc.pendingNum, -1)
		p := bc.peers[peerID]
		if p != nil {
			p.pendingNum--
			if p.pendingNum == 0 {
				p.timeout.Stop()
			} else {
				p.resetTimeout()
			}
		}
	}
}

func (bc *Blockchain) PickTwoBlocks() (first *types.Block, second *types.Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if r := bc.requesters[bc.height]; r != nil { // 我们这个链的高度
		first = r.getBlock()
	}
	if r := bc.requesters[bc.height+1]; r != nil {
		second = r.getBlock()
	}
	return first, second
}

func (bc *Blockchain) PopRequest() {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if r := bc.requesters[bc.height]; r != nil {
		close(r.done)
		delete(bc.requesters, bc.height)
		bc.height++
	} else {
		panic(fmt.Sprintf("expected requester to pop, got nothing at height: %d", bc.height))
	}
}

func (bc *Blockchain) MaxPeerHeight() int64 {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	return bc.maxPeerHeight
}

func (bc *Blockchain) SetPeerUpHeight(peerID crypto.ID, up int64) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	p := bc.peers[peerID]
	if p != nil {
		p.height = up
	} else {
		p = &peer{
			isTimeout:  false,
			pendingNum: 0,
			height:     up,
			chain:      bc,
			id:         peerID,
			timeout:    nil,
		}
		bc.peers[peerID] = p
	}
	if up > bc.maxPeerHeight {
		bc.maxPeerHeight = up
	}
}

func (bc *Blockchain) RemovePeer(peerID crypto.ID) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.removePeer(peerID)
}

func (bc *Blockchain) RedoRequest(height int64) crypto.ID {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	r := bc.requesters[height]
	peerID := r.getPeerID()
	if peerID != "" {
		bc.removePeer(peerID)
	}
	return peerID
}

func (bc *Blockchain) IsCaughtUp() bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(bc.peers) == 0 {
		return false
	}
	receivedBlockOrTimeout := bc.height > 0 || time.Since(bc.startTime) > 5*time.Second
	ourChainIsLongestAmongPeers := bc.maxPeerHeight == 0 || bc.height >= (bc.maxPeerHeight-1)
	return receivedBlockOrTimeout && ourChainIsLongestAmongPeers
}

func (bc *Blockchain) requestRoutine() {
	for {
		if !bc.IsRunning() {
			return
		}
		_, pendingNum, requestersNum := bc.GetStatus()
		switch {
		case pendingNum >= maxPendingNum:
			time.Sleep(2 * time.Millisecond)
			bc.removeTimeoutPeers()
		case requestersNum >= maxTotalRequesters:
			time.Sleep(2 * time.Millisecond)
			bc.removeTimeoutPeers()
		default:
			// 请求更多的区块
			bc.makeNextRequester()
		}
	}
}

func (bc *Blockchain) removeTimeoutPeers() {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for _, p := range bc.peers {
		if p.isTimeout {
			bc.removePeer(p.id)
		}
	}
}

// removePeer 删除指定的peer，首先遍历一遍请求队列，如果请求队列里有请求的对象是我们要删除的peer，
// 则再次向该peer发送请求，之后就是将该peer删除掉。
func (bc *Blockchain) removePeer(id crypto.ID) {
	for _, r := range bc.requesters {
		if r.peerID == id { // 删除这个节点之前，再次尝试看能不能执行请求
			select {
			case r.redoCh <- id:
			default:
			}
		}
	}
	p, ok := bc.peers[id]
	if ok {
		if p.timeout != nil {
			p.timeout.Stop()
		}
		delete(bc.peers, id)
		if p.height == bc.maxPeerHeight {
			// 在这里发现要被删除的节点的区块高度竟然等于我已知的最大区块高度，
			// 说明我可能已经落后了
			bc.updateMaxPeerHeight()
		}
	}
}

func (bc *Blockchain) updateMaxPeerHeight() {
	var max int64
	for _, p := range bc.peers {
		if p.height > max {
			max = p.height
		}
	}
	bc.maxPeerHeight = max
}

func (bc *Blockchain) makeNextRequester() {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	nextHeight := bc.height + int64(len(bc.requesters))
	if nextHeight > bc.maxPeerHeight {
		return
	}
	r := &requester{
		chain:  bc,
		peerID: "",
		height: nextHeight,
		block:  nil,
		gotCh:  make(chan struct{}, 1),
		redoCh: make(chan crypto.ID, 1),
		done:   make(chan struct{}),
	}
	bc.requesters[nextHeight] = r
	atomic.AddInt32(&bc.pendingNum, 1)
	go r.requestRoutine()
}

// pickPeer 方法接受一个参数height，表示某个区块的高度，该方法会从网络中其他节点处寻找到一个拥有
// height高度的节点。
func (bc *Blockchain) pickPeer(height int64) *peer {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	for _, p := range bc.peers {
		if p.isTimeout {
			bc.removePeer(p.id)
			continue
		}
		if p.pendingNum > maxPendingNumPerPeer {
			continue
		}
		if height > p.height {
			continue
		}
		p.pendingNum++
		return p
	}
	return nil
}

// sendError 这个方法是比较严苛的，当节点因为执行请求任务超时时，会认为任务执行者有问题，并将该问题汇报给
// Reactor，然后中断与任务执行者之间的连接。
func (bc *Blockchain) sendError(err error, peerID crypto.ID) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if !bc.IsRunning() {
		return
	}
	select {
	case bc.errorsCh <- peerError{err: err, peerID: peerID}:
	default:
		go func() { bc.errorsCh <- peerError{err: err, peerID: peerID} }()
	}
}

// sendRequest 方法接受两个参数，第一个参数height表示我们所需要的区块的高度，第二个参数peerID
// 表示我们这个需要特定高度区块的请求会发送给谁，由谁来回应我们的这个请求；然后我们将这两个参数组装
// 成一个BlockRequest，将其发送给请求队列通道，由Reactor去将通道里的请求取出发送给应该对该请求
// 作出回应的节点。
func (bc *Blockchain) sendRequest(height int64, peerID crypto.ID) {
	if !bc.IsRunning() {
		return
	}
	select {
	case bc.requestsCh <- BlockRequest{Height: height, PeerID: peerID}:
	default:
		go func() { bc.requestsCh <- BlockRequest{Height: height, PeerID: peerID} }()
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// requester

// requestRoutine 不断的从周围节点中寻找一个能够帮我完成requester请求的节点，找到的话，就继续等待
// 这个节点完成requester请求，直到请求任务被完成才退出。
func (r *requester) requestRoutine() {
LOOP:
	for {
		// 选择一个节点发送请求
		var p *peer

	PICKPEERLOOP:
		for {
			if !r.chain.IsRunning() {
				return
			}
			// 当前的请求是需要一个高度为r.height的区块，现在我们需要选一个拥有此区块节点作为我们发出请求的对象
			p = r.chain.pickPeer(r.height)
			if p == nil {
				time.Sleep(2 * time.Millisecond)
				continue PICKPEERLOOP
			}
			break PICKPEERLOOP
		}
		r.mu.Lock()
		r.peerID = p.id // 让r.peerID等于p.id，表示当前这个请求会发送给p这个节点，并由它来完成
		if p.pendingNum == 0 {
			p.resetTimeout() // 因为节点完成任务需要时间，所以我们就给这个节点设一个任务完成的期限
		}
		r.mu.Unlock()
		r.chain.sendRequest(r.height, p.id)

	WAITLOOP:
		for {
			select {
			case <-r.chain.WaitStop():
				return
			case <-r.done:
				return
			case peerID := <-r.redoCh:
				if peerID == r.peerID {
					// 这里判断请求的对象是不是重做通道里出来的peerID，如果是，则该请求的请求对象并没有完成任务，那么
					// 我们通过reset方式将请求对象重置，以便将来重新确定一个请求对象，所以这里的“重做”逻辑反映了不在
					// 一棵树上吊死。
					r.reset()
					continue LOOP
				} else {
					continue WAITLOOP
				}
			case <-r.gotCh:
				// 我们获得了一个区块
				continue WAITLOOP
			}
		}
	}
}

// reset 重置请求任务信息，如果该请求已经有了回应（有节点给自己发来了区块），则给pendingNum减一，
// 因为当初在收到block的时候，我们给pendingNum加一了。
func (r *requester) reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.block != nil {
		// 之所以在block不为空的情况下，将待做任务数量加一，是因为该block是不正确的，
		//而之前我们在设置block的时候，将待做任务数量减一了，加一是为了弥补
		atomic.AddInt32(&r.chain.pendingNum, 1)
	}
	r.peerID = ""
	r.block = nil
}

// setBlock 当我们从requester.peerID处获得一个区块时，我们就将该区块保留下来，
// 该方法的第二个参数peerID就是用来判断所得的区块是否来自requester.peerID。
func (r *requester) setBlock(block *types.Block, peerID crypto.ID) bool {
	r.mu.Lock()
	if r.block != nil || r.peerID != peerID {
		r.mu.Unlock()
		return false // 我们想要的区块都是从指定的节点处获取的，并非是随便从哪个节点那里获取的
	}
	r.block = block
	r.mu.Unlock()
	select {
	case r.gotCh <- struct{}{}:
	default:
	}
	return true
}

func (r *requester) getBlock() *types.Block {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.block
}

func (r *requester) getPeerID() crypto.ID {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.peerID
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// peer

// onTimeout 如果节点在执行请求任务时超时了，则会执行该函数，该函数向链发送一个节点超时的错误。
func (p *peer) onTimeout() {
	err := errors.New("peer did not send us anything")
	p.chain.sendError(err, p.id)
	p.isTimeout = true
}

// resetTimeout 重置节点执行请求任务的超时时间，超时时间默认是15秒。
func (p *peer) resetTimeout() {
	if p.timeout == nil {
		p.timeout = time.AfterFunc(peerTimeout, p.onTimeout)
	} else {
		p.timeout.Reset(peerTimeout)
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// peerError

func (e peerError) Error() string {
	return fmt.Sprintf("something went wrong on peer %s for %s", e.peerID, e.err)
}
