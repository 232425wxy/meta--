package apps

import (
	"github.com/232425wxy/meta--/abci"
	"github.com/232425wxy/meta--/proto/pbabci"
	"sync"
)

type KVStoreApp struct {
	mu       *sync.RWMutex
	callback abci.Callback
}

func (k *KVStoreApp) Info(req pbabci.RequestInfo) pbabci.ResponseInfo {
	//TODO implement me
	panic("implement me")
}

func (k *KVStoreApp) Echo(req pbabci.RequestEcho) pbabci.ResponseEcho {
	//TODO implement me
	panic("implement me")
}

func (k *KVStoreApp) InitChain(req pbabci.RequestInitChain) pbabci.ResponseInitChain {
	//TODO implement me
	panic("implement me")
}

func (k *KVStoreApp) Query(req pbabci.RequestQuery) pbabci.ResponseQuery {
	//TODO implement me
	panic("implement me")
}

func (k *KVStoreApp) CheckTx(req pbabci.RequestCheckTx) pbabci.ResponseCheckTx {
	//TODO implement me
	panic("implement me")
}

func (k *KVStoreApp) DeliverTx(req pbabci.RequestDeliverTx) pbabci.ResponseDeliverTx {
	//TODO implement me
	panic("implement me")
}

func (k *KVStoreApp) BeginBlock(req pbabci.RequestBeginBlock) pbabci.ResponseBeginBlock {
	//TODO implement me
	panic("implement me")
}

func (k *KVStoreApp) EndBlock(req pbabci.RequestEndBlock) pbabci.ResponseEndBlock {
	//TODO implement me
	panic("implement me")
}

func (k *KVStoreApp) Commit(req pbabci.RequestCommit) pbabci.ResponseCommit {
	//TODO implement me
	panic("implement me")
}

func (k *KVStoreApp) Redact(req pbabci.RequestRedact) pbabci.ResponseRedact {
	//TODO implement me
	panic("implement me")
}

var _ abci.Application = &KVStoreApp{}
