package p2p

import (
	"bufio"
	"crypto/cipher"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/common/async"
	"github.com/232425wxy/meta--/common/flowrate"
	"github.com/232425wxy/meta--/common/flusher"
	"github.com/232425wxy/meta--/common/number"
	"github.com/232425wxy/meta--/common/protoio"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/proto/pbp2p"
	"github.com/cosmos/gogoproto/proto"
	"golang.org/x/crypto/chacha20poly1305"
	"io"
	"math"
	"net"
	"reflect"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

// 这个文件里定义了普通通信和加密通信

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// Connection ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Connection p2p通信里的底层连接。
type Connection struct {
	service.BaseService
	conn              net.Conn // 还是要靠net.Conn来和对方取得联系的
	bufConnReader     *bufio.Reader
	bufConnWriter     *bufio.Writer
	sendMonitor       *flowrate.Monitor
	recvMonitor       *flowrate.Monitor
	sendChan          chan struct{}
	pong              chan struct{}
	channels          []*Channel
	channelsIdx       map[byte]*Channel
	onReceive         receiveCb // onReceive可以接收到的消息通过Reactor递送到指定的模块去处理
	onError           errorCb
	errored           uint32
	config            ConnectionConfig
	quitSendRoutine   chan struct{}
	doneSendRoutine   chan struct{}
	quitRecvRoutine   chan struct{}
	stopMu            sync.Mutex
	flusher           *flusher.Flusher
	pingTicker        *time.Ticker
	pongTimer         *time.Timer
	pongTimeoutCh     chan bool
	chStatsTicker     *time.Ticker // 每隔一段时间更新一下信道状态
	created           time.Time
	_maxPacketMsgSize int // 数据包的最大大小
}

// ConnectionConfig ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ConnectionConfig p2p连接的配置信息。
type ConnectionConfig struct {
	SendRate                int64         `mapstructure:"send_rate"`
	RecvRate                int64         `mapstructure:"recv_rate"`
	MaxPacketMsgPayloadSize int           `mapstructure:"max_packet_msg_payload_size"`
	FlushDur                time.Duration `mapstructure:"flusher"`
	PingInterval            time.Duration `mapstructure:"ping_interval"`
	PongTimeout             time.Duration `mapstructure:"pong_timeout"`
}

// receiveCb ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// receiveCb 收到消息后要干什么呢，这个就交给对应的信道去处理吧。
type receiveCb func(chID byte, msg []byte)

// errorCb ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// errorCb 发生错误了该怎么办呢？这也交给对应的回调函数去处理吧。
type errorCb func(err error)

// NewConnectionWithConfig ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewConnectionWithConfig 根据配置信息实例化一个p2p网络中的底层连接。
func NewConnectionWithConfig(conn net.Conn, chDescs []*ChannelDescriptor, onReceive receiveCb, onError errorCb, config ConnectionConfig) *Connection {
	connection := &Connection{
		conn:          conn,
		bufConnReader: bufio.NewReaderSize(conn, minReadBufferSize),
		bufConnWriter: bufio.NewWriterSize(conn, minWriteBufferSize),
		sendMonitor:   flowrate.NewMonitor(0, config.SendRate),
		recvMonitor:   flowrate.NewMonitor(0, config.RecvRate),
		sendChan:      make(chan struct{}, 1),
		pong:          make(chan struct{}, 1),
		onReceive:     onReceive,
		onError:       onError,
		config:        config,
		created:       time.Now(),
	}
	connection.BaseService = *service.NewBaseService(nil, "P2P/Connection")
	if config.PongTimeout >= config.PingInterval {
		connection.Logger.Error("pongTimeout must be less than pingInterval, we has adjust pongTimeout to pingInterval/2")
		config.PongTimeout = config.PingInterval / 2
	}
	var channelsIdx = make(map[byte]*Channel)
	var channels = make([]*Channel, 0)
	for _, desc := range chDescs {
		desc.FillDefaults()
		channel := &Channel{
			conn:                    connection,
			desc:                    *desc,
			sendQueue:               make(chan []byte, desc.SendQueueCapacity),
			recving:                 make([]byte, 0, desc.RecvBufferCapacity),
			maxPacketMsgPayloadSize: config.MaxPacketMsgPayloadSize,
			Logger:                  connection.Logger,
		}
		channelsIdx[desc.ID] = channel
		channels = append(channels, channel)
	}
	connection.channels = channels
	connection.channelsIdx = channelsIdx
	connection._maxPacketMsgSize = connection.maxPacketMsgSize()
	return connection
}

// SetLogger ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// SetLogger 为每个信道都设置日志记录器。
func (c *Connection) SetLogger(logger log.Logger) {
	child := logger.New("conn", c.String())
	c.BaseService.SetLogger(child)
	for _, ch := range c.channels {
		channelLogger := child.New("channel id", ch.desc.ID)
		ch.Logger = channelLogger
	}
}

func (c *Connection) Start() error {
	c.flusher = flusher.NewFlusher(c.config.FlushDur)
	c.pingTicker = time.NewTicker(c.config.PingInterval)
	c.pongTimeoutCh = make(chan bool, 1)
	c.chStatsTicker = time.NewTicker(updateStats)
	c.quitSendRoutine = make(chan struct{})
	c.doneSendRoutine = make(chan struct{})
	c.quitRecvRoutine = make(chan struct{})
	go c.sendRoutine()
	go c.recvRoutine()
	return c.BaseService.Start()
}

// StopServices ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// stopServices 关闭所有计时器，发送退出sendRoutine和recvRoutine协程的信号。
func (c *Connection) stopServices() (stopped bool) {
	c.stopMu.Lock()
	defer c.stopMu.Unlock()
	select {
	case <-c.quitSendRoutine:
		return true
	case <-c.quitRecvRoutine:
		return true
	default:

	}

	c.flusher.Stop()
	c.pingTicker.Stop()
	c.chStatsTicker.Stop()

	close(c.quitSendRoutine)
	close(c.quitRecvRoutine)
	return false
}

// FlushStop ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FlushStop 在关闭之间将信道里的消息都发送出去，然后关闭底层网络连接net.Conn。
func (c *Connection) FlushStop() {
	if c.stopServices() {
		return
	}
	// 等待sendRoutine协程将数据发送完毕，这样就不会竞争调用sendSomePacketMasgs方法。
	<-c.doneSendRoutine
	eof := c.sendSomePacketMsgs()
	for !eof {
		eof = c.sendSomePacketMsgs()
	}
	c.flush()
	_ = c.conn.Close()
}

func (c *Connection) Stop() error {
	if c.stopServices() {
		return nil
	}
	_ = c.conn.Close()
	return c.BaseService.Stop()
}

func (c *Connection) String() string {
	return fmt.Sprintf("P2P/Connection:%v", c.conn.RemoteAddr())
}

// flush ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// flush 将数据写入底层的网络连接net.Conn里。
func (c *Connection) flush() {
	c.Logger.Debug("flush message to the other side")
	err := c.bufConnWriter.Flush()
	if err != nil {
		c.Logger.Error("failed to flush message to the other side", "err", err)
	}
}

// recover ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// recover 检测网络连接是否遇到严重错误。
func (c *Connection) recover() {
	if r := recover(); r != nil {
		c.Logger.Error("connection panicked", "err", r, "stack", string(debug.Stack()))
		c.stopForError(fmt.Errorf("recovered from panic: %v", r))
	}
}

// stopForError ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// stopForError 因为某种严重错误才会导致网络连接停止。
func (c *Connection) stopForError(err error) {
	if e := c.Stop(); e != nil {
		c.Logger.Error("failed to stop connection", "err", e)
	}
	if atomic.CompareAndSwapUint32(&c.errored, 0, 1) {
		if c.onError != nil {
			c.onError(err)
		}
	}
}

// Send ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Send 向指定信道发送数据。
func (c *Connection) Send(chID byte, msg []byte) bool {
	if !c.IsRunning() {
		return false
	}
	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.Logger.Error("want to send message to the specified channel, but cannot find this channel", "channel id", chID)
		return false
	}
	success := channel.sendBytes(msg)
	if success {
		c.Logger.Debug("send message to the specified channel", "channel id", chID, "msg", fmt.Sprintf("%X", msg))
		select {
		case c.sendChan <- struct{}{}:
		// 告诉sendRoutine来活了，有数据需要发送
		default:
		}
	} else {
		c.Logger.Warn("send message failed", "channel id", chID)
	}
	return success
}

