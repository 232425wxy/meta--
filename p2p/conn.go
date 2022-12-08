package p2p

import (
	"bufio"
	"github.com/232425wxy/meta--/common/flowrate"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/log"
	"net"
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
	conn          net.Conn // 还是要靠net.Conn来和对方取得联系的
	bufConnReader *bufio.Reader
	bufConnWriter *bufio.Writer
	sendMonitor   *flowrate.Monitor
	recvMonitor   *flowrate.Monitor
	sendChan      chan struct{}
	pong          chan struct{}
	channels      []*Channel
	channelsIdx   map[byte]*Channel
	onReceive     receiveCb
	onError       errorCb
	errored       uint32
}

// ConnectionConfig ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//  ---------------------------------------------------------
// ConnectionConfig p2p连接的配置信息。
type ConnectionConfig struct {
	SendRate                int64         `mapstructure:"send_rate"`
	RecvRate                int64         `mapstructure:"recv_rate"`
	MaxPacketMsgPayloadSize int           `mapstructure:"max_packet_msg_payload_size"`
	Flusher                 time.Duration `mapstructure:"flusher"`
	PingInterval            time.Duration `mapstructure:"ping_interval"`
	PongTimeout             time.Duration `mapstructure:"pong_timeout"`
}

// receiveCb ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//  ---------------------------------------------------------
// receiveCb 收到消息后要干什么呢，这个就交给对应的信道去处理吧。
type receiveCb func(chID byte, msg []byte)

// errorCb ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//  ---------------------------------------------------------
// errorCb 发生错误了该怎么办呢？这也交给对应的回调函数去处理吧。
type errorCb func(err error)

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
