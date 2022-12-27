package proxy

import (
	"github.com/232425wxy/meta--/abci"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/proto/pbabci"
	"sync"
)

type AppConnTxsPool struct {
	service.BaseService
	mu  sync.Mutex
	app abci.Application
}

func NewAppConnTxsPool(app abci.Application) *AppConnTxsPool {
	return &AppConnTxsPool{
		BaseService: *service.NewBaseService(nil, "Proxy-Application-Txs-Pool"),
		app:         app,
	}
}

func (ac *AppConnTxsPool) CheckTx(req pbabci.RequestCheckTx) pbabci.ResponseCheckTx {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	return ac.app.CheckTx(req)
}
