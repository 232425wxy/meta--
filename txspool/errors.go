package txspool

import (
	"fmt"
	"github.com/232425wxy/meta--/types"
)

type ErrorTxAlreadyExists struct {
	Tx types.Tx
}

func (e *ErrorTxAlreadyExists) Error() string {
	return fmt.Sprintf("tx %x already exists", e.Tx)
}
