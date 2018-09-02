// Package ads1115 allows interfacing with the ads1115 5-channel, 16-bit ADC through I²C protocol.
package ads1115

import (
	"github.com/zlowred/embd"
)

const (
	defaultConfig uint16 = 0x4203
//	defaultConfig uint16 = 0xC303
	conversionRegister byte = 0
	configRegister byte = 1
	loThreshRegister byte = 2
	hoThreshRegister byte = 3
	timeout = 1000
)
// ADS1115 represents a ADS1115 16bit ADC.
type ADS1115 struct {
	Addr byte

	Bus  embd.I2CBus
}

// New creates a representation of the ads1115 converter
func New(bus embd.I2CBus, addr byte) *ADS1115 {
	dev := &ADS1115{Bus: bus, Addr: addr}
	dev.Bus.WriteWordToReg(dev.Addr, configRegister, defaultConfig)
	return dev
}

// AnalogValueAt returns the analog value at the given channel of the converter.
func (d *ADS1115) Read() (res uint16, err error) {
//	if err := d.Bus.WriteWordToReg(d.Addr, configRegister, defaultConfig); err != nil {
//		return 0, nil
//	}
//
//	var ready bool = false
//
//	var timer int64 = time.Now().Unix()
//	for ;!ready; {
//		if res, err := d.Bus.ReadWordFromReg(d.Addr, configRegister); err != nil {
//			return 0, err
//		} else {
//			ready = (res & (1 << 15)) != 0
//		}
//		if time.Now().Unix() - timer > timeout {
//			return 0, errors.New("timeout waiting for ADS1115 to complete conversion")
//		}
//	}

	if res, err := d.Bus.ReadWordFromReg(d.Addr, conversionRegister); err != nil {
		return 0, err
	} else {
		return res, nil
	}
}