// TrySend ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// TrySend 尝试向指定的信道发送数据。
func (c *Connection) TrySend(chID byte, msg []byte) bool {
	if !c.IsRunning() {
		return false
	}
	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.Logger.Error("try to send message to the specified channel, but cannot find this channel", "channel id", chID)
		return false
	}
	success := channel.trySendBytes(msg)
	if success {
		c.Logger.Debug("successfully try to send message to the specified channel", "channel id", chID)
		select {
		case c.sendChan <- struct{}{}:
		// 提醒sendRoutine协程有数据需要发送
		default:
		}
	}
	return success
}

// CanSend ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// CanSend 询问指定的信道能否发送数据。
func (c *Connection) CanSend(chID byte) bool {
	if !c.IsRunning() {
		return false
	}
	channel, ok := c.channelsIdx[chID]
	if !ok {
		c.Logger.Error("cannot find the specified channel", "channel id", chID)
		return false
	}
	return channel.canSend()
}

func (c *Connection) sendRoutine() {
	defer c.recover()
	protoWriter := protoio.NewDelimitedWriter(c.bufConnWriter)
LOOP:
	for {
		var err error
	SELECTION:
		select {
		case <-c.flusher.Ch:
			// flusher.fireRoutine被启动了，common/flusher那个地方发送了信号过来，让我们把数据刷新到网络连接里。
			c.flush()
		case <-c.chStatsTicker.C:
			// 该更新每个信道的状态了
			for _, channel := range c.channels {
				channel.updateStats()
			}
		case <-c.pingTicker.C:
			// 该给对方发送ping消息了
			if err = c.sendPing(protoWriter); err != nil {
				break SELECTION
			}
		case timeout := <-c.pongTimeoutCh:
			if timeout {
				c.Logger.Warn("failed to wait for a pong message within the timeout period")
				err = errors.New("failed to wait for a pong message within the timeout period")
			} else {
				c.stopPongTimer()
			}
		case <-c.pong:
			if err = c.sendPong(protoWriter); err != nil {
				break SELECTION
			}
		case <-c.quitSendRoutine:
			break LOOP
		case <-c.sendChan:
			eof := c.sendSomePacketMsgs()
			if !eof {
				select {
				case c.sendChan <- struct{}{}:
					// 实现自举，继续发送信号告诉自己还有数据未发送完。
				default:
				}
			}
		}
		if !c.IsRunning() {
			break LOOP
		}
		if err != nil {
			c.Logger.Error("connection failed @ sendRoutine", "err", err)
			c.stopForError(err)
			break LOOP
		}
	}
	c.stopPongTimer()
	close(c.doneSendRoutine)
}

