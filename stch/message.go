package stch

import (
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/proto/pbstch"
	"github.com/cosmos/gogoproto/proto"
	"math/big"
)

type Message interface {
	ChameleonFn()
}

type IdentityX struct {
	X  *big.Int
	ID crypto.ID
}

func (ix *IdentityX) ToProto() *pbstch.IdentityX {
	if ix == nil {
		return nil
	}
	return &pbstch.IdentityX{
		X:  ix.X.Bytes(),
		ID: string(ix.ID),
	}
}

func IdentityXFromProto(pb *pbstch.IdentityX) *IdentityX {
	if pb == nil {
		return nil
	}
	return &IdentityX{
		X:  new(big.Int).SetBytes(pb.X),
		ID: crypto.ID(pb.ID),
	}
}

func (ix *IdentityX) ChameleonFn() {

}

///////////////////////////////////////////////

func MustEncode(message Message) []byte {
	if message == nil {
		panic("message is nil")
	}
	var pb = &pbstch.Message{}
	switch msg := message.(type) {
	case *IdentityX:
		pb.Data = &pbstch.Message_IdentityX{IdentityX: msg.ToProto()}
	default:
		panic(fmt.Sprintf("unknown message type: %T", msg))
	}
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return bz
}

func MustDecode(bz []byte) (msg Message) {
	if len(bz) == 0 {
		panic("message is empty")
	}
	var pb = &pbstch.Message{}
	if err := proto.Unmarshal(bz, pb); err != nil {
		panic(err)
	}

	switch data := pb.Data.(type) {
	case *pbstch.Message_IdentityX:
		msg = IdentityXFromProto(data.IdentityX)
	default:
		panic(fmt.Sprintf("unknown message type: %T", data))
	}

	return msg
}
