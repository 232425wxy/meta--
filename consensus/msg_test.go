package consensus

import (
	"fmt"
	"github.com/232425wxy/meta--/proto/pbtypes"
	"github.com/232425wxy/meta--/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncodeDecodeMsg(t *testing.T) {
	// next view
	nv := &types.NextView{
		Type:   pbtypes.NextViewType,
		ID:     "test",
		Height: 12,
	}
	bz := MustEncode(nv)
	res := MustDecode(bz)
	switch m := res.(type) {
	case *types.NextView:
		assert.Equal(t, m.Type, nv.Type)
		assert.Equal(t, m.ID, nv.ID)
		assert.Equal(t, m.Height, nv.Height)
	default:
		panic(fmt.Sprintf("unknown message type: %T", m))
	}
}
