package p2p

import (
	"fmt"
	"github.com/232425wxy/meta--/common/cmap"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/crypto"
	"math/rand"
	"time"
)

type Switch struct {
	service.BaseService
	reactors     map[string]Reactor
	reactorsByCh map[byte]Reactor
	chDescs      []*ChannelDescriptor
	peers        *PeerSet
	dialing      *cmap.CMap
	reconnecting *cmap.CMap
	addrbook     *AddrBook
	transport    *Transport
	metrics      *Metrics
}

func NewSwitch(transport *Transport, metrics *Metrics) *Switch {
	return &Switch{
		BaseService:  *service.NewBaseService(nil, "Switch"),
		reactors:     make(map[string]Reactor),
		reactorsByCh: make(map[byte]Reactor),
		chDescs:      make([]*ChannelDescriptor, 0),
		peers:        NewPeerSet(),
		dialing:      cmap.NewCap(),
		reconnecting: cmap.NewCap(),
		transport:    transport,
		metrics:      metrics,
	}
}

func (sw *Switch) Start() error {
	for name, reactor := range sw.reactors {
		if name == "STCH" {
			continue
		}
		if err := reactor.Start(); err != nil {
			return err
		}
	}
	time.Sleep(time.Millisecond * 200)
	if sw.Reactor("STCH") != nil {
		if err := sw.Reactor("STCH").Start(); err != nil {
			return err
		}
	}
	go sw.acceptRoutine()
	if sw.addrbook != nil {
		sw.addrbook.Start()
	}
	return sw.BaseService.Start()
}

func (sw *Switch) Stop() error {
	for _, p := range sw.peers.Peers() {
		sw.stopAndRemovePeer(p, nil)
	}
	for _, reactor := range sw.reactors {
		if err := reactor.Stop(); err != nil {
			sw.Logger.Error("failed to stop reactor", "err", err)
		}
	}
	sw.addrbook.Close()
	return sw.BaseService.Stop()
}

func (sw *Switch) NetAddress() *NetAddress {
	return sw.transport.NetAddress()
}

func (sw *Switch) NodeInfo() *NodeInfo {
	return sw.transport.nodeInfo
}

func (sw *Switch) NodeKey() *NodeKey {
	return sw.transport.nodeKey
}

func (sw *Switch) Peers() *PeerSet {
	return sw.peers
}

// AddReactor ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// AddReactor 在交换机这里注册信道和对应的reactor，同时为reactor设置交换机，并将设置过交换机的reactor返回出来。
func (sw *Switch) AddReactor(name string, reactor Reactor) Reactor {
	for _, ch := range reactor.GetChannels() {
		if sw.reactorsByCh[ch.ID] != nil {
			panic(fmt.Sprintf("channel %x has multipile reactors %s", ch.ID, name))
		}
		sw.chDescs = append(sw.chDescs, ch)
		sw.reactorsByCh[ch.ID] = reactor
	}
	sw.reactors[name] = reactor
	reactor.SetSwitch(sw)
	return reactor
}

func (sw *Switch) Reactor(name string) Reactor {
	return sw.reactors[name]
}

// StopPeerForError ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// StopPeerForError 因为某些错误而关闭指定的peer，并将其从本地连接集合中删除。
func (sw *Switch) StopPeerForError(p *Peer, err error) {
	if !p.IsRunning() {
		return
	}
	sw.Logger.Error("stop peer for error", "err", err, "peer", p.NodeID())
	sw.stopAndRemovePeer(p, err)
	panic(err)
}

func (sw *Switch) SetAddrBook(addrbook *AddrBook) {
	sw.addrbook = addrbook
	sw.addrbook.AddOurAddress(sw.NetAddress())
}

func (sw *Switch) DialPeerAsync(peers []string) {
	addrs, errs := NewNetAddressStrings(peers)
	for _, err := range errs {
		sw.Logger.Error("peer's address is not right", "err", err)
	}
	if sw.addrbook != nil {
		for _, addr := range addrs {
			if !addr.Same(sw.NetAddress()) {
				sw.addrbook.AddAddress(addr)
			}
		}
	}
	for _, addr := range addrs {
		go func(addr *NetAddress) {
			if !addr.Same(sw.NetAddress()) {
				// 避免一下子给几十个节点拨号，造成资源不够用
				dur := rand.Intn(2000)
				time.Sleep(time.Duration(dur) * time.Millisecond)
				if err := sw.DialPeerWithAddress(addr); err != nil {
					if _, ok := err.(*ErrorDialingOrAlreadyHas); !ok {
						sw.Logger.Error("failed to dial peer", "err", err)
					}
				}
			}
		}(addr)
	}
}