func (c *Connection) sendSomePacketMsgs() bool {
	c.sendMonitor.Limit()
	for i := 0; i < numBatchPacketMsgs; i++ {
		if c.sendPacketMsg() {
			return true // 没有数据要发送了
		}
	}
	return false // 还有数据要发送
}

func (c *Connection) sendPacketMsg() bool {
	var leastRatio float64 = math.MaxFloat64
	var chosenChannel *Channel
	for _, channel := range c.channels {
		if !channel.isSendPending() {
			continue
		}
		// 最近发送的数据越少，信道级别越高，ratio就越小，就越可能选择这个信道来发送数据
		ratio := float64(channel.recentlySent) / float64(channel.desc.Priority)
		if ratio < leastRatio {
			leastRatio = ratio
			chosenChannel = channel
		}
	}
	if chosenChannel == nil {
		// 所有信道都没有数据要发送
		return true
	}
	n, err := chosenChannel.writePacketMsgTo(c.bufConnWriter)
	if err != nil {
		c.Logger.Error("failed to write PacketMsg in the specified channel", "channel id", chosenChannel.desc.ID, "err", err)
		c.stopForError(err)
		return true
	}
	c.sendMonitor.Update(n)
	c.flusher.Set()
	return false // 返回false表明还有信道等着发送数据
}

// sendPing ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// sendPing 向对方发送ping消息。
func (c *Connection) sendPing(writer protoio.Writer) error {
	c.Logger.Debug("send ping")
	n, err := writer.WriteMsg(wrapPacket(&pbp2p.PacketPing{}))
	if err != nil {
		c.Logger.Error("failed to send ping", "err", err)
		return err
	}
	c.sendMonitor.Update(n)
	c.Logger.Debug("wait for pong message from the peer", "wait time", c.config.PongTimeout)
	c.pongTimer = time.AfterFunc(c.config.PongTimeout, func() {
		select {
		case c.pongTimeoutCh <- true:
			// 超时时间到了还没收到pong消息，那么就往通道里发送true信号
		default:
		}
	})
	c.flush()
	return nil
}

