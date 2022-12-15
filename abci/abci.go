package abci

import (
	"github.com/232425wxy/meta--/proto/pbabci"
	"sync"
)

type Application interface {
	Info(pbabci.RequestInfo) pbabci.ResponseInfo
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

type Callback func(*pbabci.Request, *pbabci.Response)

type ReqRes struct {
	request  *pbabci.Request
	response *pbabci.Response
	wg       *sync.WaitGroup
	mu       sync.RWMutex
	done     bool
	callback func(*pbabci.Response)
}

func NewReqRes(req *pbabci.Request) *ReqRes {
	return &ReqRes{
		request:  req,
		wg:       waitGroup1(),
		callback: nil,
	}
}

func (r *ReqRes) SetCallback(cb func(*pbabci.Response)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.done {
		cb(r.response)
		return
	}
	r.callback = cb
}

func (r *ReqRes) SetDone() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.done = true
}

func (r *ReqRes) InvokeCallback() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.callback != nil {
		r.callback(r.response)
	}
}

func (r *ReqRes) GetCallback() func(*pbabci.Response) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.callback
}

func waitGroup1() *sync.WaitGroup {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return wg
}
