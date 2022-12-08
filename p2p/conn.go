package p2p

import (
	"bufio"
	"fmt"
	"github.com/232425wxy/meta--/common/flowrate"
	"github.com/232425wxy/meta--/common/flusher"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/proto/pbp2p"
	"github.com/cosmos/gogoproto/proto"
	"net"
	"sync"
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
	onReceive         receiveCb
	onError           errorCb
	errored           uint32
	config            ConnectionConfig
	quitSendRoutine   chan struct{}
	doneSendRoutine   chan struct{}
	quitRecvRoutine   chan struct{}
	stopMu            sync.Mutex
	flusher           flusher.Flusher
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
	c.BaseService.SetLogger(logger)
	for _, ch := range c.channels {
		ch.Logger = logger
	}
}

// maxPacketMsgSize ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// maxPacketMsgSize 包装一个数据包，该数据包的载荷设置成最大，然后求这个数据包的整体大小，就是最大数据包大小。
func (c *Connection) maxPacketMsgSize() int {
	msg := wrapPacket(&pbp2p.PacketMsg{
		ChannelID: 0x00,
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

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义常量

const (
	minReadBufferSize          = 1024
	minWriteBufferSize         = 65536
	defaultSendQueueCapacity   = 1        // 信道的默认发送队列大小等于1
	defaultRecvBufferCapacity  = 4096     // 信道的默认存放接收数据的缓冲区大小为4096
	defaultRecvMessageCapacity = 22020096 // 信道默认能够接收的一条消息的大小为22020096
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
