package apps

import (
	"bytes"
	"github.com/232425wxy/meta--/abci"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/bls12"
	"github.com/232425wxy/meta--/database"
	"github.com/232425wxy/meta--/proto/pbabci"
	"github.com/cosmos/gogoproto/proto"
	"sync"
)

type KVStoreApp struct {
	mu         *sync.RWMutex
	height     int64
	validators map[crypto.ID]pbabci.ValidatorUpdate
	db         database.DB
}

func NewKVStoreApp(name, dir string, backend database.BackendType) abci.Application {
	db, err := database.NewDB(name, dir, backend)
	if err != nil {
		panic(err)
	}
	return &KVStoreApp{
		height:     0,
		validators: make(map[crypto.ID]pbabci.ValidatorUpdate),
		db:         db,
	}
}

func (k *KVStoreApp) Info(req pbabci.RequestInfo) pbabci.ResponseInfo {
	return pbabci.ResponseInfo{Type: "kv-store"}
}

func (k *KVStoreApp) Echo(req pbabci.RequestEcho) pbabci.ResponseEcho {
	return pbabci.ResponseEcho{Message: "echo"}
}

// InitChain ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// InitChain 更新验证者集合。
func (k *KVStoreApp) InitChain(req pbabci.RequestInitChain) pbabci.ResponseInitChain {
	for _, update := range req.ValidatorUpdates {
		publicKey := bls12.PublicKeyFromProto(update.BLS12PublicKey)
		if update.Power <= 0 {
			// 需要将投票权等于0的投票者从系统中删除
			if err := k.db.Delete(publicKey.ToBytes()); err != nil {
				panic(err)
			}
			delete(k.validators, publicKey.ToID())
		} else {
			var value []byte
			value, err := proto.Marshal(&update)
			if err != nil {
				panic(err)
			}
			if err = k.db.Set(publicKey.ToBytes(), value); err != nil {
				panic(err)
			}
			k.validators[publicKey.ToID()] = update
		}
	}
	res := pbabci.ResponseInitChain{ValidatorUpdates: make([]*pbabci.ValidatorUpdate, 0)}
	for _, validator := range k.validators {
		res.ValidatorUpdates = append(res.ValidatorUpdates, &validator)
	}
	return res
}

func (k *KVStoreApp) Query(req pbabci.RequestQuery) pbabci.ResponseQuery {
	switch req.Path {
	case "/validator.proto":
		// 查验证者信息
		value, err := k.db.Get(req.Data)
		if err != nil {
			panic(err)
		}
		return pbabci.ResponseQuery{Height: k.height, Key: req.Data, Value: value}
	default:
		// 查交易信息
		return pbabci.ResponseQuery{}
	}
}

func (k *KVStoreApp) CheckTx(req pbabci.RequestCheckTx) pbabci.ResponseCheckTx {
	return pbabci.ResponseCheckTx{OK: true}
}

func (k *KVStoreApp) DeliverTx(req pbabci.RequestDeliverTx) pbabci.ResponseDeliverTx {
	// 交易数据tx是一对键值对，形式为："key=value"
	var key, value []byte
	var res pbabci.ResponseDeliverTx
	s := bytes.Split(req.Tx, []byte("="))
	if len(s) == 2 {
		key, value = s[0], s[1]
		res.OK = true
	} else {
		res.OK = false
	}
	key = append([]byte("tx:"), key...)
	if err := k.db.Set(key, value); err != nil {
		res.OK = false
	}
	return res
}

// BeginBlock ♏ | 作者 ⇨ 吴翔宇 | (｡･∀･)ﾉﾞ嗨
//
// BeginBlock 对犯错的validator进行惩罚。
func (k *KVStoreApp) BeginBlock(req pbabci.RequestBeginBlock) pbabci.ResponseBeginBlock {
	for _, evidence := range req.Evidences {
		val := evidence.Validator
		publicKey := bls12.PublicKeyFromProto(val.BLS12PublicKey)
		k.validators[publicKey.ToID()] = pbabci.ValidatorUpdate{
			BLS12PublicKey: val.BLS12PublicKey,
			Power:          k.validators[publicKey.ToID()].Power - 1,
		}
	}
	return pbabci.ResponseBeginBlock{OK: true}
}

func (k *KVStoreApp) EndBlock(req pbabci.RequestEndBlock) pbabci.ResponseEndBlock {
	res := pbabci.ResponseEndBlock{Height: k.height, ValidatorUpdates: make([]*pbabci.ValidatorUpdate, 0)}
	for _, validator := range k.validators {
		res.ValidatorUpdates = append(res.ValidatorUpdates, &validator)
	}
	return res
}

func (k *KVStoreApp) Commit(req pbabci.RequestCommit) pbabci.ResponseCommit {
	return pbabci.ResponseCommit{OK: true}
}

func (k *KVStoreApp) Redact(req pbabci.RequestRedact) pbabci.ResponseRedact {
	//TODO implement me
	panic("implement me")
}

var _ abci.Application = &KVStoreApp{}
