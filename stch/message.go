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

type FnX struct {
	From crypto.ID
	Data *big.Int
	X    *big.Int // 对方的身份标识
}

func (fx *FnX) ToProto() *pbstch.FnX {
	if fx == nil {
		return nil
	}
	return &pbstch.FnX{
		From: string(fx.From),
		Data: fx.Data.Bytes(),
		X:    fx.X.Bytes(),
	}
}

func FnXFromProto(pb *pbstch.FnX) *FnX {
	if pb == nil {
		return nil
	}
	return &FnX{
		From: crypto.ID(pb.From),
		Data: new(big.Int).SetBytes(pb.Data),
		X:    new(big.Int).SetBytes(pb.X),
	}
}

func (fx *FnX) ChameleonFn() {

}

type PublicKeySeg struct {
	From      crypto.ID
	PublicKey *big.Int
}

func (pks *PublicKeySeg) ToProto() *pbstch.PublicKeySeg {
	if pks == nil {
		return nil
	}
	return &pbstch.PublicKeySeg{
		From:      string(pks.From),
		PublicKey: pks.PublicKey.Bytes(),
	}
}

func PublicKeySegFromProto(pb *pbstch.PublicKeySeg) *PublicKeySeg {
	if pb == nil {
		return nil
	}
	return &PublicKeySeg{
		From:      crypto.ID(pb.From),
		PublicKey: new(big.Int).SetBytes(pb.PublicKey),
	}
}

func (pks *PublicKeySeg) ChameleonFn() {

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
	case *FnX:
		pb.Data = &pbstch.Message_Fnx{Fnx: msg.ToProto()}
	case *PublicKeySeg:
		pb.Data = &pbstch.Message_PublicKeySeg{PublicKeySeg: msg.ToProto()}
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
	case *pbstch.Message_Fnx:
		msg = FnXFromProto(data.Fnx)
	case *pbstch.Message_PublicKeySeg:
		msg = PublicKeySegFromProto(data.PublicKeySeg)
	default:
		panic(fmt.Sprintf("unknown message type: %T", data))
	}

	return msg
}
