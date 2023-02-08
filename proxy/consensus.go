package proxy

import (
	"github.com/232425wxy/meta--/abci"
	"github.com/232425wxy/meta--/common/service"
	"github.com/232425wxy/meta--/proto/pbabci"
)

type AppConnConsensus struct {
	service.BaseService
	application abci.Application
}

func NewAppConnConsensus(app abci.Application) *AppConnConsensus {
	return &AppConnConsensus{
		BaseService: *service.NewBaseService(nil, "AppConn_Consensus"),
		application: app,
	}
}

func (app *AppConnConsensus) InitChain(req pbabci.RequestInitChain) pbabci.ResponseInitChain {
	return app.application.InitChain(req)
}

func (app *AppConnConsensus) BeginBlock(req pbabci.RequestBeginBlock) pbabci.ResponseBeginBlock {
	return app.application.BeginBlock(req)
}

func (app *AppConnConsensus) DeliverTx(req pbabci.RequestDeliverTx) pbabci.ResponseDeliverTx {
	return app.application.DeliverTx(req)
}

func (app *AppConnConsensus) EndBlock(req pbabci.RequestEndBlock) pbabci.ResponseEndBlock {
	return app.application.EndBlock(req)
}

func (app *AppConnConsensus) Commit(req pbabci.RequestCommit) pbabci.ResponseCommit {
	return app.application.Commit(req)
}
