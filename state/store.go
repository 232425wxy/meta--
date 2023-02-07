package state

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/database"
	"github.com/232425wxy/meta--/proto/pbstate"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"github.com/232425wxy/meta--/types"
	"github.com/cosmos/gogoproto/proto"
	"sync"
)

type StoreState struct {
	db database.DB
}

func NewStoreState(db database.DB) *StoreState {
	return &StoreState{db: db}
}

func (s *StoreState) LoadFromDBOrGenesis(genesis *types.Genesis) *State {
	state, err := s.LoadState()
	if err != nil {
		return &State{}
	}
	if state.IsEmpty() {
		state = MakeGenesisState(genesis)
	}
	return state
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
	state := StateFromProto(pb)
	return state, nil
}

func (s *StoreState) SaveState(state *State) error {
	return s.db.SetSync(StoreStateKey, state.ToBytes())
}

func (s *StoreState) Bootstrap(state *State) error {
	return s.SaveState(state)
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

/**********************************************************************************************************************/

type StoreBlock struct {
	db     database.DB
	mu     sync.RWMutex
	height int64
}

func NewStoreBlock(db database.DB) *StoreBlock {
	sb := new(StoreBlock)
	bz, err := db.Get(StoreBlockKey)
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		sb.height = 0
		sb.db = db
		return sb
	}
	pb := new(pbstate.StoreBlock)
	if err = proto.Unmarshal(bz, pb); err != nil {
		panic(err)
	}
	return &StoreBlock{db: db, height: pb.Height}
}

// Height
//
// Height 反映当前区块链的高度和区块数量。
func (sb *StoreBlock) Height() int64 {
	sb.mu.RLock()
	sb.mu.RUnlock()
	return sb.height
}

func (sb *StoreBlock) LoadBlockByHeight(height int64) *types.Block {
	pb := &pbtypes.Block{}
	bz, err := sb.db.Get(calcBlockHeightKey(height))
	if err != nil {
		panic(err)
	}
	if err = proto.Unmarshal(bz, pb); err != nil {
		panic(err)
	}
	block := &types.Block{
		Header: types.HeaderFromProto(pb.Header),
		Body:   types.DataFromProto(pb.Body),
	}
	return block
}

func (sb *StoreBlock) LoadBlockByHash(hash []byte) *types.Block {
	pb := &pbtypes.BlockHeight{}
	bz, err := sb.db.Get(calcBlockHashKey(hash))
	if err != nil {
		panic(err)
	}
	if err = proto.Unmarshal(bz, pb); err != nil {
		panic(err)
	}
	return sb.LoadBlockByHeight(pb.Height)
}

func (sb *StoreBlock) SaveBlock(block *types.Block) {
	if block == nil {
		panic("cannot save nil block")
	}
	pb := block.ToProto()
	bz, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}
	bh := &pbtypes.BlockHeight{Height: block.Header.Height}
	bzh, err := proto.Marshal(bh)
	if err != nil {
		panic(err)
	}
	if err = sb.db.SetSync(calcBlockHashKey(block.Hash()), bzh); err != nil {
		panic(err)
	}
	if err = sb.db.SetSync(calcBlockHeightKey(block.Header.Height), bz); err != nil {
		panic(err)
	}
}

func calcBlockHeightKey(height int64) []byte {
	return append([]byte("block-height:"), fmt.Sprintf("%d", height)...)
}

func calcBlockHashKey(hash []byte) []byte {
	return append([]byte("block-hash:"), hash...)
}

func (sb *StoreBlock) DB() database.DB {
	return sb.db
}
