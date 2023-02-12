package rpc

import (
	"github.com/go-kit/kit/transport/http"
	"net"
	"strings"
	"sync"
)

const (
	ProtocolHTTP
	ProtocolTCP
)

type dialer func(string, string) (net.Conn, error)

type Client struct {
	address  string
	username string
	password string
	client   *http.Client
	mu       sync.Mutex
}

func NewClientWithHTTPClient(remote string, client *http.Client) (*Client, error) {

}

func createHTTPDialer(remote string) (protocol string, address string, d dialer) {
	parts := strings.SplitN(remote, "://", 2)
	if len(parts) == 1 {
		protocol, address = ProtocolTCP, remote
	} else if len(parts) == 2 {
		protocol, address = parts[0], parts[1]
	}
}