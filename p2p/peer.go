package p2p

import (
	"github.com/232425wxy/meta--/common/cmap"
	"github.com/232425wxy/meta--/common/service"
	"net"
	"time"
)

type peerConn struct {
	conn net.Conn
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
	channels      []byte
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
