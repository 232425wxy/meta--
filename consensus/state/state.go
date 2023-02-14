package state

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/crypto"
	"github.com/232425wxy/meta--/database"
	"github.com/232425wxy/meta--/proto/pbstate"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"github.com/232425wxy/meta--/stch"
	"github.com/232425wxy/meta--/store"
	"github.com/232425wxy/meta--/types"
	"github.com/cosmos/gogoproto/proto"
	"time"
)

var StoreStateKey = []byte("meta--/store-state")
var ValidatorsKey = []byte("meta--/state/validators")

type State struct {
	InitialHeight   int64
	LastBlockHeight int64
	PreviousBlock   *types.Block
	LastBlockTime   time.Time
	Validators      *types.ValidatorSet
	BlockStore      *store.BlockStore
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

func (s *State) SetBlockStore(store *store.BlockStore) {
	s.BlockStore = store
}

func (s *State) MakeBlock(height int64, txs []types.Tx, proposer crypto.ID, lastBlockHash []byte) *types.Block {
	block := &types.Block{
		Header: &types.Header{PreviousBlockHash: lastBlockHash, Height: height, Timestamp: time.Now(), Proposer: proposer},
		Body:   &types.Data{Txs: txs},
	}
	//_txs := make([][]byte, len(txs))
	//for i, tx := range txs {
	//	_txs[i] = tx
	//}
	//block.Body.RootHash = merkle.ComputeMerkleRoot(_txs)
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

func (s *State) RedactBlock(height int64, txIndex int, key, value []byte) {
	//block := s.BlockStore.LoadBlockByHeight(height)
	task := &stch.Task{
		BlockHeight: height,
		TxIndex:     txIndex,
		Key:         key,
		Value:       value,
	}
	s.Chameleon.AppendRedactTask(task)
}

type StoreState struct {
	db database.DB
}

func NewStoreState(db database.DB) *StoreState {
	return &StoreState{db: db}
}

func (s *StoreState) LoadFromDBOrGenesis(genesis *types.Genesis) *State {
	stat, err := s.LoadState()
	if err != nil {
		return &State{}
	}
	if stat.IsEmpty() {
		stat = MakeGenesisState(genesis)
	}
	return stat
}

func (s *StoreState) LoadState() (*State, error) {
	bz, err := s.db.Get(StoreStateKey)
	if err != nil {
		return nil, err
	}
	if len(bz) == 0 {
		return &State{}, nil
	}
	pb := new(pbstate.State)
	err = proto.Unmarshal(bz, pb)
	if err != nil {
		return nil, err
	}
	stat := StateFromProto(pb)
	return stat, nil
}

func (s *StoreState) SaveState(stat *State) error {
	return s.db.SetSync(StoreStateKey, stat.ToBytes())
}

func (s *StoreState) Bootstrap(stat *State) error {
	return s.SaveState(stat)
}

func (s *StoreState) SaveValidators(height int64, validators *types.ValidatorSet) error {
	pb := validators.ToProto()
	bz, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	return s.db.SetSync(calcValidatorsKey(height), bz)
}

func (s *StoreState) LoadValidators(height int64) (*types.ValidatorSet, error) {
	bz, err := s.db.Get(calcValidatorsKey(height))
	if err != nil {
		return nil, err
	}
	if len(bz) == 0 {
		return nil, errors.New("validators retrieved from db is empty")
	}
	pb := &pbtypes.ValidatorSet{}
	if err = proto.Unmarshal(bz, pb); err != nil {
		return nil, err
	}
	return types.ValidatorSetFromProto(pb), nil
}

func calcValidatorsKey(height int64) []byte {
	return append(ValidatorsKey, fmt.Sprintf("%d", height)...)
}
