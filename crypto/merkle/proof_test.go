package merkle

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetSplitPoint(t *testing.T) {
	s := getSplitPoint(4)
	t.Log(s)
	s = getSplitPoint(8)
	t.Log(s)
	s = getSplitPoint(9)
	t.Log(s)
	s = getSplitPoint(20)
	t.Log(s)
}

func TestTrailsByteSlices(t *testing.T) {
	trails, root := trailsFromByteSlices(items)
	t.Log(len(trails))
	t.Log(root)
}

// 26条交易数据
var items = [][]byte{
	[]byte("a"),
	[]byte("b"),
	[]byte("c"),
	[]byte("d"),
	[]byte("e"),
	[]byte("f"),
	[]byte("g"),
	[]byte("h"),
	[]byte("i"),
	[]byte("j"),
	[]byte("k"),
	[]byte("l"),
	[]byte("m"),
	[]byte("n"),
	[]byte("o"),
	[]byte("p"),
	[]byte("q"),
	[]byte("r"),
	[]byte("s"),
	[]byte("t"),
	[]byte("u"),
	[]byte("v"),
	[]byte("w"),
	[]byte("x"),
	[]byte("y"),
	[]byte("z"),
}

func TestComputeRootHash(t *testing.T) {
	root, proofs := ProofsFromByteSlices(items)
	t.Log(root)
	for _, proof := range proofs {
		assert.Equal(t, proof.ComputeRootHash(), root)
	}
}

func BenchmarkCompareTendermint(b *testing.B) {
	_, proofs := ProofsFromByteSlices(items)
	for i := 0; i < b.N; i++ {
		for _, proof := range proofs {
			proof.ComputeRootHash()
		}
	}
	b.ReportAllocs()
}

func TestProofString(t *testing.T) {
	_, proofs := ProofsFromByteSlices(items)
	fmt.Println(proofs[0])
}

func TestComputeMerkleRoot(t *testing.T) {
	root1, _ := ProofsFromByteSlices(items)
	root2 := ComputeMerkleRoot(items)
	assert.Equal(t, root2, root1)
}

func BenchmarkCompareProofsToCompute(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ComputeMerkleRoot(items)
		// 30802	     39289 ns/op	   20048 B/op	     358 allocs/op
	}
	for i := 0; i < b.N; i++ {
		ProofsFromByteSlices(items)
		// 27400	     44545 ns/op	   26224 B/op	     610 allocs/op
	}
	b.ReportAllocs()
}