// sendPong ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// sendPong 向对方回复pong消息。
func (c *Connection) sendPong(writer protoio.Writer) error {
	c.Logger.Debug("send pong")
	n, err := writer.WriteMsg(wrapPacket(&pbp2p.PacketPong{}))
	if err != nil {
		c.Logger.Error("failed to send pong message")
		return err
	}
	c.sendMonitor.Update(n)
	c.flush()
	return nil
}

func (c *Connection) recvRoutine() {
	defer c.recover()
	protoReader := protoio.NewDelimitedReader(c.bufConnReader, c._maxPacketMsgSize)
LOOP:
	for {
		c.recvMonitor.Limit()
		var packet pbp2p.Packet
		n, err := protoReader.ReadMsg(&packet)
		c.recvMonitor.Update(n)
		if err != nil {
			select {
			case <-c.quitRecvRoutine:
				break LOOP
			default:
			}
			if c.IsRunning() {
				if err == io.EOF {
					c.Logger.Info("connection is closed @ recvRoutine by the other side")
				} else {
					c.Logger.Error("connection failed @ recvRoutine", "err", err)
				}
				c.stopForError(err)
			}
			break LOOP
		}
		switch pkt := packet.Sum.(type) {
		case *pbp2p.Packet_PacketPing:
			c.Logger.Debug("receive ping")
			select {
			case c.pong <- struct{}{}:
			default:
			}
		case *pbp2p.Packet_PacketPong:
			c.Logger.Debug("receive pong")
			select {
			case c.pongTimeoutCh <- false:
			default:
			}
		case *pbp2p.Packet_PacketMsg:
			if err = c.handlePacketMsg(pkt.PacketMsg); err != nil {
				c.stopForError(err)
				break LOOP
			}
		default:
			err = fmt.Errorf("unknown message type %q", reflect.TypeOf(packet))
			c.Logger.Error("connection failed @ recvRoutine", "err", err)
			c.stopForError(err)
			break LOOP
		}
	}
}

func (c *Connection) handlePacketMsg(msg *pbp2p.PacketMsg) (err error) {
	channelID := byte(msg.ChannelID)
	channel, ok := c.channelsIdx[channelID]
	if !ok {
		err = fmt.Errorf("no channel %X can handle this packet message", channelID)
		c.Logger.Error("connection failed @ recvRoutine", "err", err)
		return err
	}
	msgBytes, err := channel.recvPacketMsg(*msg)
	if err != nil {
		if c.IsRunning() {
			c.Logger.Error("connection failed @ recvRoutine", "err", err)
		}
		return err
	}
	if msgBytes != nil {
		c.Logger.Debug("receive packet message from channel", "channel id", channelID)
		c.onReceive(channelID, msgBytes)
	}
	return nil
}

// stopPongTimer ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// stopPongTimer 一般在收到对方发来的pong消息后关闭该计时器。
func (c *Connection) stopPongTimer() {
	if c.pongTimer != nil {
		_ = c.pongTimer.Stop()
		c.pongTimer = nil
	}
}

// maxPacketMsgSize ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// maxPacketMsgSize 包装一个数据包，该数据包的载荷设置成最大，然后求这个数据包的整体大小，就是最大数据包大小。
func (c *Connection) maxPacketMsgSize() int {
	msg := wrapPacket(&pbp2p.PacketMsg{
		ChannelID: 0xff,
		EOF:       true,
		Data:      make([]byte, c.config.MaxPacketMsgPayloadSize),
	})
	bz, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return len(bz)
}

// Channel ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Channel 信道。
type Channel struct {
	conn                    *Connection
	desc                    ChannelDescriptor
	sendQueue               chan []byte
	sendQueueSize           int32
	recving                 []byte
	sending                 []byte
	recentlySent            int64
	maxPacketMsgPayloadSize int // 最近发送的字节数，遵循指数回退变化规则
	Logger                  log.Logger
}

