package stopwatch

import (
	"fmt"
	"time"
)

type I2C interface {
	Tx(uint16, []byte, []byte) error
}

// PowerOn follows M5GFX's StopWatch PM1/IOE1 sequence. Register operations
// preserve unrelated bits; IOE1 pin numbering is one-based (PYG5 = bit 4).
func PowerOn(bus I2C, sleep func(time.Duration)) error {
	type op struct {
		addr            uint16
		reg, clear, set byte
		delay           time.Duration
	}
	ops := []op{
		{0x6e, 0x09, 0xff, 0, 0}, {0x6e, 0x0a, 0xff, 0, 0}, {0x6e, 0x06, 0, 0x17, 0},
		// IOE1 can start up later than the PMIC; reads below retry for 500 ms.
		{0x4f, 0x23, 0xff, 0, 0}, {0x4f, 0x13, 0x9d, 0, 0},
		{0x4f, 0x03, 0, 0x9d, 0}, {0x4f, 0x05, 0, 0x99, 10 * time.Millisecond},
		{0x4f, 0x05, 0x18, 0, 8 * time.Millisecond}, {0x4f, 0x05, 0, 0x18, 2 * time.Millisecond},
		{0x4f, 0x14, 0x02, 0, 0}, {0x4f, 0x04, 0, 0x02, 0}, {0x4f, 0x06, 0x02, 0, 0},
	}
	for _, o := range ops {
		var value [1]byte
		var err error
		for retry := 0; retry < 50; retry++ {
			err = bus.Tx(o.addr, []byte{o.reg}, value[:])
			if err == nil {
				break
			}
			sleep(10 * time.Millisecond)
		}
		if err != nil {
			return fmt.Errorf("power read addr=0x%02x reg=0x%02x: %w", o.addr, o.reg, err)
		}
		value[0] = (value[0] &^ o.clear) | o.set
		if err = bus.Tx(o.addr, []byte{o.reg, value[0]}, nil); err != nil {
			return fmt.Errorf("power write addr=0x%02x reg=0x%02x: %w", o.addr, o.reg, err)
		}
		if o.delay != 0 {
			sleep(o.delay)
		}
	}
	return nil
}
