package txspool

import (
	"github.com/232425wxy/meta--/common/clist"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/proto/pbtxspool"
	"github.com/232425wxy/meta--/types"
	"time"
)

type Reactor struct {
	p2p.BaseReactor
	cfg  *config.TxsPoolConfig
	pool *TxsPool
}

func NewReactor(cfg *config.TxsPoolConfig, pool *TxsPool) *Reactor {
	return &Reactor{
		BaseReactor: *p2p.NewBaseReactor("TxsPool"),
		cfg:         cfg,
		pool:        pool,
	}
}

func (r *Reactor) InitPeer(peer *p2p.Peer) *p2p.Peer {
	return peer
}

func (r *Reactor) SetLogger(logger log.Logger) {
	r.Logger = logger
}

func (r *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	largestTx := make([]byte, r.cfg.MaxTxBytes)
	msg := pbtxspool.Message{Txs: &pbtxspool.Txs{Txs: [][]byte{largestTx}}}
	return []*p2p.ChannelDescriptor{
		{ID: p2p.TxsChannel, Priority: 10, RecvMessageCapacity: msg.Size()},
	}
}

func (r *Reactor) AddPeer(peer *p2p.Peer) {
	//r.Logger.Debug("add peer", "peer_id", peer.NodeID())
	go r.broadcastTxRoutine(peer)
}

func (r *Reactor) Receive(chID byte, src *p2p.Peer, msg []byte) {
	message := &pbtxspool.Message{}
	err := message.Unmarshal(msg)
	if err != nil {
		r.Logger.Error("receive wrong message tx", "src peer", src, "err", err)
		r.Switch.StopPeerForError(src, err)
		return
	}
	//r.Logger.Debug("receive tx", "src_peer", src, "tx", fmt.Sprintf("%x", msg))
	for _, tx := range message.Txs.Txs {
		err = r.pool.CheckTx(tx, src.NodeID())
		if err != nil {
			if _, ok := err.(*ErrorTxAlreadyExists); !ok {
				r.Logger.Error("check tx failed", "src_peer", src, "err", err)
			}
		}
	}
}

func (r *Reactor) broadcastTxRoutine(peer *p2p.Peer) {
	var element *clist.Element
	for {
		if !r.IsRunning() || !peer.IsRunning() {
			return
		}

		if element == nil {
			select {
			case <-r.pool.WaitTxs():
				if element = r.pool.TxsHead(); element == nil {
					continue
				}
			case <-peer.WaitStop():
				return
			case <-r.WaitStop():
				return
			}
		}

		peerState, ok := peer.Get(types.PeerStateKey).(interface{ GetHeight() int64 })
		if !ok {
			// 共识模块还没有将peer节点的信息存储下来
			time.Sleep(100 * time.Millisecond)
			continue
		}
		tx := element.Value.(*poolTx)
		if peerState.GetHeight() < tx.height-1 {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		if !tx.senders.Has(string(peer.NodeID())) {
			// 如果这个节点没给我发送过该交易数据，那么我就会发送该交易数据给这个节点
			msg := pbtxspool.Message{Txs: &pbtxspool.Txs{Txs: [][]byte{tx.tx}}}
			bz, err := msg.Marshal()
			if err != nil {
				panic(err)
			}
			success := peer.Send(p2p.TxsChannel, bz)
			if !success {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			//r.Logger.Debug("successfully send tx to peer", "peer", peer.NodeID(), "tx", fmt.Sprintf("%x", tx.tx))
		}

		select {
		case <-element.NextWaitChan():
			element = element.Next()
		case <-peer.WaitStop():
			return
		case <-r.WaitStop():
			return
		}
	}
}