// ChannelDescriptor ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ChannelDescriptor 用来表明信道的信息。
type ChannelDescriptor struct {
	ID                  byte
	Priority            int
	SendQueueCapacity   int
	RecvBufferCapacity  int
	RecvMessageCapacity int
}

func (des *ChannelDescriptor) FillDefaults() {
	if des.SendQueueCapacity == 0 {
		des.SendQueueCapacity = defaultSendQueueCapacity
	}
	if des.RecvBufferCapacity == 0 {
		des.RecvBufferCapacity = defaultRecvBufferCapacity
	}
	if des.RecvMessageCapacity == 0 {
		des.RecvMessageCapacity = defaultRecvMessageCapacity
	}
}

// sendBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// sendBytes 往信道里发送数据，如果信道里数据是满的，则等待10秒，如果10秒后信道里还是满的，则超时，直接退出。
func (ch *Channel) sendBytes(bz []byte) bool {
	select {
	case ch.sendQueue <- bz:
		// 往发送队列里推送数据，然后队列的长度相应的也要加一
		atomic.AddInt32(&ch.sendQueueSize, 1)
		return true
	case <-time.After(defaultSendTimeout):
		return false
	}
}

// trySendBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// trySendBytes 尝试往信道里发送数据，如果信道里是满的，则直接退出。
func (ch *Channel) trySendBytes(bz []byte) bool {
	select {
	case ch.sendQueue <- bz:
		atomic.AddInt32(&ch.sendQueueSize, 1)
		return true
	default:
		return false
	}
}

// isSendPending ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// isSendPending 判断信道里是否还有数据需要发送出去。
func (ch *Channel) isSendPending() bool {
	if len(ch.sending) == 0 {
		if len(ch.sendQueue) == 0 {
			return false
		}
		ch.sending = <-ch.sendQueue
	}
	return true
}

// nextPacketMsg ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// nextPacketMsg 将信道里的数据打包成数据包，并返回，信道里的数据可能比较多，一次性打包不完，那么剩下的
// 就下次再打包发送出去。
func (ch *Channel) nextPacketMsg() pbp2p.PacketMsg {
	packet := pbp2p.PacketMsg{ChannelID: int32(ch.desc.ID)}
	maxSize := ch.maxPacketMsgPayloadSize
	packet.Data = ch.sending[:number.MinInt(maxSize, len(ch.sending))]
	if len(ch.sending) <= maxSize {
		packet.EOF = true
		ch.sending = nil
		atomic.AddInt32(&ch.sendQueueSize, -1) // 信道里的数据发送干净了
	} else {
		packet.EOF = false
		ch.sending = ch.sending[len(packet.Data):]
	}
	return packet
}

// writePacketMsgTo ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// writePacketMsgTo 从信道里打包一个数据包，然后写入到io.Writer里，这里的io.Writer，
// 实际上是 bufio.NewWriter(net.Conn)。
func (ch *Channel) writePacketMsgTo(w io.Writer) (n int, err error) {
	packet := ch.nextPacketMsg()
	n, err = protoio.NewDelimitedWriter(w).WriteMsg(wrapPacket(&packet))
	atomic.AddInt64(&ch.recentlySent, int64(n))
	return n, err
}

// recvPacketMsg ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// recvPacketMsg 解析收到的数据包，并将数据包里的数据返回出来。
func (ch *Channel) recvPacketMsg(packet pbp2p.PacketMsg) ([]byte, error) {
	var recvCap, received = ch.desc.RecvMessageCapacity, len(ch.recving) + len(packet.Data)
	if recvCap < received {
		return nil, fmt.Errorf("received message exceeds available capacity: %v < %v", recvCap, received)
	}
	ch.recving = append(ch.recving, packet.Data...)
	if packet.EOF {
		// 对方发送的数据包已经完整接收到了
		msg := ch.recving
		ch.recving = ch.recving[:0]
		return msg, nil
	}
	// 对方发送的数据包还没有完整接收到
	return nil, nil
}

// canSend ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// canSend 判断信道是否还能再发数据，默认情况下，信道的发送队列里只能存储一条数据。
func (ch *Channel) canSend() bool {
	queueSize := int(atomic.LoadInt32(&ch.sendQueueSize))
	return queueSize < defaultSendQueueCapacity
}

