package stopwatch

import (
	"fmt"
	"time"
)

// InitCodec follows M5Unified's _microphone_enabled_cb_stopwatch and
// _microphone_post_start_cb_stopwatch. The ES8311 derives its clock from
// BCLK (32 Fs * 8), so TinyGo need not supply its unsupported MCLK output.
func InitCodec(bus I2C, sleep func(time.Duration), startI2S func() error) error {
	write := func(addr uint16, reg, value byte) error {
		var err error
		for i := 0; i < 4; i++ {
			err = bus.Tx(addr, []byte{reg, value}, nil)
			if err == nil {
				return nil
			}
			sleep(5 * time.Millisecond)
		}
		return fmt.Errorf("mic write addr=0x%02x reg=0x%02x: %w", addr, reg, err)
	}
	update := func(reg, clear, set byte) error {
		var data [1]byte
		var err error
		for i := 0; i < 4; i++ {
			err = bus.Tx(0x4f, []byte{reg}, data[:])
			if err == nil {
				break
			}
			sleep(5 * time.Millisecond)
		}
		if err != nil {
			return fmt.Errorf("mic power read reg=0x%02x: %w", reg, err)
		}
		return write(0x4f, reg, data[0]&^clear|set)
	}
	// GPIO3 = bit2: audio rail on. GPIO10 = high-bank bit1: amplifier off.
	for _, op := range [][3]byte{{0x14, 0x02, 0}, {0x04, 0, 0x02}, {0x06, 0x02, 0}, {0x13, 0x04, 0}, {0x03, 0, 0x04}, {0x05, 0, 0x04}} {
		if err := update(op[0], op[1], op[2]); err != nil {
			return err
		}
	}
	sleep(10 * time.Millisecond)
	for _, op := range [][2]byte{
		{0x00, 0x80}, // reset / control state machine on
		{0x01, 0xba}, // BCLK source; enable ADC clocks
		{0x02, 0x18}, // BCLK multiplier x8 (16-bit stereo slots)
		{0x0a, 0x0c}, // ADC serial output: Philips I2S, 16 bits, left slot
		{0x0d, 0x01}, // analog power up (also re-written after clocks start)
		{0x0e, 0x02}, // enable PGA and ADC modulator
		{0x12, 0x02}, // DAC off
		{0x14, 0x17}, // differential analog microphone, +21 dB PGA
		{0x17, 0xcb}, // +6 dB ADC volume
		{0x1c, 0x64}, // EQ bypass / dynamic high-pass filter
	} {
		if err := write(0x18, op[0], op[1]); err != nil {
			return err
		}
	}
	if err := startI2S(); err != nil {
		return fmt.Errorf("mic I2S: %w", err)
	}
	// ES8311 capture must be armed while the clocks run, including on cold boot.
	for i := 0; i < 3; i++ {
		sleep(30 * time.Millisecond)
		if err := write(0x18, 0x0d, 0x01); err != nil {
			return err
		}
	}
	return nil
}
