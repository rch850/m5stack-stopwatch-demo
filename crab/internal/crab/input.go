package crab

// Button debounces both edges using elapsed milliseconds rather than frame count.
type Button struct {
	Pressed   bool
	candidate bool
	since     int64
}

func (b *Button) Update(pressed bool, now int64) bool {
	if pressed != b.candidate {
		b.candidate, b.since = pressed, now
	}
	if now-b.since >= 15 {
		b.Pressed = b.candidate
	}
	return b.Pressed
}

const MaxClosure = 4

func Approach(closure int, pressed bool) int {
	if pressed && closure < MaxClosure {
		return closure + 1
	}
	if !pressed && closure > 0 {
		return closure - 1
	}
	return closure
}
