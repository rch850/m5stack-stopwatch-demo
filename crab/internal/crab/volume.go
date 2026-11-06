package crab

import "math"

const (
	MaxMouth    = 8
	MouthX      = 212
	MouthY      = 264
	MouthWidth  = 44
	MouthHeight = 48
	// Signed 16-bit PCM units after the codec's +21 dB PGA / +6 dB ADC gain.
	NoiseGate  = 120
	FullVolume = 4000
)

// Volume maps AC RMS to a mouth opening, with a fast attack and slow release.
// Samples are signed PCM carried in uint16, as returned by TinyGo ReadMono.
type Volume struct{ envelope float64 }

func (v *Volume) Update(samples []uint16) int {
	if len(samples) == 0 {
		return v.opening()
	}
	var sum, squares int64
	for _, raw := range samples {
		s := int64(int16(raw))
		sum += s
		squares += s * s
	}
	n := float64(len(samples))
	mean := float64(sum) / n
	variance := float64(squares)/n - mean*mean
	rms := math.Sqrt(math.Max(0, variance))
	target := math.Max(0, rms-NoiseGate)
	if target > v.envelope {
		v.envelope += (target - v.envelope) * 0.75
	} else {
		v.envelope *= 0.80
		v.envelope += target * 0.20
	}
	return v.opening()
}

func (v *Volume) opening() int {
	if v.envelope < 40 {
		return 0
	}
	level := int(math.Sqrt(v.envelope/(FullVolume-NoiseGate))*MaxMouth + 0.5)
	if level > MaxMouth {
		return MaxMouth
	}
	return level
}
