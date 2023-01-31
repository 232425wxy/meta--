package proxy

import (
	"github.com/232425wxy/meta--/abci"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/log"
)

type AppConns struct {
	service.BaseService
	consensus   *AppConnConsensus
	txspool     *AppConnTxsPool
	application abci.Application
}

func NewAppConns(application abci.Application, logger log.Logger) *AppConns {
	conns := &AppConns{application: application}
	conns.BaseService = *service.NewBaseService(nil, "AppConns")
	conns.SetLogger(logger.New("module", "proxy_conns"))
	return conns
}

func (conns *AppConns) Consensus() *AppConnConsensus {
	return conns.consensus
}

func (conns *AppConns) TxsPool() *AppConnTxsPool {
	return conns.txspool
}

func (conns *AppConns) Start() error {
	conns.consensus = NewAppConnConsensus(conns.application)
	conns.txspool = NewAppConnTxsPool(conns.application, conns.Logger.New("module", "ProxyAppTxsPool"))
	return conns.BaseService.Start()
}
