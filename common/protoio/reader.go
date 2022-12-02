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

// NewDelimitedReader ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewDelimitedReader 给定一个具有能够从底层数据池读取数据的io.Reader，例如net.Conn，然后根据它
// 新建一个reader，这个reader可以从底层数据池里读取数据，将其构造成一个proto.Message数据。
func NewDelimitedReader(r io.Reader, maxSize int) Reader {
	return &reader{r: r, maxSize: maxSize, bytesRead: 0}
}

// DelimitedReadFromData ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// DelimitedFromData 给定一个数据data，它是字节切片形式的，然后我们需要从data数据里提炼出
// proto.Message对象，所以我们需要先构建一个基于data的io.Reader，然后再新建一个 reader 来
// 调用 ReadMsg 方法实现上述目标。
func DelimitedReadFromData(data []byte, msg proto.Message) error {
	buf := bytes.NewReader(data)
	r := NewDelimitedReader(buf, len(data))
	_, err := r.ReadMsg(msg)
	return err
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义一个reader，可以实现ReadByte函数

type Reader interface {
	ReadMsg(msg proto.Message) (int, error)
	Close() error
}

// reader ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// reader 是一个结构体，内部有一个 bytes.Buffer 字段实现读取数据的能力，它可以一次读取一个字节。
type reader struct {
	r         io.Reader
	maxSize   int
	bytesRead int
}

// ReadByte ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ReadByte 方法一次只从数据缓冲区中读取一个字节，每读取一个字节，就给读取的字节数加1。
func (r *reader) ReadByte() (byte, error) {
	bz := make([]byte, 1)
	n, err := r.r.Read(bz)
	r.bytesRead += n
	if err != nil {
		return 0x00, err
	}
	return bz[0], nil
}

// ReadMsg ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ReadMsg 该方法主要在网络传输中使用，从对等方读取网络传输的数据，然后将数据解码到给定的 proto.Message 中，
// 返回读取的数据字节数和中间可能产生的错误。
func (r *reader) ReadMsg(msg proto.Message) (int, error) {
	l, err := binary.ReadUvarint(r)
	defer func() {
		r.bytesRead = 0
	}()
	if err != nil {
		return r.bytesRead, err
	}
	length := int(l)
	if l >= uint64(^uint(0)>>1) || length < 0 || r.bytesRead+length < 0 {
		return r.bytesRead, fmt.Errorf("protoio: invalid out-of-range message length %v", l)
	}
	if length > r.maxSize {
		return r.bytesRead, fmt.Errorf("protoio: message exceeds max size (%d > %d)", length, r.maxSize)
	}
	buf := make([]byte, length)
	n, err := io.ReadFull(r.r, buf)
	if err != nil {
		return n + r.bytesRead, err
	}
	return n + r.bytesRead, proto.Unmarshal(buf, msg)
}

// Close ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Close 如果reader里的 io.Reader 可以被close掉，那就把它close掉。
func (r *reader) Close() error {
	if closer, ok := r.r.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
