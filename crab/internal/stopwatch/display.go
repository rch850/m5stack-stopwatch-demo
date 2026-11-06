//go:build tinygo && esp32s3

package stopwatch

import (
	"example.com/stopwatch-crab/internal/crab"
	"machine"
	"runtime/volatile"
	"time"
	"unsafe"
)

const (
	cs        = machine.GPIO39
	clk       = machine.GPIO40
	clockMask = uint32(1 << (40 - 32))
	dataMask  = uint32(1<<(41-32) | 1<<(42-32) | 1<<(46-32) | 1<<(45-32))
)

type Display struct {
	set, clear *volatile.Register32
	quad       [16]uint32
	line       [crab.Size * 2]byte
}

func (d *Display) Init() error {
	cs.Configure(machine.PinConfig{Mode: machine.PinOutput})
	cs.High()
	if err := machine.I2C0.Configure(machine.I2CConfig{SDA: machine.GPIO47, SCL: machine.GPIO48, Frequency: 100000}); err != nil {
		return err
	}
	if err := PowerOn(machine.I2C0, time.Sleep); err != nil {
		return err
	}
	for _, pin := range []machine.Pin{clk, machine.GPIO41, machine.GPIO42, machine.GPIO46, machine.GPIO45} {
		pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
		pin.Low()
	}
	// All QSPI pins share GPIO bank 1. Volatile stores update just these bits.
	set, _ := clk.PortMaskSet()
	clear, _ := clk.PortMaskClear()
	d.set = (*volatile.Register32)(unsafe.Pointer(set))
	d.clear = (*volatile.Register32)(unsafe.Pointer(clear))
	for n := 0; n < 16; n++ {
		for i, pin := range []uint{41, 42, 46, 45} {
			if n&(1<<i) != 0 {
				d.quad[n] |= 1 << (pin - 32)
			}
		}
	}
	d.command(0x11)
	time.Sleep(150 * time.Millisecond)
	d.command(0xc4, 0x80)
	d.command(0x35, 0x80)
	d.command(0x44, 0x01, 0xd2)
	d.command(0x53, 0x20)
	d.command(0x20)
	d.command(0x36, 0)
	d.command(0x3a, 0x55) // RGB565
	d.command(0x51, 0xa0)
	d.command(0x29)
	return nil
}

// Commands use one data lane; the 0x32 pixel payload uses four lanes.
// Software QSPI avoids depending on a non-public TinyGo QSPI API.
func (d *Display) serial(b byte) {
	for bit := uint(0); bit < 8; bit++ {
		d.clear.Set(clockMask | dataMask)
		if b&0x80 != 0 {
			d.set.Set(1 << (41 - 32))
		}
		d.set.Set(clockMask)
		b <<= 1
	}
	d.clear.Set(clockMask)
}

func (d *Display) command(cmd byte, args ...byte) {
	cs.Low()
	for _, b := range []byte{0x02, 0, cmd, 0} {
		d.serial(b)
	}
	for _, b := range args {
		d.serial(b)
	}
	cs.High()
}

// Draw updates an even-aligned rectangle. It streams one scanline at a time.
func (d *Display) Draw(x, y, w, h, left, right, mouth int) {
	// M5GFX uses a six-column controller offset for the StopWatch panel.
	xs, xe := x+6, x+w-1+6
	d.command(0x2a, byte(xs>>8), byte(xs), byte(xe>>8), byte(xe))
	d.command(0x2b, byte(y>>8), byte(y), byte((y+h-1)>>8), byte(y+h-1))
	cs.Low()
	for _, b := range []byte{0x32, 0, 0x2c, 0} {
		d.serial(b)
	}
	for yy := y; yy < y+h; yy++ {
		line := d.line[:w*2]
		crab.RenderRow(line, x, yy, left, right, mouth)
		for _, b := range line {
			d.clear.Set(clockMask | dataMask)
			d.set.Set(d.quad[b>>4])
			d.set.Set(clockMask)
			d.clear.Set(clockMask | dataMask)
			d.set.Set(d.quad[b&15])
			d.set.Set(clockMask)
		}
	}
	d.clear.Set(clockMask)
	cs.High()
}
