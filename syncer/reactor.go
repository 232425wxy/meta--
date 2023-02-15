package syncer

import (
	"bytes"
	"fmt"
	state2 "github.com/232425wxy/meta--/consensus/state"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/p2p"
	"github.com/232425wxy/meta--/proto/pbsyncer"
	"github.com/232425wxy/meta--/store"
	"github.com/232425wxy/meta--/types"
	"time"
)

type Reactor struct {
	p2p.BaseReactor
	initialState  *state2.State
	blockExecutor *state2.BlockExecutor
	blockStore    *store.BlockStore
	chain         *Blockchain
	requestsCh    chan BlockRequest
	errorsCh      chan peerError
}

func NewReactor(stat *state2.State, blockExecutor *state2.BlockExecutor, blockStore *store.BlockStore, logger log.Logger) *Reactor {
	if stat.LastBlockHeight != blockStore.Height() {
		panic(fmt.Sprintf("state height %d and store height %d mismatch", stat.LastBlockHeight, blockStore.Height()))
	}
	requestsCh := make(chan BlockRequest, maxTotalRequesters)
	errorsCh := make(chan peerError, 20)

	startHeight := blockStore.Height() + 1
	if startHeight == 1 {
		startHeight = stat.InitialHeight
	}
	chain := NewBlockchain(startHeight, requestsCh, errorsCh)
	r := &Reactor{
		BaseReactor:   *p2p.NewBaseReactor("Syncer"),
		initialState:  stat,
		blockExecutor: blockExecutor,
		blockStore:    blockStore,
		chain:         chain,
		requestsCh:    requestsCh,
		errorsCh:      errorsCh,
	}
	r.SetLogger(logger)
	return r
}

func (r *Reactor) SetLogger(logger log.Logger) {
	r.BaseReactor.SetLogger(logger)
	r.chain.SetLogger(logger)
}

func (r *Reactor) Start() error {
	if err := r.chain.Start(); err != nil {
		return err
	}
	go r.processRoutine()
	return nil
}

func (r *Reactor) Stop() error {
	if err := r.chain.Stop(); err != nil {
		return err
	}
	return nil
}

func (r *Reactor) GetChannels() []*p2p.ChannelDescriptor {
	return []*p2p.ChannelDescriptor{
		{
			ID:                  p2p.SyncerChannel,
			Priority:            5,
			SendQueueCapacity:   1000,
			RecvBufferCapacity:  200 * 1024,
			RecvMessageCapacity: 1024 * 1024 * 10,
		},
	}
}

func (r *Reactor) AddPeer(p *p2p.Peer) {
	msgBytes, err := EncodeMsg(&pbsyncer.StatusResponse{Height: r.blockStore.Height()})
	if err != nil {
		panic(err)
	}
	// 即便此处发送失败了也没问题，将来在processRoutine进程中会继续尝试发送
	p.Send(p2p.SyncerChannel, msgBytes)
}

func (r *Reactor) RemovePeer(p *p2p.Peer, reason error) {
	r.chain.removePeer(p.NodeID())
}

func (r *Reactor) Receive(chID byte, src *p2p.Peer, msgBytes []byte) {
	msg, err := DecodeMsg(msgBytes)
	if err != nil {
		r.Logger.Error("failed to decode message", "sender_id", src.NodeID(), "err", err)
		r.Switch.StopPeerForError(src, err)
		return
	}

	switch msg := msg.(type) {
	case *pbsyncer.BlockRequest:
		r.respondToPeer(msg, src)
	case *pbsyncer.BlockResponse:
		block := types.BlockFromProto(msg.Block)
		r.chain.AddBlock(src.NodeID(), block)
	case *pbsyncer.NoBlockResponse:
		r.Logger.Warn("peer does not have expected block", "peer_id", src.NodeID(), "height", msg.Height)
	case *pbsyncer.StatusRequest:
		bz, err := EncodeMsg(&pbsyncer.StatusResponse{Height: r.blockStore.Height()})
		if err != nil {
			r.Logger.Error("failed to encode StatusResponse message", "err", err)
			return
		}
		src.TrySend(p2p.SyncerChannel, bz)
	case *pbsyncer.StatusResponse:
		r.chain.SetPeerUpHeight(src.NodeID(), msg.Height)
	default:
		r.Logger.Warn(fmt.Sprintf("unknown message type: %T", msg))
	}
}

func (r *Reactor) BroadcastStatusRequest() {
	bz, err := EncodeMsg(&pbsyncer.StatusRequest{})
	if err != nil {
		r.Logger.Error("failed to encode StatusRequest message", "err", err)
		return
	}
	r.Switch.Broadcast(p2p.SyncerChannel, bz)
}

