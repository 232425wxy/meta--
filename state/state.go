package state

import (
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/crypto/merkle"
	"github.com/232425wxy/meta--/proto/pbstate"
	"github.com/232425wxy/meta--/types"
	"github.com/cosmos/gogoproto/proto"
	"time"
)

type State struct {
	InitialHeight   int64
	LastBlockHeight int64
	LastBlock       *types.SimpleBlock
	LastBlockTime   time.Time
	Validators      *types.ValidatorSet
}

func (s *State) Copy() *State {
	return &State{
		InitialHeight:   s.InitialHeight,
		LastBlockHeight: s.LastBlockHeight,
		LastBlock:       s.LastBlock,
		LastBlockTime:   s.LastBlockTime,
		Validators:      s.Validators.Copy(),
	}
}

func (s *State) MakeBlock(height int64, txs []types.Tx, proposer crypto.ID, lastBlockHash []byte) *types.Block {
	block := &types.Block{
		LastBlock: types.SimpleBlock{Hash: lastBlockHash},
		Header:    types.Header{Height: height, Timestamp: time.Now(), Proposer: proposer},
		Data:      types.Data{Txs: txs},
	}
	_txs := make([][]byte, len(txs))
	for i, tx := range txs {
		copy(_txs[i], tx)
	}
	block.Data.RootHash = merkle.ComputeMerkleRoot(_txs)
	block.Hash()
	return block
}

func (s *State) MakeGenesisState(gen *types.Genesis) (*State, error) {
	return &State{
		InitialHeight:   gen.InitialHeight,
		LastBlockHeight: 0,
		LastBlock:       &types.SimpleBlock{},
		LastBlockTime:   gen.GenesisTime,
		Validators:      types.NewValidatorSet(gen.Validators),
	}, nil
}

func (s *State) ToProto() *pbstate.State {
	if s == nil {
		return nil
	}
	return &pbstate.State{
		InitialHeight:   s.InitialHeight,
		LastBlockHeight: s.LastBlockHeight,
		LastBlock:       s.LastBlock.ToProto(),
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
		LastBlock:       types.SimpleBlockFromProto(pb.LastBlock),
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
