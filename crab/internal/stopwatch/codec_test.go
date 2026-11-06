package stopwatch

import (
	"errors"
	"strings"
	"testing"
	"time"
)

type codecBus struct {
	fakeBus
	started   bool
	armed     int
	failCodec bool
}

func (b *codecBus) Tx(addr uint16, w, r []byte) error {
	if b.failCodec && addr == 0x18 {
		return errors.New("codec absent")
	}
	if b.started && addr == 0x18 && len(w) == 2 && w[0] == 0x0d && w[1] == 1 {
		b.armed++
	}
	return b.fakeBus.Tx(addr, w, r)
}

func TestCodecPowerAndClockSequence(t *testing.T) {
	b := &codecBus{}
	// Preserve display power and unrelated GPIO levels while enabling audio.
	b.registers[0x4f][5] = 0x99
	b.registers[0x4f][6] = 0xa3
	b.registers[0x4f][0x13] = 0xff
	b.registers[0x4f][0x14] = 0xff
	clockWait := time.Duration(0)
	err := InitCodec(b, func(d time.Duration) {
		if b.started {
			clockWait += d
		}
	}, func() error {
		if b.registers[0x4f][5] != 0x9d || b.registers[0x4f][6] != 0xa1 {
			t.Fatal("audio power/PA configuration clobbered other bits")
		}
		if b.registers[0x4f][0x13] != 0xfb || b.registers[0x4f][0x14] != 0xfd {
			t.Fatal("pins must be push-pull")
		}
		if b.registers[0x18][0x0a] != 0x0c || b.registers[0x18][1] != 0xba {
			t.Fatal("codec must output 16-bit I2S using BCLK")
		}
		b.started = true
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if b.armed != 3 || clockWait != 90*time.Millisecond {
		t.Fatalf("missing post-clock capture arming: %d, %s", b.armed, clockWait)
	}
}

func TestCodecFailsBeforeStartingI2S(t *testing.T) {
	b := &codecBus{failCodec: true}
	starts := 0
	err := InitCodec(b, func(time.Duration) {}, func() error { starts++; return nil })
	if err == nil || !strings.Contains(err.Error(), "addr=0x18") || starts != 0 {
		t.Fatalf("must report codec failure before starting DMA: %v, starts=%d", err, starts)
	}
	b = &codecBus{}
	err = InitCodec(b, func(time.Duration) {}, func() error { return errors.New("unsupported") })
	if err == nil || b.armed != 0 {
		t.Fatal("must not arm capture after I2S failure")
	}
}
