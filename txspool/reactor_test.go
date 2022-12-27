package txspool

import (
	"github.com/232425wxy/meta--/common/clist"
	"github.com/232425wxy/meta--/common/cmap"
	"github.com/232425wxy/meta--/proto/pbtxspool"
	"testing"
)

func TestLargestTx(t *testing.T) {
	largestTx := make([]byte, 1024*1024*3)
	msg := pbtxspool.Message{Txs: &pbtxspool.Txs{Txs: [][]byte{largestTx}}}
	t.Log(msg.Size())
}

func TestElement(t *testing.T) {
	txs := clist.NewList()
	ptx := &poolTx{
		tx:      []byte("hello"),
		senders: cmap.NewCap(),
		height:  10,
	}
	ptx.senders.Set("test-v", struct{}{})

	elem := txs.Push(ptx)

	m := cmap.NewCap()
	m.Set(txKey(ptx.tx), elem)

	element := m.Get(txKey(ptx.tx)).(*clist.Element)
	tx := element.Value.(*poolTx)
	tx.senders.Set("test-x", struct{}{})

	t.Log(ptx.senders.Size())
}
