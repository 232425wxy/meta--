package consensus

import "github.com/232425wxy/meta--/crypto"

type Message interface {
	ValidateBasic() error
}

type MessageInfo struct {
	Msg    Message   `json:"msg"`
	NodeID crypto.ID `json:"node_id"`
}
