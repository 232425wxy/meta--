package p2p

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/232425wxy/meta--/common/protoio"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/proto/pbp2p"
)

type Transport struct {
	addr             *NetAddress
	listener         net.Listener
	acceptc          chan accept
	closed           chan struct{}
	connSet          *ConnSet
	dialTimeout      time.Duration
	handshakeTimeout time.Duration
	nodeInfo         *NodeInfo
	nodeKey          *NodeKey
	config           *config.P2PConfig
}

type accept struct {
	addr     *NetAddress
	conn     net.Conn
	nodeInfo *NodeInfo
	err      error
}

// peerConfig ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// peerConfig 用来存储对方节点所掌握的信道信息、处理消息的反应器等消息。
type peerConfig struct {
	chDescs      []*ChannelDescriptor
	onPeerError  func(peer *Peer, err error)
	reactorsByCh map[byte]Reactor
	metrics      *Metrics
}

func NewTransport(addr *NetAddress, nodeInfo *NodeInfo, nodeKey *NodeKey, config *config.P2PConfig) *Transport {
	return &Transport{
		addr:             addr,
		acceptc:          make(chan accept),
		closed:           make(chan struct{}),
		connSet:          NewConnSet(),
		dialTimeout:      defaultDialTimeout,
		handshakeTimeout: defaultHandshakeTimeout,
		nodeInfo:         nodeInfo,
		nodeKey:          nodeKey,
		config:           config,
	}
}

// Listen ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Listen 监听网络中新的连接。
func (t *Transport) Listen() error {
	ln, err := net.Listen("tcp", t.addr.DialString())
	if err != nil {
		return err
	}
	t.listener = ln
	go t.acceptPeers()
	return nil
}

// Close ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Close 关闭网络连接监听器。
func (t *Transport) Close() error {
	close(t.closed)
	return t.listener.Close()
}

// Cleanup ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Cleanup 将与指定的peer相关的底层网络连接从集合中删除，同时关闭该底层连接。
func (t *Transport) Cleanup(p *Peer) {
	t.connSet.RemoveAddr(p.RemoteAddr())
	_ = p.conn.Close()
}

func (t *Transport) NetAddress() *NetAddress {
	return t.addr
}

func (t *Transport) Accept(config peerConfig) (*Peer, error) {
	select {
	case a := <-t.acceptc:
		if a.err != nil {
			return nil, a.err
		}
		return t.wrapPeer(a.conn, a.nodeInfo, config, a.addr), nil
	case <-t.closed:
		return nil, errors.New("p2p/transport is closed")
	}
}

// Dial ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Dial 给指定的地址拨号，与一个新的peer取得联系。
func (t *Transport) Dial(addr *NetAddress, config peerConfig) (*Peer, error) {
	c, err := addr.DialTimeout(t.dialTimeout)
	if err != nil {
		return nil, err
	}
	if err = t.filterConn(c); err != nil {
		return nil, err
	}
	peerInfo, err := handshake(c, t.handshakeTimeout, t.nodeInfo)
	if err != nil {
		return nil, err
	}
	peer := t.wrapPeer(c, peerInfo, config, addr)
	return peer, nil
}

// AddChannel ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// AddChannel 在peer节点处注册信道。
func (t *Transport) AddChannel(chID byte) {
	if !t.nodeInfo.HasChannel(chID) {
		t.nodeInfo.Channels = append(t.nodeInfo.Channels, chID)
	}
}

// acceptPeers ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// acceptPeers 时刻监听网络中的新连接，将其包装成peer加入到peer集合中。
func (t *Transport) acceptPeers() {
	for {
		c, err := t.listener.Accept()
		if err != nil {
			select {
			case _, ok := <-t.closed:
				if !ok {
					return
				}
			default:
			}
			t.acceptc <- accept{err: err}
			return
		}
		go func(c net.Conn) {
			var addr *NetAddress
			var peerInfo *NodeInfo
			// 这与之前过滤的连接不同，这次的连接是别人主动给我们拨号的
			err = t.filterConn(c)
			if err == nil {
				peerInfo, err = handshake(c, t.handshakeTimeout, t.nodeInfo)
				if err == nil {
					addr = NewNetAddress(peerInfo.ID(), c.RemoteAddr())
				}
			}
			select {
			case t.acceptc <- accept{addr: addr, conn: c, nodeInfo: peerInfo, err: err}:
			case <-t.closed:
				_ = c.Close()
				return
			}
		}(c)
	}
}

// wrapPeer ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// wrapPeer 方法在包装peer节点时，除了会用到入参里的peerConfig配置信息，还会用到Transport自己所携带的ConnectionConfig，
// 用它去创建底层的p2p连接Connection。
func (t *Transport) wrapPeer(c net.Conn, nodeInfo *NodeInfo, config peerConfig, addr *NetAddress) *Peer {
	pc := newPeerConn(c, addr)
	peer := newPeer(pc, t.config, nodeInfo, config.reactorsByCh, config.chDescs, config.onPeerError, config.metrics)
	return peer
}

// filterConn ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// filterConn 过滤掉重复的连接，如果给定的连接并不重复，则将其加入到连接集合中。
func (t *Transport) filterConn(c net.Conn) error {
	if t.connSet.HasConn(c) {
		_ = c.Close()
		return fmt.Errorf("has already connect to %q", c.RemoteAddr().String())
	}
	host, _, err := net.SplitHostPort(c.RemoteAddr().String())
	if err != nil {
		return err
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return err
	}
	t.connSet.Add(c, ips)
	return nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// handshake ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// handshake 和对方节点握手并交换节点信息。
func handshake(c net.Conn, handshakeTimout time.Duration, nodeInfo *NodeInfo) (*NodeInfo, error) {
	if err := c.SetDeadline(time.Now().Add(handshakeTimout)); err != nil {
		return nil, err
	}
	errc := make(chan error, 2)
	pbInfo := &pbp2p.NodeInfo{}
	go func(errc chan<- error, c net.Conn) {
		_, err := protoio.NewDelimitedWriter(c).WriteMsg(nodeInfo.ToProto())
		errc <- err
	}(errc, c)
	go func(errc chan<- error, c net.Conn) {
		_, err := protoio.NewDelimitedReader(c, maxNodeInfoSize).ReadMsg(pbInfo)
		errc <- err
	}(errc, c)
	for i := 0; i < cap(errc); i++ {
		if err := <-errc; err != nil {
			return nil, err
		}
	}
	peerInfo := NodeInfoFromProto(pbInfo)
	if err := bls12.AddBLSPublicKey(peerInfo.PublicKey); err != nil {
		return nil, err
	}
	// 设置空的超时时间，意味着该连接不会超时。
	return peerInfo, c.SetDeadline(time.Time{})
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级常量

const (
	defaultDialTimeout      = time.Second
	defaultHandshakeTimeout = 3 * time.Second
)