// respondToPeer 方法对向我提出索要区块请求的节点给出回应，返回值是一个bool类型，无论我有没有对方想要的区块，只要我成功
// 回应了对方（即使回复说我没有你要的区块），就返回true，否则返回false。
func (r *Reactor) respondToPeer(req *pbsyncer.BlockRequest, src *p2p.Peer) bool {
	block := r.blockStore.LoadBlockByHeight(req.Height)
	if block != nil {
		pbBlock := block.ToProto()
		msgBytes, err := EncodeMsg(&pbsyncer.BlockResponse{Block: pbBlock})
		if err != nil {
			r.Logger.Error("could not marshal BlockResponse message", "err", err)
			return false
		}
		return src.TrySend(p2p.SyncerChannel, msgBytes)
	}
	r.Logger.Warn("peer asked for a block that we don't have", "peer_id", src.NodeID(), "height", req.Height)
	msgBytes, err := EncodeMsg(&pbsyncer.NoBlockResponse{Height: req.Height})
	if err != nil {
		r.Logger.Error("could not marshal NoBlockResponse message", "err", err)
		return false
	}
	return src.TrySend(p2p.SyncerChannel, msgBytes)
}

func (r *Reactor) processRoutine() {
	syncTicker := time.NewTicker(10 * time.Millisecond)
	defer syncTicker.Stop()

	statusUpdateTicker := time.NewTicker(10 * time.Second)
	defer statusUpdateTicker.Stop()

	switchToConsensusTicker := time.NewTicker(time.Second)
	defer switchToConsensusTicker.Stop()

	stat := r.initialState

	didProcessCh := make(chan struct{}, 1)

	go func() {
		for {
			select {
			case <-r.WaitStop():
				return
			case <-r.chain.WaitStop():
				return
			case request := <-r.requestsCh:
				p := r.Switch.Peers().GetPeer(request.PeerID)
				if p == nil { // 这样的情况一般来说是不会发生的
					continue
				}
				msgBytes, err := EncodeMsg(&pbsyncer.BlockRequest{Height: request.Height})
				if err != nil {
					r.Logger.Error("failed to encode BlockRequest message", "err", err)
					continue
				}
				if ok := p.TrySend(p2p.SyncerChannel, msgBytes); !ok {
					r.Logger.Warn("send request failed, maybe send queue is full, so, request is dropped", "receiver_id", p.NodeID())
				}
			case err := <-r.errorsCh:
				p := r.Switch.Peers().GetPeer(err.peerID)
				if p != nil {
					r.Switch.StopPeerForError(p, err)
				}
			case <-statusUpdateTicker.C: // 定时器提醒我们需要
				go r.BroadcastStatusRequest()
			}
		}
	}()

LOOP:
	for {
		select {
		case <-switchToConsensusTicker.C:
			height, _, _ := r.chain.GetStatus()
			if r.chain.IsCaughtUp() {
				r.Logger.Info("time to switch to consensus reactor", "height", height)
				if err := r.chain.Stop(); err != nil {
					r.Logger.Error("failed to stop syncer chain", "err", err)
				}
				if conReactor, ok := r.Switch.Reactor("CONSENSUS").(consensusReactor); ok {
					conReactor.SwitchToConsensus(stat)
				}
				break LOOP
			}
		case <-syncTicker.C:
			select {
			case didProcessCh <- struct{}{}:
			default:
			}
		case <-didProcessCh:
			first, second := r.chain.PickTwoBlocks()
			if first == nil || second == nil {
				continue LOOP
			} else {
				didProcessCh <- struct{}{}
			}
			if !bytes.Equal(second.Header.PreviousBlockHash, first.ChameleonHash.Hash) {
				peerID1 := r.chain.RedoRequest(first.Header.Height)
				if p := r.Switch.Peers().GetPeer(peerID1); p != nil {
					r.Switch.StopPeerForError(p, fmt.Errorf("provide invalid blocks"))
				}
				peerID2 := r.chain.RedoRequest(second.Header.Height)
				if p := r.Switch.Peers().GetPeer(peerID2); p != nil {
					r.Switch.StopPeerForError(p, fmt.Errorf("provide invalid block"))
				}
				continue LOOP
			} else {
				r.chain.PopRequest()
				r.blockStore.SaveBlock(first)
				var err error
				stat, err = r.blockExecutor.ApplyBlock(stat, first)
				if err != nil {
					panic("failed to apply committed block")
				}
			}
			continue LOOP
		case <-r.WaitStop():
			break LOOP
		}
	}
}

type consensusReactor interface {
	SwitchToConsensus(stat *state2.State)
}

// for test

func (r *Reactor) BlockStore() *store.BlockStore {
	return r.blockStore
}