func (sw *Switch) DialPeerWithAddress(addr *NetAddress) error {
	if sw.addrbook.IsOurAddress(addr) {
		return nil
	}
	if sw.IsDialingOrExisting(addr) {
		return &ErrorDialingOrAlreadyHas{address: addr.DialString()}
	}
	sw.addrbook.AddAddress(addr)
	sw.addrbook.MarkAttempt(addr)
	sw.dialing.Set(string(addr.ID), addr)
	defer sw.dialing.Delete(string(addr.ID))
	p, err := sw.transport.Dial(addr, peerConfig{
		chDescs:      sw.chDescs,
		onPeerError:  sw.StopPeerForError,
		reactorsByCh: sw.reactorsByCh,
		metrics:      sw.metrics,
	})
	if err != nil {
		go sw.reconnectToPeer(addr)
		sw.Logger.Error(fmt.Sprintf("failed to dial peer: %s", addr.DialString()), "err", err)
		return err
	}
	if err = sw.addPeer(p); err != nil {
		sw.transport.Cleanup(p)
		if p.IsRunning() {
			_ = p.Stop()
		}
	}
	sw.Logger.Trace(fmt.Sprintf("successfully dialed the address: %s", addr.DialString()))
	return nil
}

func (sw *Switch) IsDialingOrExisting(addr *NetAddress) bool {
	return sw.dialing.Has(string(addr.ID)) || sw.peers.HasPeer(addr.ID)
}

func (sw *Switch) Broadcast(chID byte, msg []byte) {
	for _, peer := range sw.peers.peers {
		peer := peer
		go func(p *Peer) {
			peer.Send(chID, msg)
		}(peer)
	}
}

func (sw *Switch) SendToPeer(chID byte, peerID crypto.ID, msg []byte) bool {
	for _, peer := range sw.peers.peers {
		if peer.nodeInfo.NodeID == peerID {
			return peer.Send(chID, msg)
		}
	}
	return false
}

func (sw *Switch) reconnectToPeer(addr *NetAddress) {
	if sw.reconnecting.Has(string(addr.ID)) {
		return
	}
	sw.reconnecting.Set(string(addr.ID), addr)
	defer sw.reconnecting.Delete(string(addr.ID))
	for i := 0; i < reconnectAttempts; i++ {
		if !sw.IsRunning() {
			return
		}
		err := sw.DialPeerWithAddress(addr)
		if err != nil {
			sw.Logger.Error("failed reconnect to peer", "err", err, "times", i+1, "address", addr.String())
		} else {
			return
		}
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	}
}

func (sw *Switch) acceptRoutine() {
	for {
		p, err := sw.transport.Accept(peerConfig{
			chDescs:      sw.chDescs,
			onPeerError:  sw.StopPeerForError,
			reactorsByCh: sw.reactorsByCh,
			metrics:      sw.metrics,
		})
		if err != nil {
			sw.Logger.Error("failed to accept new peer", "err", err)
			break
		}
		if err = sw.addPeer(p); err != nil {
			if _, ok := err.(*ErrorAlreadyHasPeer); !ok {
				sw.Logger.Warn("failed to add new peer", "new peer", p.NodeID(), "err", err)
			}
		}
		sw.addrbook.AddAddress(p.NetAddress())
	}
}

// addPeer ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// addPeer 方法所做的不仅仅是将新的peer加入到switch里，还会在各个reactor那里初始化新peer，并将peer启动。
func (sw *Switch) addPeer(p *Peer) error {
	if !sw.filterPeer(p) {
		return &ErrorAlreadyHasPeer{id: p.NodeID()}
	}
	p.SetLogger(sw.Logger.New("module", "Peer", "peer_id", p.NodeID()))
	if !sw.IsRunning() {
		sw.Logger.Error("cannot add new peer, switch is not running", "new peer", p.NodeID())
		return nil
	}
	for _, reactor := range sw.reactors {
		p = reactor.InitPeer(p)
	}
	if err := p.Start(); err != nil {
		return err
	}
	sw.peers.AddPeer(p)
	sw.metrics.Peers.Add(float64(1))
	for _, reactor := range sw.reactors {
		reactor.AddPeer(p)
	}
	sw.Logger.Debug("switch added peer", "new_peer", p.NodeID())
	return nil
}

func (sw *Switch) stopAndRemovePeer(p *Peer, reason error) {
	sw.transport.Cleanup(p)
	if !p.IsRunning() {
		if err := p.Stop(); err != nil {
			sw.Logger.Error("failed to stop peer", "err", err)
		}
	}
	for _, reactor := range sw.reactors {
		reactor.RemovePeer(p, reason)
	}
	if sw.peers.RemovePeer(p) {
		sw.metrics.Peers.Add(float64(-1))
	}
}

// filterPeer ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// filterPeer 过滤掉重复的peer。
func (sw *Switch) filterPeer(p *Peer) bool {
	if sw.peers.HasPeer(p.NodeID()) {
		return false
	}
	return true
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级常量

const (
	reconnectAttempts = 20
)
