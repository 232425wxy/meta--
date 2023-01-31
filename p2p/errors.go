package p2p

import (
	"fmt"
	"github.com/232425wxy/meta--/crypto"
)

type ErrorAlreadyHasPeer struct {
	id crypto.ID
}

func (e *ErrorAlreadyHasPeer) Error() string {
	return fmt.Sprintf("already has this peer: peer_id = %s", e.id)
}

type ErrorDialingOrAlreadyHas struct {
	address string
}

func (e *ErrorDialingOrAlreadyHas) Error() string {
	return fmt.Sprintf("dialing or already has this peer: peer_address = %s", e.address)
}
