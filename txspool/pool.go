package txspool

import (
	"github.com/232425wxy/meta--/config"
	"sync"
)

type TxsPool struct {
	cfg               *config.TxsPoolConfig
	height            int64
	txsBytes          int
	notifiedAvailable bool
	txsAvailable      chan struct{}
	updateMu          sync.RWMutex
}
