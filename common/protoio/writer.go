package protoio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/cosmos/gogoproto/proto"
	"io"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 项目级全局函数

// NewDelimitedWriter ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewDelimitedWriter 实例化一个 *writer，该对象实现了 WriteMsg 方法，可以将message序列化成字
// 节切片，然后将其写入到底层的io.Writer里，例如p2p包的net.Conn，然后对等方利用 reader 的 ReadMsg
// 方法读取数据，重新打包成 proto.Message。
func NewDelimitedWriter(w io.Writer) Writer {
	return &writer{w: w}
}

// DelimitedWriteToBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// DelimitedWriteToBytes 方法接受一个 proto.Message 作为参数，然后将其序列化成字节切片并返回出来。
func DelimitedWriteToBytes(msg proto.Message) ([]byte, error) {
	buf := &bytes.Buffer{}
	_, err := NewDelimitedWriter(buf).WriteMsg(msg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义一个writer，可以将proto.Message消息写入到底层的io.Writer里

type Writer interface {
	WriteMsg(msg proto.Message) (int, error)
	Close() error
}

type writer struct {
	w io.Writer
}

// WriteMsg ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// WriteMsg 方法接受一个 proto.Message 作为参数，该方法将message序列化成字节切片，然后将其
// 写入到底层的io.Writer里，例如p2p包的net.Conn，然后对等方利用 reader 的 ReadMsg 方法读取
// 数据，重新打包成 proto.Message。
func (w *writer) WriteMsg(msg proto.Message) (int, error) {
	if m, ok := msg.(interface{ MarshalTo([]byte) (int, error) }); ok {
		size, ok := getSize(m)
		if ok {
			buf := make([]byte, size+binary.MaxVarintLen64)
			lenOff := binary.PutUvarint(buf, uint64(size))
			n, err := m.MarshalTo(buf[lenOff:])
			if n != size {
				return 0, fmt.Errorf("protoio: get TxsNumInPool %q, but marshal size %q", size, n)
			}
			if err != nil {
				return 0, err
			}
			_, err = w.w.Write(buf[:lenOff+n])
			return lenOff + n, err
		}
	}

	// 无法调用MarshalTo方法来序列化msg
	data, err := proto.Marshal(msg)
	if err != nil {
		return 0, err
	}
	length := uint64(len(data))
	buf := make([]byte, binary.MaxVarintLen64+len(data))
	lenOff := binary.PutUvarint(buf, length)
	n := copy(buf[lenOff:], data)
	return w.w.Write(buf[:lenOff+n])
}

// Close ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Close 如果 writer 中的io.Writer实现了Close方法，那么就调用Close方法。
func (w *writer) Close() error {
	if closer, ok := w.w.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 不可导出的工具函数

// getSize ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// getSize 方法接受一个实例对象，如果该对象实现了 TxsNumInPool() int 方法或者 ProtoSize() int 方法，就调用这些
// 方法，来获得该对象的大小。
func getSize(v interface{}) (int, bool) {
	if sz, ok := v.(interface{ Size() int }); ok {
		return sz.Size(), true
	} else if sz, ok := v.(interface{ ProtoSize() int }); ok {
		return sz.ProtoSize(), true
	} else {
		return 0, false
	}
}
