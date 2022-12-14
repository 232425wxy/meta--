package sha256

import (
	"crypto/sha256"
	"encoding/hex"
	"hash"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// API 接口

// New ♏ | (o゜▽゜)o☆吴翔宇
//
// New 实例化一个sha256哈希函数。
func New() hash.Hash {
	return sha256.New()
}

// Sum ♏ | (o゜▽゜)o☆吴翔宇
//
// Sum 方法接受一个字节切片bz，然后返回该切片的sha256哈希值，哈希值的长度为32字节。
func Sum(bz []byte) Hash {
	return sha256.Sum256(bz)
}

// Sum20 ♏ | (o゜▽゜)o☆吴翔宇
//
// Sum20 方法接受一个字节切片bz作为输入参数，然后计算该切片的sha256哈希值，但是只返
// 回哈希值的前20个字节。
func Sum20(bz []byte) [20]byte {
	res := [20]byte{}
	sum := sha256.Sum256(bz)
	copy(res[:], sum[:])
	return res
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 辅助变量

// Hash ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// Hash 定义一个长度为32的字节数组。
type Hash [32]byte

// String ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
//	---------------------------------------------------------
//
// String 返回32字节长度数组的16进制编码结果。
func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 项目级可导出全局变量

// Size32 ♏ | (o゜▽゜)o☆吴翔宇
//
// Size32 sha256哈希值的长度，固定为32字节，256比特。
const Size32 = 32
