package p2p

import (
	"fmt"
	"github.com/232425wxy/meta--/common/cmap"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/log"
	"net"
	"sync"
	"time"
)

type peerConn struct {
	conn net.Conn // 这里的conn是最原始的net.Conn，在将来会将其包装成 Connection
	addr *NetAddress
	ip   net.IP // 该字段在RemoteIP方法中被赋值。
}

func newPeerConn(conn net.Conn, addr *NetAddress) peerConn {
	return peerConn{
		conn: conn,
		addr: addr,
	}
}

func (pc peerConn) RemoteIP() net.IP {
	if pc.ip != nil {
		return pc.ip
	}

	host, _, err := net.SplitHostPort(pc.conn.RemoteAddr().String())
	if err != nil {
		panic(err)
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		panic(err)
	}

	pc.ip = ips[0]

	return pc.ip
}

type Peer struct {
	service.BaseService
	peerConn
	connection    *Connection
	nodeInfo      *NodeInfo
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

func newPeer(pc peerConn, cfg *config.P2PConfig, nodeInfo *NodeInfo, reactorsByCh map[byte]Reactor, chDescs []*ChannelDescriptor, onPeerError func(peer *Peer, err error), metrics *Metrics) *Peer {
	p := &Peer{
		BaseService:   *service.NewBaseService(nil, "Peer"),
		peerConn:      pc,
		nodeInfo:      nodeInfo,
		Data:          cmap.NewCap(),
		metrics:       metrics,
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
	p.connection = NewConnectionWithConfig(pc.conn, chDescs, onReceive, onError, cfg)
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

func (p *Peer) NodeInfo() *NodeInfo {
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

// peer 集合

type PeerSet struct {
	mu      sync.RWMutex
	indexes map[crypto.ID]*peerIndex
	peers   []*Peer
}

type peerIndex struct {
	peer  *Peer
	index int
}

func NewPeerSet() *PeerSet {
	return &PeerSet{indexes: make(map[crypto.ID]*peerIndex), peers: make([]*Peer, 0)}
}

// AddPeer ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// AddPeer 往peer节点集合中加入新的节点，如果该节点已经存在了，就什么也不做，直接返回。
func (ps *PeerSet) AddPeer(peer *Peer) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if ps.indexes[peer.NodeID()] != nil {
		return
	}
	index := len(ps.peers)
	ps.peers = append(ps.peers, peer)
	ps.indexes[peer.NodeID()] = &peerIndex{peer: peer, index: index}
}

// HasPeer ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// HasPeer 给定某个peer节点的id，判断该peer节点存不存在，如果不存在就返回false。
func (ps *PeerSet) HasPeer(peerID crypto.ID) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.indexes[peerID] != nil
}

// HasIP ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// HasIP 判断在peer节点集合中是否已经存在给定的IP地址，这在给节点拨号时很有用，避免重复拨号。
func (ps *PeerSet) HasIP(ip net.IP) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	for _, peer := range ps.peers {
		if peer.RemoteIP().Equal(ip) {
			return true
		}
	}
	return false
}

// GetPeer ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// GetPeer 根据peer节点的ID获取对应的peer节点。
func (ps *PeerSet) GetPeer(peerID crypto.ID) *Peer {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	if _, ok := ps.indexes[peerID]; ok {
		return ps.indexes[peerID].peer
	}
	return nil
}

// RemovePeer ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// RemovePeer 从peer集合中删除指定的peer节点。
func (ps *PeerSet) RemovePeer(peer *Peer) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	indexIt := ps.indexes[peer.NodeID()]
	if indexIt == nil {
		return false
	}
	index := indexIt.index
	peers := make([]*Peer, len(ps.peers)-1)
	copy(peers, ps.peers)
	delete(ps.indexes, peer.NodeID())
	// 如果要删除的节点是最后一个节点，那么直接截取前n-1个节点就行了
	if index == len(ps.peers)-1 {
		ps.peers = peers
		return true
	}
	// 如果要删除的节点是中间某个节点，就将最后那个节点和中间这个节点调换一下，然后改一下索引位置就行了。
	lastPeer := ps.peers[len(ps.peers)-1]
	peers[index] = lastPeer
	ps.indexes[lastPeer.NodeID()].index = index
	ps.peers = peers
	return true
}

// Size ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Size 返回peer集合中peer节点数量。
func (ps *PeerSet) Size() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return len(ps.peers)
}

func (ps *PeerSet) Peers() []*Peer {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return ps.peers
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级常量

const (
	metricsTickerDuration = 10 * time.Second
)
