package txspool

import (
	"fmt"
	"github.com/232425wxy/meta--/abci/apps"
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/database"
	"github.com/232425wxy/meta--/proxy"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestPool(t *testing.T) {
	app := apps.NewKVStoreApp("kvstore", "data", database.GoLevelDBBackend)
	proxyer := proxy.NewAppConnTxsPool(app, nil)
	pool := NewTxsPool(config.DefaultTxsPoolConfig(), proxyer, 0)
	txs := make([][]byte, 0)
	for i := 0; i < 1024; i++ {
		tx := []byte(fmt.Sprintf("key:%d=value:%d", i, i))
		txs = append(txs, tx)
	}

	go func() {
		for {
			select {
			case <-pool.TxsAvailable():
				_txs := pool.ReapMaxBytes(4096)
				for _, tx := range _txs {
					t.Log(string(tx))
				}
				pool.Update(0, _txs)
			}
		}
	}()

	for _, tx := range txs {
		err := pool.CheckTx(tx, "test-v")
		assert.Nil(t, err)
	}
	//err := pool.CheckTx([]byte(fmt.Sprintf("key:%d=value:%d", 10, 10)), "test-x")
	//assert.NotNil(t, err)
	time.Sleep(time.Second * 3)
	t.Log(pool.TxsNumInPool())
}
