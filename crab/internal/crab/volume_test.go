package crab

import (
	"bytes"
	"testing"
)

func signal(amplitude, dc int) []uint16 {
	samples := make([]uint16, 256)
	for i := range samples {
		s := dc + amplitude
		if i%2 == 0 {
			s = dc - amplitude
		}
		samples[i] = uint16(int16(s))
	}
	return samples
}

func TestVolumeNoiseDCAndLoudness(t *testing.T) {
	for _, s := range [][]uint16{signal(0, 0), signal(0, 12000), signal(NoiseGate-1, 0)} {
		var v Volume
		for i := 0; i < 20; i++ {
			if got := v.Update(s); got != 0 {
				t.Fatalf("silence/DC/noise opened mouth: %d", got)
			}
		}
	}
	last := 0
	for _, amplitude := range []int{300, 700, 1500, FullVolume, 32767} {
		var v Volume
		got := 0
		for i := 0; i < 20; i++ {
			got = v.Update(signal(amplitude, 0))
		}
		if got < last || got > MaxMouth {
			t.Fatalf("non-monotonic volume %d -> %d after %d", amplitude, got, last)
		}
		last = got
	}
	if last != MaxMouth {
		t.Fatal("full-scale input must fully open mouth")
	}
	var v Volume
	if got := v.Update(signal(FullVolume, 0)); got < MaxMouth-1 {
		t.Fatalf("attack too slow: %d", got)
	}
	if got := v.Update(signal(0, 0)); got == 0 {
		t.Fatal("release should be gradual")
	}
	for i := 0; i < 40; i++ {
		v.Update(signal(0, 0))
	}
	if got := v.Update(signal(0, 0)); got != 0 {
		t.Fatalf("mouth did not close: %d", got)
	}
	// -32768 must be widened before squaring (or taking its absolute value).
	if got := v.Update([]uint16{0x8000, 0x7fff}); got != MaxMouth {
		t.Fatalf("signed PCM overflow: %d", got)
	}
}

func TestMouthDirtyRegion(t *testing.T) {
	base, actual := make([]byte, Size*2), make([]byte, Size*2)
	for mouth := 1; mouth <= MaxMouth; mouth++ {
		changed := false
		for y := 0; y < Size; y++ {
			RenderRow(base, 0, y, 0, 0, 0)
			RenderRow(actual, 0, y, 0, 0, mouth)
			for x := 0; x < Size; x++ {
				if bytes.Equal(base[x*2:x*2+2], actual[x*2:x*2+2]) {
					continue
				}
				changed = true
				if x < MouthX || x >= MouthX+MouthWidth || y < MouthY || y >= MouthY+MouthHeight {
					t.Fatalf("mouth %d changed outside region at %d,%d", mouth, x, y)
				}
			}
			// The scanline slice used for partial refresh must exactly match a full frame.
			partial := make([]byte, MouthWidth*2)
			RenderRow(partial, MouthX, y, 0, 0, mouth)
			if !bytes.Equal(partial, actual[MouthX*2:(MouthX+MouthWidth)*2]) {
				t.Fatal("partial rendering mismatch")
			}
		}
		if !changed {
			t.Fatalf("mouth %d not visible", mouth)
		}
	}
}
