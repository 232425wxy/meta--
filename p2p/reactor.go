package p2p

import "github.com/232425wxy/meta--/common/service"

const (
	LeaderProposeChannel   byte = 0x01
	ReplicaVoteChannel     byte = 0x02
	ReplicaNextViewChannel byte = 0x03 // 副本节点给主节点发送NextView消息时用到的通道
	ReplicaStateChannel    byte = 0x04
	TxsChannel             byte = 0x05
	SyncerChannel          byte = 0x06
	STCHChannel            byte = 0x07
)

type Reactor interface {
	service.Service
	SetSwitch(*Switch)
	GetChannels() []*ChannelDescriptor
	InitPeer(peer *Peer) *Peer
	AddPeer(peer *Peer)
	RemovePeer(peer *Peer, reason error)
	Receive(chID byte, peer *Peer, msg []byte)
}

type BaseReactor struct {
	service.BaseService
	Switch *Switch
}

func NewBaseReactor(name string) *BaseReactor {
	return &BaseReactor{
		BaseService: *service.NewBaseService(nil, name),
		Switch:      nil,
	}
}

func (br *BaseReactor) SetSwitch(s *Switch) {
	br.Switch = s
}

func (br *BaseReactor) InitPeer(peer *Peer) *Peer { return peer }

func (br *BaseReactor) GetChannels() []*ChannelDescriptor { return nil }

func (br *BaseReactor) AddPeer(peer *Peer) {}

func (br *BaseReactor) RemovePeer(peer *Peer, reason error) {}

func (br *BaseReactor) Receive(chID byte, peer *Peer, msg []byte) {}
