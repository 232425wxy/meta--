package consensus

import (
	"github.com/232425wxy/meta--/json"
	"sync"
)

type PeerState struct {
	mu     sync.Mutex
	Height int64 `json:"height"`
	Round  int16 `json:"round"`
	Step   Step  `json:"step"`
}

func (ps *PeerState) EncodeToBytes() ([]byte, error) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return json.Encode(ps)
}
