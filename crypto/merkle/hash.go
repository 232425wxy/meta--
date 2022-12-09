package merkle

import (
	"github.com/232425wxy/meta--/crypto/sha256"
)

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义包级的全局函数

// emptyHash ♏ | (o゜▽゜)o☆吴翔宇
//
// emptyHash 计算[]byte{}（空字节切片）的哈希值。
func emptyHash() []byte {
	h := sha256.Sum([]byte{})
	return h[:]
}

// leafHash ♏ | (o゜▽゜)o☆吴翔宇
//
// leafHash 方法接受一个字节切片leaf作为输入参数，该方法在构建默克尔树时用于计算叶子节点的哈希值。
//
//	sha256.Sum(append([]byte{0}, leaf...))
func leafHash(leaf []byte) []byte {
	h := sha256.Sum(append(leafPrefix, leaf...))
	return h[:]
}

// innerHash ♏ | (o゜▽゜)o☆吴翔宇
//
// innerHash 方法接受两个字节切片right和left作为输入参数，该方法将right和left拼接起来，计算拼
// 接结果的哈希值。
//
//	sha256.Sum(append([]byte{1}, append(right, left...)...))
func innerHash(left, right []byte) []byte {
	h := sha256.Sum(append(innerPrefix, append(left, right...)...))
	return h[:]
}

/*⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓⛓*/

// 定义计算默克尔树时需要用到的变量

// leafPrefix ♏ | (o゜▽゜)o☆吴翔宇
//
// leafPrefix 在计算叶子节点时用得上
var leafPrefix = []byte{0}

// innerPrefix ♏ | (o゜▽゜)o☆吴翔宇
//
// innerPrefix 在计算左右节点的哈希值的时候用得上。
var innerPrefix = []byte{1}
