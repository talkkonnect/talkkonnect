// Package pca9955b allows interfacing with the pca9955b 16-channel, 8-bit PWM Controller through I2C protocol.
package pca9955b

import (
	"sync"

	"github.com/zlowred/embd"
)

type AutoIncrementMode byte
type Register byte
type PinMode byte

const (
	AutoIncrement_00_43 AutoIncrementMode = 0
	AutoIncrement_08_17 = 1
	AutoIncrement_00_27 = 2
	AutoIncrement_06_17 = 3
)

const (
	MODE1 Register = 0x00
	MODE2 = 0x01
	LEDOUT0 = 0x02
	LEDOUT1 = 0x03
	LEDOUT2 = 0x04
	LEDOUT3 = 0x05
	GRPPWM = 0x06
	GRPFREQ = 0x07
	PWM0 = 0x08
	PWM1 = 0x09
	PWM2 = 0x0a
	PWM3 = 0x0b
	PWM4 = 0x0c
	PWM5 = 0x0d
	PWM6 = 0x0e
	PWM7 = 0x0f
	PWM8 = 0x10
	PWM9 = 0x11
	PWM10 = 0x12
	PWM11 = 0x13
	PWM12 = 0x14
	PWM13 = 0x15
	PWM14 = 0x16
	PWM15 = 0x17
	IREF0 = 0x18
	IREF1 = 0x19
	IREF2 = 0x1a
	IREF3 = 0x1b
	IREF4 = 0x1c
	IREF5 = 0x1d
	IREF6 = 0x1e
	IREF7 = 0x1f
	IREF8 = 0x20
	IREF9 = 0x21
	IREF10 = 0x22
	IREF11 = 0x23
	IREF12 = 0x24
	IREF13 = 0x25
	IREF14 = 0x26
	IREF15 = 0x27
)

const (
	OFF PinMode = 0x0
	ON = 0x01
	PWM = 0x02
	PWMGRP = 0x03
)

// PCA9955B represents a PCA9955B PWM generator.
type PCA9955B struct {
	Bus               embd.I2CBus
	Addr              byte

	initialized       bool
	mu                sync.RWMutex
	autoIncrementMode AutoIncrementMode
	data              [38]byte
}

// New creates a new PCA9955B interface.
func New(bus embd.I2CBus, addr byte) *PCA9955B {
	return &PCA9955B{
		Bus:  bus,
		Addr: addr,
	}
}

func (d *PCA9955B) init() error {
	if d.initialized {
		return nil
	}
	d.data[4] = 0xFF
	for i := 0x16; i < 0x26; i++ {
		d.data[i] = 0xFF
	}
	d.initialized = true
	return nil
}

func (d *PCA9955B) Reset() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.init(); err != nil {
		return err
	}

	if err := d.Bus.WriteByte(byte(0), byte(6)); err != nil {
		return err
	}

	return nil
}
func (d *PCA9955B) WriteToRegister(register Register, value byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.init(); err != nil {
		return err
	}

	if err := d.Bus.WriteByteToReg(d.Addr, byte(register), value); err != nil {
		return err
	}
	return nil
}

func (d *PCA9955B) ReadRegister(register Register) (val byte, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.init(); err != nil {
		return 0, err
	}

	if reg, err := d.Bus.ReadByteFromReg(d.Addr, byte(register)); err != nil {
		return 0, err
	} else {
		return reg, nil
	}
}

func (d *PCA9955B) SetOutput(pin byte, value byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.init(); err != nil {
		return err
	}

	reg := pin >> 2
	regoffset := (pin & 3) << 1
	var ledval PinMode = PWM
	if value == 0 {
		ledval = OFF
	} else if value == 255 {
		ledval = ON
	}

	regval, err := d.Bus.ReadByteFromReg(d.Addr, byte(LEDOUT0) + reg);

	if err != nil {
		return err
	}

	regval &^= 3 << regoffset
	regval |= byte(ledval) << regoffset

	if err := d.Bus.WriteByteToReg(d.Addr, byte(LEDOUT0) + reg, regval); err != nil {
		return err
	}
	if err := d.Bus.WriteByteToReg(d.Addr, byte(PWM0) + pin, value); err != nil {
		return err
	}
	if err := d.Bus.WriteByteToReg(d.Addr, byte(IREF0) + pin, 0xFF); err != nil {
		return err
	}

	return nil
}

func (d *PCA9955B) Update() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.init(); err != nil {
		return err
	}

	for idx, val := range d.data {
		if err := d.Bus.WriteByteToReg(d.Addr, byte(0x82 + idx), val); err != nil {
			return err
		}
	}

	return nil
}
