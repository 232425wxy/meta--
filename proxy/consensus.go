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

func (app *AppConnConsensus) InitChain(chain pbabci.RequestInitChain) pbabci.ResponseInitChain {
	return pbabci.ResponseInitChain{}
}

func (app *AppConnConsensus) BeginBlock(block pbabci.RequestBeginBlock) pbabci.ResponseBeginBlock {
	return pbabci.ResponseBeginBlock{}
}

func (app *AppConnConsensus) DeliverTx(tx pbabci.RequestDeliverTx) pbabci.ResponseDeliverTx {
	return pbabci.ResponseDeliverTx{}
}

func (app *AppConnConsensus) EndBlock(block pbabci.RequestEndBlock) pbabci.ResponseEndBlock {
	return pbabci.ResponseEndBlock{}
}

func (app *AppConnConsensus) Commit(commit pbabci.RequestCommit) pbabci.ResponseCommit {
	return pbabci.ResponseCommit{}
}
