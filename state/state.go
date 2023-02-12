package state

import (
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/merkle"
	"github.com/232425wxy/meta--/proto/pbstate"
	"github.com/232425wxy/meta--/stch"
	"github.com/232425wxy/meta--/types"
	"github.com/cosmos/gogoproto/proto"
	"time"
)

var StoreStateKey = []byte("meta--/store-state")
var StoreBlockKey = []byte("meta--/store-block")
var ValidatorsKey = []byte("meta--/state/validators")

type State struct {
	InitialHeight   int64
	LastBlockHeight int64
	PreviousBlock   *types.Block
	LastBlockTime   time.Time
	Validators      *types.ValidatorSet
	Chameleon       *stch.Chameleon
}

func (s *State) Copy() *State {
	return &State{
		InitialHeight:   s.InitialHeight,
		LastBlockHeight: s.LastBlockHeight,
		PreviousBlock:   s.PreviousBlock,
		LastBlockTime:   s.LastBlockTime,
		Validators:      s.Validators.Copy(),
	}
}

func (s *State) SetChameleon(ch *stch.Chameleon) {
	s.Chameleon = ch
}

func (s *State) MakeBlock(height int64, txs []types.Tx, proposer crypto.ID, lastBlockHash []byte) *types.Block {
	block := &types.Block{
		Header: &types.Header{PreviousBlockHash: lastBlockHash, Height: height, Timestamp: time.Now(), Proposer: proposer},
		Body:   &types.Data{Txs: txs},
	}
	_txs := make([][]byte, len(txs))
	for i, tx := range txs {
		copy(_txs[i], tx)
	}
	block.Body.RootHash = merkle.ComputeMerkleRoot(_txs)
	s.Chameleon.Hash(block)
	return block
}

func MakeGenesisState(gen *types.Genesis) *State {
	return &State{
		InitialHeight:   gen.InitialHeight,
		LastBlockHeight: 0,
		PreviousBlock:   &types.Block{},
		LastBlockTime:   gen.GenesisTime,
		Validators:      types.NewValidatorSet(gen.Validators),
	}
}

func (s *State) ToProto() *pbstate.State {
	if s == nil {
		return nil
	}
	return &pbstate.State{
		InitialHeight:   s.InitialHeight,
		LastBlockHeight: s.LastBlockHeight,
		PreviousBlock:   s.PreviousBlock.ToProto(),
		LastBlockTime:   s.LastBlockTime,
		Validators:      s.Validators.ToProto(),
	}
}

func StateFromProto(pb *pbstate.State) *State {
	if pb == nil {
		return nil
	}
	return &State{
		InitialHeight:   pb.InitialHeight,
		LastBlockHeight: pb.LastBlockHeight,
		PreviousBlock:   types.BlockFromProto(pb.PreviousBlock),
		LastBlockTime:   pb.LastBlockTime,
		Validators:      types.ValidatorSetFromProto(pb.Validators),
	}
}

func (s *State) ToBytes() []byte {
	pb := s.ToProto()
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return bz
}

func (s *State) IsEmpty() bool {
	return s.Validators == nil || len(s.Validators.Validators) == 0
}
