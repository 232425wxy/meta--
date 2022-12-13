package config

import (
	"path/filepath"
	"time"
)

// P2PConfig ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// P2PConfig 结构体定义了p2p模块中一些常用的配置：
//  1. 配置文件的存储位置：Home
//  2. 节点监听的地址：ListenAddress
//  3. 存储地址簿内容的文件地址：AddrBook
//  4. 将信道中的数据发送给对方节点的时间周期：FlushDuration
//  5. 数据包的最大载荷：MaxPacketPayloadSize
//  6. 发送数据的速率上限：SendRate
//  7. 接收数据的速率上限：RecvRate
type P2PConfig struct {
	Home                    string        `mapstructure:"home"`
	ListenAddress           string        `mapstructure:"listen_address"`
	AddrBook                string        `mapstructure:"addr_book"`
	FlushDuration           time.Duration `mapstructure:"flush_duration"`
	MaxPacketMsgPayloadSize int           `mapstructure:"max_packet_msg_payload_size"`
	SendRate                int64         `mapstructure:"send_rate"`
	RecvRate                int64         `mapstructure:"recv_rate"`
}

func DefaultP2PConfig() *P2PConfig {
	return &P2PConfig{
		ListenAddress:           "tcp://0.0.0.0:26656",
		AddrBook:                defaultAddrBookPath, // config/addrbook.json
		FlushDuration:           100 * time.Millisecond,
		MaxPacketMsgPayloadSize: 1024,
		SendRate:                5120000,
		RecvRate:                5120000,
	}
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 包级变量

var (
	defaultConfigDir = "config"

	defaultAddrBookName = "addrbook.json"

	defaultAddrBookPath = filepath.Join(defaultConfigDir, defaultAddrBookName)
)
