package p2p

import (
	"encoding/hex"
	"github.com/232425wxy/meta--/common/protoio"
	config2 "github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/proto/pbp2p"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	config := config2.DefaultP2PConfig()
	chDescs := []*ChannelDescriptor{
		{ID: 0x01, Priority: 1, SendQueueCapacity: 1},
		{ID: 0x02, Priority: 1, SendQueueCapacity: 1},
	}
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

func TestConnection_Send(t *testing.T) {
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

	msg := []byte("hai! hai!")
	assert.True(t, clientConn.Send(0x01, msg))
	_, err = server.Read(make([]byte, len(msg)))
	if err != nil {
		t.Error(err)
	}
	assert.True(t, clientConn.CanSend(0x01))
	msg = []byte("小八嘎")
	assert.True(t, clientConn.TrySend(0x01, msg))
	_, err = server.Read(make([]byte, len(msg)))
	if err != nil {
		t.Error(err)
	}
	assert.False(t, clientConn.CanSend(0x02))
	assert.False(t, clientConn.Send(0x03, []byte("ban")))
	assert.False(t, clientConn.TrySend(0x04, []byte("ban")))
}

func TestConnectionReceive(t *testing.T) {
	server, client := netPipe()
	defer func() {
		_ = server.Close()
		_ = client.Close()
	}()
	receivedCh := make(chan []byte)
	errorsCh := make(chan error)
	onReceive := func(chID byte, msg []byte) { receivedCh <- msg }
	onError := func(err error) { errorsCh <- err }
	clientConn := createConn(client)
	clientConn.onReceive = onReceive
	clientConn.onError = onError
	err := clientConn.Start()
	assert.Nil(t, err)
	defer func() {
		_ = clientConn.Stop()
	}()

	serverConn := createConn(server)
	err = serverConn.Start()
	assert.Nil(t, err)
	defer func() {
		_ = serverConn.Stop()
	}()

	msg := []byte("greeting")
	assert.True(t, serverConn.Send(0x01, msg))

	select {
	case received := <-receivedCh:
		assert.Equal(t, received, msg)
	case err = <-errorsCh:
		t.Fatalf("expected %s, but got %v", msg, err)
	case <-time.After(time.Millisecond * 500):
		t.Fatalf("cannot receive %s message in 500ms", msg)
	}
}

func TestConnection_Status(t *testing.T) {
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
	status := clientConn.Status()
	assert.NotNil(t, status)
	assert.Equal(t, status.Channels[0].SendQueueSize, 0)
	assert.Equal(t, status.Channels[0].SendQueueCapacity, 1)
}

func TestPongTimeout(t *testing.T) {
	server, client := netPipe()
	defer func() {
		_ = server.Close()
		_ = client.Close()
	}()
	receivedCh := make(chan []byte)
	errorsCh := make(chan error)
	var onReceive receiveCb = func(chID byte, msg []byte) {
		receivedCh <- msg
	}
	var onError errorCb = func(err error) {
		errorsCh <- err
	}
	clientConn := createConn(client)
	clientConn.onReceive = onReceive
	clientConn.onError = onError
	err := clientConn.Start()
	assert.Nil(t, err)
	defer func() {
		_ = clientConn.Stop()
	}()

	serverGotPing := make(chan struct{})
	go func() {
		var pkt pbp2p.Packet
		_, err = protoio.NewDelimitedReader(server, 1024).ReadMsg(&pkt)
		assert.Nil(t, err)
		serverGotPing <- struct{}{}
	}()
	<-serverGotPing

	pongTimeoutExpired := clientConn.config.PongTimeout + 100*time.Millisecond
	select {
	case msg := <-receivedCh:
		t.Fatalf("expected error, but got %X", msg)
	case err = <-errorsCh:
		assert.NotNil(t, err)
	case <-time.After(pongTimeoutExpired):
		t.Fatalf("expected to receive error before %v timed out", pongTimeoutExpired)
	}
}

func TestConnectionMultiPong(t *testing.T) {
	server, client := netPipe()
	defer func() {
		_ = server.Close()
		_ = client.Close()
	}()
	receivedCh := make(chan []byte)
	errorsCh := make(chan error)
	var onReceive receiveCb = func(chID byte, msg []byte) {
		receivedCh <- msg
	}
	var onError errorCb = func(err error) {
		errorsCh <- err
	}
	clientConn := createConn(client)
	clientConn.onReceive = onReceive
	clientConn.onError = onError
	err := clientConn.Start()
	assert.Nil(t, err)
	defer func() {
		_ = clientConn.Stop()
	}()
	protoWriter := protoio.NewDelimitedWriter(server)
	_, err = protoWriter.WriteMsg(wrapPacket(&pbp2p.PacketPong{}))
	assert.Nil(t, err)
	_, err = protoWriter.WriteMsg(wrapPacket(&pbp2p.PacketPong{}))
	assert.Nil(t, err)
	_, err = protoWriter.WriteMsg(wrapPacket(&pbp2p.PacketPong{}))
	assert.Nil(t, err)
	serverGotPing := make(chan struct{})
	go func() {
		var pkt pbp2p.Packet
		_, err = protoio.NewDelimitedReader(server, 1024).ReadMsg(&pkt)
		assert.Nil(t, err)
		serverGotPing <- struct{}{}
		_, err = protoWriter.WriteMsg(wrapPacket(&pbp2p.PacketPong{}))
		assert.Nil(t, err)
	}()
	<-serverGotPing
	pongTimeoutExpired := clientConn.config.PongTimeout + 100*time.Millisecond
	select {
	case msg := <-receivedCh:
		t.Fatalf("expected no data, but got %X", msg)
	case err = <-errorsCh:
		t.Fatalf("expected no error, but got %v", err)
	case <-time.After(pongTimeoutExpired):
		assert.True(t, clientConn.IsRunning())
	}
}

func TestConnectionMultiPing(t *testing.T) {
	//log.PrintOrigins(true)
	server, client := netPipe()
	defer func() {
		_ = server.Close()
		_ = client.Close()
	}()
	receivedCh := make(chan []byte)
	errorsCh := make(chan error)
	var onReceive receiveCb = func(chID byte, msg []byte) {
		receivedCh <- msg
	}
	var onError errorCb = func(err error) {
		errorsCh <- err
	}
	clientConn := createConn(client)
	clientConn.onReceive = onReceive
	clientConn.onError = onError
	err := clientConn.Start()
	assert.Nil(t, err)
	defer func() {
		_ = clientConn.Stop()
	}()

	protoReader := protoio.NewDelimitedReader(server, 1024)
	protoWriter := protoio.NewDelimitedWriter(server)
	var pkt pbp2p.Packet

	_, err = protoWriter.WriteMsg(wrapPacket(&pbp2p.PacketPing{}))
	assert.Nil(t, err)
	_, err = protoReader.ReadMsg(&pkt)
	assert.Nil(t, err)

	_, err = protoWriter.WriteMsg(wrapPacket(&pbp2p.PacketPing{}))
	assert.Nil(t, err)
	_, err = protoReader.ReadMsg(&pkt)
	assert.Nil(t, err)

	_, err = protoWriter.WriteMsg(wrapPacket(&pbp2p.PacketPing{}))
	assert.Nil(t, err)
	_, err = protoReader.ReadMsg(&pkt)
	assert.Nil(t, err)

	assert.True(t, clientConn.IsRunning())
}

func TestPingPong(t *testing.T) {
	server, client := netPipe()
	defer func() {
		_ = server.Close()
		_ = client.Close()
	}()
	receivedCh := make(chan []byte)
	errorsCh := make(chan error)
	var onReceive receiveCb = func(chID byte, msg []byte) {
		receivedCh <- msg
	}
	var onError errorCb = func(err error) {
		errorsCh <- err
	}
	clientConn := createConn(client)
	clientConn.onReceive = onReceive
	clientConn.onError = onError
	err := clientConn.Start()
	assert.Nil(t, err)
	defer func() {
		_ = clientConn.Stop()
	}()

	serverGotPing := make(chan struct{})
	go func() {
		protoReader := protoio.NewDelimitedReader(server, 1024)
		protoWriter := protoio.NewDelimitedWriter(server)
		var pkt pbp2p.Packet

		_, err = protoReader.ReadMsg(&pkt)
		assert.Nil(t, err)
		serverGotPing <- struct{}{}

		_, err = protoWriter.WriteMsg(wrapPacket(&pbp2p.PacketPong{}))
		assert.Nil(t, err)

		time.Sleep(clientConn.config.PingInterval)
		_, err = protoReader.ReadMsg(&pkt)
		assert.Nil(t, err)
		serverGotPing <- struct{}{}

		_, err = protoWriter.WriteMsg(wrapPacket(&pbp2p.PacketPong{}))
		assert.Nil(t, err)
	}()
	<-serverGotPing
	<-serverGotPing

	pongTimeoutExpired := clientConn.config.PongTimeout + 100*time.Millisecond
	select {
	case msg := <-receivedCh:
		t.Fatalf("expected no data, but got %X", msg)
	case err = <-errorsCh:
		t.Fatalf("expected no error, but got %v", err)
	case <-time.After(pongTimeoutExpired):
		assert.True(t, clientConn.IsRunning())
	}
}

func TestConnectionStopAndReturnError(t *testing.T) {
	log.PrintOrigins(true)
	server, client := netPipe()
	defer func() {
		_ = server.Close()
		_ = client.Close()
	}()
	receivedCh := make(chan []byte)
	errorsCh := make(chan error)
	var onReceive receiveCb = func(chID byte, msg []byte) {
		receivedCh <- msg
	}
	var onError errorCb = func(err error) {
		errorsCh <- err
	}
	clientConn := createConn(client)
	clientConn.onReceive = onReceive
	clientConn.onError = onError
	err := clientConn.Start()
	assert.Nil(t, err)
	defer func() {
		_ = clientConn.Stop()
	}()

	if err = client.Close(); err != nil {
		t.Error(err)
	}

	select {
	case msg := <-receivedCh:
		t.Fatalf("expected no data, but got %x", msg)
	case err = <-errorsCh:
		assert.NotNil(t, err)
		assert.False(t, clientConn.IsRunning())
	case <-time.After(500 * time.Millisecond):
		t.Fatal("didn't receive error in 500ms")
	}
}

func createClientConnAndServerConn(t *testing.T, errorsCh chan interface{}) (*Connection, *Connection) {
	server, client := netPipe()

	clientConn := createConn(client)
	clientConn.Logger = clientConn.Logger.New("module", "client")
	err := clientConn.Start()
	assert.Nil(t, err)

	var onReceive receiveCb = func(chID byte, msg []byte) {}
	var onError errorCb = func(err error) {
		errorsCh <- err
	}
	config := config2.DefaultP2PConfig()
	chDescs := []*ChannelDescriptor{{ID: 0x01, Priority: 1, SendQueueCapacity: 1}}
	serverConn := NewConnectionWithConfig(server, chDescs, onReceive, onError, config)
	logger := log.New()
	logger.SetHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	serverConn.SetLogger(logger)
	serverConn.onError = onError
	serverConn.Logger = serverConn.Logger.New("module", "server")
	err = serverConn.Start()
	assert.Nil(t, err)
	return clientConn, serverConn
}

func expectSend(ch chan interface{}) bool {
	select {
	case <-ch:
		return true
	case <-time.After(5 * time.Second):
		return false
	}
}

func TestConnectionReadErrorEncoding(t *testing.T) {
	log.PrintOrigins(true)
	errorsCh := make(chan interface{})
	client, server := createClientConnAndServerConn(t, errorsCh)

	// 非正常编码数据
	_, err := client.conn.Write([]byte{'a', 'b', 'c'})
	assert.Nil(t, err)
	// 服务端接收数据后，发现没法解码
	assert.True(t, expectSend(errorsCh))

	if err = client.Stop(); err != nil {
		t.Error(err)
	}
	if err = server.Stop(); err != nil {
		t.Error(err)
	}
}

func TestConnectionReadErrorUnknownChannel(t *testing.T) {
	errorsCh := make(chan interface{})
	client, server := createClientConnAndServerConn(t, errorsCh)
	msg := []byte("freedom america, gun shot every day")
	// server收到后，懵逼了
	assert.False(t, client.Send(0x03, msg))
	assert.True(t, client.Send(0x02, msg))
	assert.True(t, expectSend(errorsCh))
	if err := client.Stop(); err != nil {
		t.Error(err)
	}
	if err := server.Stop(); err != nil {
		t.Error(err)
	}
}

func TestConnectionReadErrorLongMessage(t *testing.T) {
	log.PrintOrigins(true)
	errorsCh := make(chan interface{})
	receivedCh := make(chan interface{})

	client, server := createClientConnAndServerConn(t, errorsCh)
	defer func() {
		_ = client.Stop()
		_ = server.Stop()
	}()
	server.onReceive = func(chID byte, msg []byte) {
		receivedCh <- msg
	}

	protoWriter := protoio.NewDelimitedWriter(client.conn)

	var packet = pbp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      make([]byte, client.config.MaxPacketMsgPayloadSize),
	}

	_, err := protoWriter.WriteMsg(wrapPacket(&packet))
	assert.Nil(t, err)
	assert.True(t, expectSend(receivedCh))

	packet = pbp2p.PacketMsg{
		ChannelID: 0x01,
		EOF:       true,
		Data:      make([]byte, client.config.MaxPacketMsgPayloadSize+23),
	}
	_, err = protoWriter.WriteMsg(wrapPacket(&packet))
	assert.Error(t, err)
	assert.True(t, expectSend(errorsCh))
}

