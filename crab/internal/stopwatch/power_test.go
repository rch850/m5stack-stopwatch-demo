package stopwatch

import (
	"errors"
	"testing"
	"time"
)

type fakeBus struct {
	registers      [128][256]byte
	failReads      int
	failWrites     bool
	writes         int
	failNextWrites int
	ioeNeedsConfig bool
	configReads    int
}

func (b *fakeBus) Tx(addr uint16, w, r []byte) error {
	if len(r) > 0 {
		if addr == 0x4f && w[0] == 0x23 {
			b.configReads++
		}
		if addr == 0x4f && b.ioeNeedsConfig {
			return errors.New("IOE1 not configured")
		}
		if b.failReads > 0 {
			b.failReads--
			return errors.New("not ready")
		}
		r[0] = b.registers[addr][w[0]]
		return nil
	}
	if b.failWrites {
		return errors.New("write failed")
	}
	if b.failNextWrites > 0 {
		b.failNextWrites--
		return errors.New("not ready")
	}
	if addr == 0x4f && w[0] == 0x23 && w[1] == 0 {
		b.ioeNeedsConfig = false
	}
	b.writes++
	b.registers[addr][w[0]] = w[1]
	return nil
}

func TestPowerOnPreservesOtherBitsAndRetries(t *testing.T) {
	b := fakeBus{failReads: 2}
	for _, a := range []int{0x6e, 0x4f} {
		for i := range b.registers[a] {
			b.registers[a][i] = 0xff
		}
	}
	var delays []time.Duration
	if err := PowerOn(&b, func(d time.Duration) { delays = append(delays, d) }); err != nil {
		t.Fatal(err)
	}
	for _, s := range []struct {
		addr, reg int
		want      byte
	}{{0x6e, 9, 0}, {0x6e, 10, 0}, {0x6e, 6, 0xff}, {0x4f, 0x23, 0}, {0x4f, 0x13, 0x62}, {0x4f, 3, 0xff}, {0x4f, 5, 0xff}, {0x4f, 0x14, 0xfd}, {0x4f, 4, 0xff}, {0x4f, 6, 0xfd}} {
		if got := b.registers[s.addr][s.reg]; got != s.want {
			t.Fatalf("%x:%x=%x, want %x", s.addr, s.reg, got, s.want)
		}
	}
	if len(delays) != 5 || delays[3] != 8*time.Millisecond {
		t.Fatalf("unexpected reset timing %v", delays)
	}
}

func TestPowerOnStopsOnErrors(t *testing.T) {
	for _, b := range []*fakeBus{{failReads: 100}, {failWrites: true}} {
		if err := PowerOn(b, func(time.Duration) {}); err == nil {
			t.Fatal("expected I2C error")
		}
		// Two direct PMIC writes precede the first read. A failed read must
		// never be followed by a write of an invalid read/modify/write value.
		wantWrites := 0
		if b.failReads > 0 {
			wantWrites = 2
		}
		if b.writes != wantWrites {
			t.Fatalf("must stop at the failing operation: writes=%d, want %d", b.writes, wantWrites)
		}
	}
}

func TestPowerOnConfiguresIOEBeforeReading(t *testing.T) {
	// Reproduce the observed startup failure: IOE1 rejects reads until its
	// configuration is written. Transient write NACKs must also be retried.
	b := fakeBus{ioeNeedsConfig: true, failNextWrites: 2}
	if err := PowerOn(&b, func(time.Duration) {}); err != nil {
		t.Fatal(err)
	}
	if b.ioeNeedsConfig || b.configReads != 0 {
		t.Fatal("IOE1 configuration must be written directly, without a read")
	}
	if b.registers[0x4f][5]&0x99 != 0x99 {
		t.Fatal("display power/reset sequence did not finish")
	}
}
