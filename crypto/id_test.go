package crypto

import "testing"

func TestID_Bytes(t *testing.T) {
	var id ID = 826
	t.Log(id.ToBytes())
}