// updateStats ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// updateStats 将信道已经发送的数据总和乘上0.8，相当于指数回退了。
func (ch *Channel) updateStats() {
	atomic.StoreInt64(&ch.recentlySent, int64(float64(atomic.LoadInt64(&ch.recentlySent))*0.8))
}

type ChannelStatus struct {
	ID                byte
	SendQueueCapacity int
	SendQueueSize     int
	Priority          int
	RecentlySent      int64
}

// ConnectionStatus ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ConnectionStatus 反应P2P网络底层连接状态的结构体。
type ConnectionStatus struct {
	Duration    time.Duration
	SendMonitor flowrate.Status
	RecvMonitor flowrate.Status
	Channels    []ChannelStatus
}

func (c *Connection) Status() ConnectionStatus {
	var status ConnectionStatus
	status.Duration = time.Since(c.created)
	status.SendMonitor = c.sendMonitor.Status()
	status.RecvMonitor = c.recvMonitor.Status()
	status.Channels = make([]ChannelStatus, len(c.channels))
	for i := 0; i < len(c.channels); i++ {
		status.Channels[i].ID = c.channels[i].desc.ID
		status.Channels[i].Priority = c.channels[i].desc.Priority
		status.Channels[i].SendQueueCapacity = c.channels[i].desc.SendQueueCapacity
		status.Channels[i].SendQueueSize = int(atomic.LoadInt32(&c.channels[i].sendQueueSize))
		status.Channels[i].RecentlySent = atomic.LoadInt64(&c.channels[i].recentlySent)
	}
	return status
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义常量

const (
	minReadBufferSize          = 1024
	minWriteBufferSize         = 65536
	defaultSendQueueCapacity   = 1        // 信道的默认发送队列大小等于1
	defaultRecvBufferCapacity  = 4096     // 信道的默认存放接收数据的缓冲区大小为4096
	defaultRecvMessageCapacity = 22020096 // 信道默认能够接收的一条消息的大小为22020096
	updateStats                = 2 * time.Second
	defaultSendTimeout         = 10 * time.Second // 信道里如果消息满了，则可以等待10秒钟，这10秒里如果信道里的消息还没被发送出去，那么就超时了
	numBatchPacketMsgs         = 10               // 一次可以批量发送10个数据包
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// wrapPacket ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// wrapPacket 包装数据包。
func wrapPacket(pb proto.Message) (msg *pbp2p.Packet) {
	switch pb := pb.(type) {
	case *pbp2p.Packet:
		msg = pb
	case *pbp2p.PacketMsg:
		msg = &pbp2p.Packet{Sum: &pbp2p.Packet_PacketMsg{PacketMsg: pb}}
	case *pbp2p.PacketPing:
		msg = &pbp2p.Packet{Sum: &pbp2p.Packet_PacketPing{PacketPing: pb}}
	case *pbp2p.PacketPong:
		msg = &pbp2p.Packet{Sum: &pbp2p.Packet_PacketPong{PacketPong: pb}}
	default:
		panic(fmt.Errorf("unknown packet type %T", pb))
	}
	return msg
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 加密通信

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

type SecretConnection struct {
	net.Conn
	recvAead     cipher.AEAD
	sendAead     cipher.AEAD
	remPublicKey crypto.PublicKey
	recvMu       sync.Mutex
	recvBuffer   []byte
	recvNonce    *[aeadNonceSize]byte
	sendMu       sync.Mutex
	sendNonce    *[aeadNonceSize]byte
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 加密通信用到的常量

const (
	aeadNonceSize = chacha20poly1305.NonceSize
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量

// authSig ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// authSig 在验证公钥的过程中用到的签名结构体。
type authSig struct {
	Key *bls12.PublicKey
	Sig []byte
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// shareAuthSignature ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// shareAuthSignature 互相交换认证签名。
func shareAuthSignature(sc net.Conn, publicKey *bls12.PublicKey, signature []byte) (recvMsg authSig, err error) {
	var taskSet, _ = async.Parallel(
		func(i int) (val interface{}, abort bool, err error) {

		})
}

// incrNonce ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// incrNonce 递增nonce值。
func incrNonce(nonce *[aeadNonceSize]byte) {
	counter := binary.LittleEndian.Uint64(nonce[4:])
	if counter == math.MaxUint64 {
		counter = 0
	}
	counter++
	binary.LittleEndian.PutUint64(nonce[4:], counter)
}
