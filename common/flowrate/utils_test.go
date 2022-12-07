package flowrate

import "testing"

func TestPercent_String(t *testing.T) {
	p := percentOf(3, 3)
	t.Log(p)
}
