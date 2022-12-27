package abci

import (
	"github.com/232425wxy/meta--/proto/pbabci"
)

type Application interface {
	Info(pbabci.RequestInfo) pbabci.ResponseInfo // 返回代理应用的信息
	Echo(pbabci.RequestEcho) pbabci.ResponseEcho
	InitChain(pbabci.RequestInitChain) pbabci.ResponseInitChain
	Query(pbabci.RequestQuery) pbabci.ResponseQuery
	CheckTx(pbabci.RequestCheckTx) pbabci.ResponseCheckTx
	DeliverTx(pbabci.RequestDeliverTx) pbabci.ResponseDeliverTx
	BeginBlock(pbabci.RequestBeginBlock) pbabci.ResponseBeginBlock
	EndBlock(pbabci.RequestEndBlock) pbabci.ResponseEndBlock
	Commit(pbabci.RequestCommit) pbabci.ResponseCommit
	Redact(pbabci.RequestRedact) pbabci.ResponseRedact
}
