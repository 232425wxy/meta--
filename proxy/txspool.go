package proxy

import (
	"github.com/232425wxy/meta--/abci"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/log"
	"github.com/232425wxy/meta--/proto/pbabci"
	"sync"
)

type AppConnTxsPool struct {
	service.BaseService
	mu  sync.Mutex
	app abci.Application
}

func NewAppConnTxsPool(app abci.Application, logger log.Logger) *AppConnTxsPool {
	return &AppConnTxsPool{
		BaseService: *service.NewBaseService(logger, "Proxy-Application-Txs-Pool"),
		mu:          sync.Mutex{},
		app:         app,
	}
}

func (ac *AppConnTxsPool) CheckTx(req pbabci.RequestCheckTx) pbabci.ResponseCheckTx {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	//ac.Logger.Debug("check tx", "tx", fmt.Sprintf("%x", req.Tx))
	return ac.app.CheckTx(req)
}
