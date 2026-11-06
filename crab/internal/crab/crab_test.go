package crab

import (
	"bytes"
	"testing"
)

func TestDebounceAndRelease(t *testing.T) {
	var b Button
	for _, s := range []struct {
		now       int64
		raw, want bool
	}{{0, false, false}, {10, true, false}, {14, false, false}, {17, true, false}, {31, true, false}, {32, true, true}, {40, false, true}, {45, true, true}, {50, false, true}, {65, false, false}} {
		if got := b.Update(s.raw, s.now); got != s.want {
			t.Fatalf("at %d: %v, want %v", s.now, got, s.want)
		}
	}
}

func TestIndependentClawsAndDirtyRegions(t *testing.T) {
	base, left, right := make([]byte, Size*2), make([]byte, Size*2), make([]byte, Size*2)
	changedLeft, changedRight := false, false
	for y := 0; y < Size; y++ {
		RenderRow(base, 0, y, 0, 0)
		RenderRow(left, 0, y, 4, 0)
		RenderRow(right, 0, y, 0, 4)
		for x := 0; x < Size; x++ {
			for side, line := range [][]byte{left, right} {
				if bytes.Equal(base[x*2:x*2+2], line[x*2:x*2+2]) {
					continue
				}
				x0 := 58
				hint := 172
				if side == 1 {
					x0 = 288
					hint = 278
				}
				if !((x >= x0 && x < x0+120 && y >= 102 && y < 222) || (x >= hint && x < hint+16 && y >= 370 && y < 386)) {
					t.Fatalf("side %d changed outside dirty regions: %d,%d", side, x, y)
				}
				if side == 0 {
					changedLeft = true
				} else {
					changedRight = true
				}
			}
		}
	}
	if !changedLeft || !changedRight {
		t.Fatal("both claws must visibly change")
	}
	for _, pressed := range []bool{true, false} {
		v := 0
		for i := 0; i < 20; i++ {
			v = Approach(v, pressed)
			if v < 0 || v > MaxClosure {
				t.Fatal("closure overflow")
			}
		}
	}
}
