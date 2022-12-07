package p2p

import (
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/proto/pbp2p"
	"net"
	"strconv"
	"strings"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义网络地址

// NetAddress ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NetAddress 网络地址由节点的ID、IP地址和监听的端口号组成：id@ip:port。
type NetAddress struct {
	ID   crypto.ID `json:"ID"`
	IP   net.IP    `json:"IP"`
	Port int       `json:"Port"`
}

// IDAddressString ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// IDAddressString 给定节点的id和配置文件中的地址，例如：tcp://0.0.0.0"25556，然后组合成新的
// 字符串：id@0.0.0.0:25556。
func IDAddressString(id crypto.ID, addr string) string {
	if strings.Contains(addr, "://") {
		return fmt.Sprintf("%s@%s", id, strings.Split(addr, "://")[1])
	}
	return fmt.Sprintf("%s@%s", id, addr)
}

// NewNetAddress ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewNetAddress 给定节点的id和节点的网络地址，生成NetAddress实例，tendermint里面还要验证节点id的合法性，
// 我觉得根本没必要，多此一举。
func NewNetAddress(id crypto.ID, addr net.Addr) *NetAddress {
	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		panic("only support tcp network")
	}
	return &NetAddress{
		ID:   id,
		IP:   tcpAddr.IP,
		Port: tcpAddr.Port,
	}
}

// NewNetAddressString ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewNetAddressString 该方法以一个"id@ip:port"格式的字符串作为输入，然后解析该字符串，实例化一个
// NetAddress 对象，换句话说，该方法常以 IDAddressString 方法的返回值作为输入。
func NewNetAddressString(str string) (*NetAddress, error) {
	strs := strings.Split(str, "@")
	var id crypto.ID = crypto.ID(strs[0])
	addr := strs[1]
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		return nil, err
	}
	return &NetAddress{
		ID:   id,
		IP:   ip,
		Port: int(port),
	}, nil
}

// NewNetAddressStrings ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewNetAddressStrings 批量调用 NetAddressString 方法为一堆地址生成 NetAddress 实例。
func NewNetAddressStrings(addrs []string) ([]*NetAddress, []error) {
	result := make([]*NetAddress, 0)
	errs := make([]error, 0)
	for _, str := range addrs {
		addr, err := NewNetAddressString(str)
		if err != nil {
			errs = append(errs, err)
		} else {
			result = append(result, addr)
		}
	}
	return result, errs
}

// ToProto ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToProto 将NetAddress转换为protobuf。
func (addr *NetAddress) ToProto() pbp2p.NetAddress {
	return pbp2p.NetAddress{
		ID:   string(addr.ID),
		IP:   addr.IP.String(),
		Port: int64(addr.Port),
	}
}

// NetAddressFromProto ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NetAddressFromProto 将protobuf形式的NetAddress转换为自定义的NetAddress。
func NetAddressFromProto(pbAddr pbp2p.NetAddress) *NetAddress {
	return &NetAddress{
		ID:   crypto.ID(pbAddr.ID),
		IP:   net.ParseIP(pbAddr.IP),
		Port: int(pbAddr.Port),
	}
}