func TestConnectionReadErrorUnknownMsgType(t *testing.T) {
	errorsCh := make(chan interface{})
	client, server := createClientConnAndServerConn(t, errorsCh)
	defer func() {
		_ = client.Stop()
		_ = server.Stop()
	}()
	_, err := protoio.NewDelimitedWriter(client.conn).WriteMsg(&pbp2p.NodeInfo{NodeID: "1234567890"})
	assert.Nil(t, err)
	assert.True(t, expectSend(errorsCh))
}

func TestConnVectors(t *testing.T) {

	testCases := []struct {
		testName string
		msg      proto.Message
		expBytes string
	}{
		{"PacketPing", &pbp2p.PacketPing{}, "0a00"},
		{"PacketPong", &pbp2p.PacketPong{}, "1200"},
		{"PacketMsg", &pbp2p.PacketMsg{ChannelID: 1, EOF: false, Data: []byte("data transmitted over the wire")}, "1a2208011a1e64617461207472616e736d6974746564206f766572207468652077697265"},
	}

	for _, tc := range testCases {
		tc := tc

		pm := wrapPacket(tc.msg)
		bz, err := pm.Marshal()
		require.NoError(t, err, tc.testName)

		require.Equal(t, tc.expBytes, hex.EncodeToString(bz), tc.testName)
	}
}

func TestMConnectionTrySend(t *testing.T) {
	log.PrintOrigins(true)
	server, client := netPipe()
	defer func() {
		_ = server.Close()
		_ = client.Close()
	}()

	clientConn := createConn(client)
	err := clientConn.Start()
	require.Nil(t, err)
	defer func() {
		_ = clientConn.Stop()
	}()

	msg := []byte("Semicolon-Woman")
	resultCh := make(chan string, 2)
	assert.True(t, clientConn.TrySend(0x01, msg))
	_, err = server.Read(make([]byte, len(msg)))
	require.NoError(t, err)
	assert.True(t, clientConn.CanSend(0x01))
	assert.True(t, clientConn.TrySend(0x01, msg))
	assert.False(t, clientConn.CanSend(0x01))
	go func() {
		clientConn.TrySend(0x01, msg)
		resultCh <- "TrySend"
	}()
	assert.False(t, clientConn.CanSend(0x01))
	assert.False(t, clientConn.TrySend(0x01, msg))
	assert.Equal(t, "TrySend", <-resultCh)
}
