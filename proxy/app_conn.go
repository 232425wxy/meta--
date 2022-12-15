package proxy

import (
	"github.com/232425wxy/meta--/abci"
	"github.com/232425wxy/meta--/proto/pbabci"
)

type AppConnTxsPool interface {
	SetResponseCallback(callback abci.Callback)
	Error() error
	CheckTx(pbabci.RequestCheckTx) *abci.ReqRes
}
