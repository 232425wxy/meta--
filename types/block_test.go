package types

import (
	"fmt"
	"io"
	"testing"
)

func TestName(t *testing.T) {
	txs := make(Txs, 4)
	for i := 0; i < 4; i++ {
		txs[i] = []byte(fmt.Sprintf("num=%d", i))
	}

	_txs := make([][]byte, len(txs))
	for i, tx := range txs {
		_txs[i] = tx
		//copy(_txs[i], tx)
	}

	for i := 0; i < 4; i++ {
		fmt.Println(_txs[i])
	}
}

func TestEof(t *testing.T) {
	t.Log(io.EOF)
}
