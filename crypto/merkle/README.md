# 默克尔树

## 简介

构建区块中所有交易的默克尔树，构建的默克尔树最高不能超过**100**层，也就是说，一个区块里含有的交易数量不能超过2^100。

## 用法

- **仅计算交易的根哈希值**

>func ComputeMerkleRoot(items [][]byte) []byte
> 
> items是区块里的交易数据集合，返回值是由items构建的默克尔树的根哈希值。

```go
items := [][]byte{[]byte{'a'}, []byte{'b'}}
root := ComputeMerkleRoot(items)
```

上面代码里`items`是交易数据集合，里面只有两条交易，`root`就是`items`构建的默克尔树的根哈希值。

- **计算交易的根哈希值和每个交易的Proof**

>func ProofsFromByteSlices(items [][]byte) (rootHash []byte, proofs []*Proof)
> 
> items参数是区块里的交易数据集合，返回的第一个参数是由items构建的默克尔树的根哈希值，第二个参数是每笔交易的Proof，Proof可以用来验证交易是否被篡改。