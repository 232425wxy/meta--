package crypto

import (
	"encoding/hex"
	"fmt"
	"sync"
)

// ID ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ID 节点的唯一身份标识，对节点公钥的前10个字节内容进行
// 16进制编码就得到了节点的ID，ID的字符串长度等于20。
type ID string

// ToBytes ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// ToBytes 返回ID的字节切片形式，ID是长度为20的字符串，利用 hex.DecodeString 方法解码，获得
// 长度为10的字节切片。
func (id ID) ToBytes() []byte {
	bz, _ := hex.DecodeString(string(id))
	return bz
}

// FromBytesToID ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// FromBytesToID 给定ID的字节切片形式，利用 hex.EncodeToString 方法将其编码为16进制的字符串。
func FromBytesToID(bz []byte) (ID, error) {
	if len(bz) != 20 {
		return "", fmt.Errorf("cannot convert bytes to ID: %q", "the length of the given bytes is not 20")
	}
	return ID(hex.EncodeToString(bz)), nil
}

// IDSet ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// IDSet 节点id的集合，可以往里面加入节点id，也可以判断一个节点的id在不在里面。
type IDSet struct {
	mu  sync.RWMutex
	IDs []ID
}

// NewIDSet ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// NewIDSet 实例化一个空的ID集合。
func NewIDSet(size int) *IDSet {
	return &IDSet{IDs: make([]ID, size)}
}

// Contains ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Contains 给定一个节点的ID，判断该ID在不在集合中。
func (set *IDSet) Contains(id ID) int {
	set.mu.RLock()
	defer set.mu.RUnlock()
	for i, _id := range set.IDs {
		if _id == id {
			return i
		}
	}
	return -1
}

// AddID ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// AddID 给定一个节点的ID，将其添加到集合中。
func (set *IDSet) AddID(id ID) {
	if set.Contains(id) > 0 {
		return
	}
	set.mu.Lock()
	defer set.mu.Unlock()
	set.IDs = append(set.IDs, id)
}

// RemoveID ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// RemoveID 给定一个节点的ID，将其从集合中删除。
func (set *IDSet) RemoveID(id ID) {
	index := set.Contains(id)

	if index < 0 {
		// 不存在，直接返回
		return
	}
	set.mu.Lock()
	defer set.mu.Unlock()
	if index == len(set.IDs)-1 {
		set.IDs = set.IDs[:len(set.IDs)-1]
		return
	}
	set.IDs = append(set.IDs[:index], set.IDs[index+1:]...)
}

// Size ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Size 返回集合中有多少个节点的ID。
func (set *IDSet) Size() int {
	return len(set.IDs)
}
