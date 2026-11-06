//go:build tinygo && esp32s3

package stopwatch

import (
	"machine"
	"time"
)

type Microphone struct{}

func (m *Microphone) Init() error {
	err := InitCodec(machine.I2C0, time.Sleep, func() error {
		return machine.I2S0.Configure(machine.I2SConfig{
			SCK: machine.GPIO17, WS: machine.GPIO15, SDI: machine.GPIO16, SDO: machine.NoPin,
			Mode: machine.I2SModeReceiver, Standard: machine.I2StandardPhilips,
			ClockSource: machine.I2SClockSourceInternal, DataFormat: machine.I2SDataFormat16bit,
			AudioFrequency: 16000,
		})
	})
	if err != nil {
		machine.I2S0.Enable(false)
	}
	return err
}

func (m *Microphone) Read(samples []uint16) (int, error) { return machine.I2S0.ReadMono(samples) }
