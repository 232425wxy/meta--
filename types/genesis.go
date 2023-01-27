package types

import (
	mos "github.com/232425wxy/meta--/common/os"
	mjson "github.com/232425wxy/meta--/json"
	"time"
)

type Genesis struct {
	GenesisTime   time.Time    `json:"genesis_time"`
	InitialHeight int64        `json:"initial_height"`
	Validators    []*Validator `json:"validators"`
}

func (gen *Genesis) SaveAs(file string) error {
	bz, err := mjson.EncodeIndent(gen, "", "	")
	if err != nil {
		return err
	}
	return mos.WriteFile(file, bz, 0644)
}

func GenesisReadFromFile(file string) (*Genesis, error) {
	gen := &Genesis{}
	bz, err := mos.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err = mjson.Decode(bz, gen); err != nil {
		return nil, err
	}
	return gen, nil
}
