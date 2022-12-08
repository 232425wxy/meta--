package p2p

import (
	"github.com/232425wxy/meta--/log"
	"github.com/stretchr/testify/assert"
	"net"
	"os"
	"testing"
	"time"
)

func netPipe() (net.Conn, net.Conn) {
	return net.Pipe()
}

func createConn(conn net.Conn) *Connection {
	var onReceive receiveCb = func(chID byte, msg []byte) {

	}
	var onError errorCb = func(err error) {

	}
	logger := log.New()
	logger.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	//log.PrintOrigins(true)
	config := ConnectionConfig{
		SendRate:                5120000,
		RecvRate:                5120000,
		MaxPacketMsgPayloadSize: 1024,
		FlushDur:                50 * time.Millisecond,
		PingInterval:            90 * time.Millisecond,
		PongTimeout:             45 * time.Millisecond,
	}
	chDescs := []*ChannelDescriptor{{ID: 0x01, Priority: 1, SendQueueCapacity: 1}}
	c := NewConnectionWithConfig(conn, chDescs, onReceive, onError, config)
	c.SetLogger(logger)
	return c
}

func TestConnectionSendFlushStop(t *testing.T) {
	server, client := netPipe()
	defer func() {
		_ = server.Close()
		_ = client.Close()
	}()

	clientConn := createConn(client)
	err := clientConn.Start()
	assert.Nil(t, err)
	defer func() {
		_ = clientConn.Stop()
	}()

	msg := []byte("abc")
	assert.True(t, clientConn.Send(0x01, msg))
	msgLength := 14
	errCh := make(chan error)
	go func() {
		msgB := make([]byte, msgLength)
		_, err = server.Read(msgB)
		if err != nil {
			t.Error(err)
			return
		}
		errCh <- err
	}()
	clientConn.FlushStop()
	timer := time.NewTimer(3 * time.Second)
	select {
	case <-errCh:
	case <-timer.C:
		t.Error("timed out waiting for msg")
	}
}
