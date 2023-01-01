package state

import (
	"errors"
	"fmt"
	"github.com/232425wxy/meta--/database"
	"github.com/232425wxy/meta--/proto/pbstate"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"github.com/232425wxy/meta--/types"
	"github.com/cosmos/gogoproto/proto"
)

type Store struct {
	db database.DB
}

func NewStore(db database.DB) *Store {
	return &Store{db: db}
}

func (s *Store) LoadFromDBOrGenesis(genesis *types.Genesis) *State {
	state, err := s.LoadState()
	if err != nil {
		return &State{}
	}
	if state.IsEmpty() {
		state = MakeGenesisState(genesis)
	}
	return state
}

func (s *Store) LoadState() (*State, error) {
	bz, err := s.db.Get(StateKey)
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

func (s *Store) SaveState(state *State) error {
	return s.db.SetSync(StateKey, state.ToBytes())
}

func (s *Store) Bootstrap(state *State) error {
	return s.SaveState(state)
}

func (s *Store) SaveValidators(height int64, validators *types.ValidatorSet) error {
	pb := validators.ToProto()
	bz, err := proto.Marshal(pb)
	if err != nil {
		return err
	}
	return s.db.SetSync(calcValidatorsKey(height), bz)
}

func (s *Store) LoadValidators(height int64) (*types.ValidatorSet, error) {
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
