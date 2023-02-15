package stch

import (
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/proto/pbstch"
	"github.com/232425wxy/meta--/types"
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

func (ix *IdentityX) ChameleonFn() {}

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

func (fx *FnX) ChameleonFn() {}

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

func (pks *PublicKeySeg) ChameleonFn() {}

type AlphaExpKAndHK struct {
	AlphaExpK *big.Int
	HK        *big.Int
}

func (ah *AlphaExpKAndHK) ToProto() *pbstch.AlphaExpKAndHK {
	if ah == nil {
		return nil
	}
	return &pbstch.AlphaExpKAndHK{
		AlphaExpK: ah.AlphaExpK.Bytes(),
		HK:        ah.HK.Bytes(),
	}
}

func AlphaExpKAndHKFromProto(pb *pbstch.AlphaExpKAndHK) *AlphaExpKAndHK {
	if pb == nil {
		return nil
	}
	return &AlphaExpKAndHK{
		AlphaExpK: new(big.Int).SetBytes(pb.AlphaExpK),
		HK:        new(big.Int).SetBytes(pb.HK),
	}
}

func (ah *AlphaExpKAndHK) ChameleonFn() {}

type LeaderSchnorrSig struct {
	Flag        bool // 标志S是否是负数
	S           *big.Int
	D           *big.Int
	BlockHeight int64
	TxIndex     int
	NewTx       types.Tx
}

func (ss *LeaderSchnorrSig) ToProto() *pbstch.SchnorrSig {
	if ss == nil {
		return nil
	}
	return &pbstch.SchnorrSig{
		Flag:        ss.Flag,
		From:        pbstch.From_Leader,
		S:           ss.S.Bytes(),
		D:           ss.D.Bytes(),
		BlockHeight: ss.BlockHeight,
		TxIndex:     int64(ss.TxIndex),
		Tx:          ss.NewTx,
	}
}

func LeaderSchnorrSigFromProto(pb *pbstch.SchnorrSig) *LeaderSchnorrSig {
	if pb == nil {
		return nil
	}
	return &LeaderSchnorrSig{
		Flag:        pb.Flag,
		S:           new(big.Int).SetBytes(pb.S),
		D:           new(big.Int).SetBytes(pb.D),
		BlockHeight: pb.BlockHeight,
		TxIndex:     int(pb.TxIndex),
		NewTx:       pb.Tx,
	}
}

func (ss *LeaderSchnorrSig) ChameleonFn() {}

type ReplicaSchnorrSig struct {
	Flag        bool // 标志S是否是负数
	S           *big.Int
	D           *big.Int
	BlockHeight int64
	TxIndex     int
	NewTx       types.Tx
}

func (ss *ReplicaSchnorrSig) ToProto() *pbstch.SchnorrSig {
	if ss == nil {
		return nil
	}
	return &pbstch.SchnorrSig{
		Flag:        ss.Flag,
		From:        pbstch.From_Replica,
		S:           ss.S.Bytes(),
		D:           ss.D.Bytes(),
		BlockHeight: ss.BlockHeight,
		TxIndex:     int64(ss.TxIndex),
		Tx:          ss.NewTx,
	}
}

func ReplicaSchnorrSigFromProto(pb *pbstch.SchnorrSig) *ReplicaSchnorrSig {
	if pb == nil {
		return nil
	}
	return &ReplicaSchnorrSig{
		Flag:        pb.Flag,
		S:           new(big.Int).SetBytes(pb.S),
		D:           new(big.Int).SetBytes(pb.D),
		BlockHeight: pb.BlockHeight,
		TxIndex:     int(pb.TxIndex),
		NewTx:       pb.Tx,
	}
}

func (ss *ReplicaSchnorrSig) ChameleonFn() {}

type RandomVerification struct {
	GSigmaExpSK *big.Int
	RedactName  string
	R2          *big.Int
}

func (fv *RandomVerification) ToProto() *pbstch.FinalVer {
	if fv == nil {
		return nil
	}
	return &pbstch.FinalVer{
		Val:       fv.GSigmaExpSK.Bytes(),
		RedactStr: fv.RedactName,
		R2:        fv.R2.Bytes(),
	}
}

func FinalVerFromProto(pb *pbstch.FinalVer) *RandomVerification {
	if pb == nil {
		return nil
	}
	return &RandomVerification{
		GSigmaExpSK: new(big.Int).SetBytes(pb.Val),
		RedactName:  pb.RedactStr,
		R2:          new(big.Int).SetBytes(pb.R2),
	}
}

func (fv *RandomVerification) ChameleonFn() {}

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
	case *LeaderSchnorrSig:
		pb.Data = &pbstch.Message_SchnorrSig{SchnorrSig: msg.ToProto()}
	case *ReplicaSchnorrSig:
		pb.Data = &pbstch.Message_SchnorrSig{SchnorrSig: msg.ToProto()}
	case *AlphaExpKAndHK:
		pb.Data = &pbstch.Message_AlphaExpKAndHK{AlphaExpKAndHK: msg.ToProto()}
	case *RandomVerification:
		pb.Data = &pbstch.Message_FinalVer{FinalVer: msg.ToProto()}
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
	case *pbstch.Message_SchnorrSig:
		switch data.SchnorrSig.From {
		case pbstch.From_Leader:
			msg = LeaderSchnorrSigFromProto(data.SchnorrSig)
		case pbstch.From_Replica:
			msg = ReplicaSchnorrSigFromProto(data.SchnorrSig)
		}
	case *pbstch.Message_AlphaExpKAndHK:
		msg = AlphaExpKAndHKFromProto(data.AlphaExpKAndHK)
	case *pbstch.Message_FinalVer:
		msg = FinalVerFromProto(data.FinalVer)
	default:
		panic(fmt.Sprintf("unknown message type: %T", data))
	}

	return msg
}
