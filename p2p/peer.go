package p2p

import (
	"fmt"
	"github.com/232425wxy/meta--/common/cmap"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/log"
	"net"
	"time"
)

type peerConn struct {
	conn net.Conn // 这里的conn是最原始的net.Conn，在将来会将其包装成 Connection
	addr *NetAddress
	ip   net.IP
}

func newPeerConn(conn net.Conn, addr *NetAddress) peerConn {
	return peerConn{
		conn: conn,
		addr: addr,
	}
}

type Peer struct {
	service.BaseService
	peerConn
	connection    *Connection
	nodeInfo      NodeInfo
	Data          *cmap.CMap
	metrics       *Metrics
	metricsTicker *time.Ticker
}

type PeerOption func(peer *Peer)

func PeerOptionSetMetrics(m *Metrics) PeerOption {
	return func(peer *Peer) {
		peer.metrics = m
	}
}

func newPeer(pc peerConn, config ConnectionConfig, nodeInfo NodeInfo, reactorsByCh map[byte]Reactor, chDescs []*ChannelDescriptor, onPeerError func(peer *Peer, err error), options ...PeerOption) *Peer {
	p := &Peer{
		BaseService:   *service.NewBaseService(nil, "Peer"),
		peerConn:      pc,
		nodeInfo:      nodeInfo,
		Data:          cmap.NewCap(),
		metrics:       P2PMetrics(),
		metricsTicker: time.NewTicker(metricsTickerDuration),
	}
	// 在这里给每个模块的reactor注册各自信道收到消息后的处理方法。
	var onReceive receiveCb = func(chID byte, msg []byte) {
		reactor := reactorsByCh[chID]
		if reactor == nil {
			panic(fmt.Sprintf("unknown channel id %x", chID))
		}
		labels := []string{"peer_id", string(nodeInfo.NodeID), "channel_id", fmt.Sprintf("%x", chID)}
		p.metrics.PeerReceiveBytesTotal.With(labels...).Add(float64(len(msg)))
		reactor.Receive(chID, p, msg)
	}
	var onError errorCb = func(err error) {
		onPeerError(p, err)
	}
	p.connection = NewConnectionWithConfig(pc.conn, chDescs, onReceive, onError, config)
	// 这个地方一般是把Switch那里的metrics下放到Peer这里
	for _, opt := range options {
		opt(p)
	}
	return p
}

func (p *Peer) NodeID() crypto.ID {
	return p.nodeInfo.ID()
}

func (p *Peer) String() string {
	return fmt.Sprintf("Peer{%s}", p.NodeID())
}

func (p *Peer) SetLogger(l log.Logger) {
	p.Logger = l
	p.connection.SetLogger(l)
}

func (p *Peer) Start() error {
	if err := p.connection.Start(); err != nil {
		return err
	}
	if err := p.BaseService.Start(); err != nil {
		return err
	}
	go p.metricsReport()
	return nil
}

func (p *Peer) Stop() error {
	p.metricsTicker.Stop()
	p.connection.FlushStop()
	return p.BaseService.Stop()
}

func (p *Peer) NodeInfo() NodeInfo {
	return p.nodeInfo
}

func (p *Peer) NetAddress() *NetAddress {
	return p.addr
}

// Status ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Status 返回p2p/connection的状态。
func (p *Peer) Status() ConnectionStatus {
	return p.connection.Status()
}

// Send ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Send 给对方的指定信道发送消息。
func (p *Peer) Send(chID byte, msg []byte) bool {
	if !p.IsRunning() {
		return false
	}
	res := p.connection.Send(chID, msg)
	if res {
		labels := []string{"peer_id", string(p.NodeID()), "channel_id", fmt.Sprintf("%x", chID)}
		p.metrics.PeerSendBytesTotal.With(labels...).Add(float64(len(msg)))
	}
	return res
}

// TrySend ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// TrySend 尝试发送数据。
func (p *Peer) TrySend(chID byte, msg []byte) bool {
	if !p.IsRunning() {
		return false
	}
	res := p.connection.TrySend(chID, msg)
	if res {
		labels := []string{"peer_id", string(p.NodeID()), "channel_id", fmt.Sprintf("%x", chID)}
		p.metrics.PeerSendBytesTotal.With(labels...).Add(float64(len(msg)))
	}
	return res
}

func (p *Peer) Get(key string) interface{} {
	return p.Data.Get(key)
}

func (p *Peer) Set(key string, data interface{}) {
	p.Data.Set(key, data)
}

// RemoteAddr ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// RemoteAddr 返回对方节点的网络地址。
func (p *Peer) RemoteAddr() net.Addr {
	return p.conn.RemoteAddr()
}

func (p *Peer) metricsReport() {
	for {
		select {
		case <-p.metricsTicker.C:
			status := p.connection.Status()
			var sendQueueSize float64
			for _, chStats := range status.Channels {
				sendQueueSize += float64(chStats.SendQueueSize)
			}
			p.metrics.PeerPendingSendBytes.With("peer_id", string(p.NodeID())).Set(sendQueueSize)
		case <-p.WaitStop():
			return
		}
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级常量

const (
	metricsTickerDuration = 10 * time.Second
)
