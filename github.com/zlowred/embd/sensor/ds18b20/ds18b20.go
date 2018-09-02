// Support for the very popular DS18B20 1-WIre temperature sensor
package ds18b20

import (
	"errors"
	"sync"
	"fmt"

	"github.com/zlowred/embd"
)

type DS18B20_Resolution int

const (
	Resolution_9bit DS18B20_Resolution = iota
	Resolution_10bit
	Resolution_11bit
	Resolution_12bit
)

// DS18B20 represents a DS18B20 temperature sensor.
type DS18B20 struct {
	Device embd.W1Device

	Raw    int16
	mu     sync.Mutex
}

// New returns a handle to a DS18B20 sensor.
func New(device embd.W1Device) *DS18B20 {
	return &DS18B20{Device: device}
}

func ds_crc_int(crc byte, data byte) byte {
	var i byte
	crc = crc ^ data
	for i = 0; i < 8; i++ {
		if crc & 0x01 != 0 {
			crc = (crc >> 1) ^ 0x8C
		} else {
			crc >>= 1
		}
	}
	return crc
}

func ds_crc(data []byte) byte {
	var crc byte = 0
	for _, x := range data {
		crc = ds_crc_int(crc, x)
	}
	return crc
}

func (sensor *DS18B20) measure() error {
	if err := sensor.Device.OpenFile(); err != nil {
		return err
	}
	defer sensor.Device.CloseFile()

	err := sensor.Device.WriteByte(0x44)

	return err
}

func (sensor *DS18B20) wait() error {
	err := sensor.Device.OpenFile()
	if err != nil {
		return err
	}
	defer sensor.Device.CloseFile()

	buf := make([]byte, 1)

	var res byte = 0
	for res != 255 && err == nil {
		var n int
		n, err = sensor.Device.File().Read(buf)
		if n > 0 {
			res = buf[0]
		}
	}

	return err
}

func (sensor *DS18B20) read() error {
	err := sensor.Device.OpenFile()
	if err != nil {
		return err
	}
	defer sensor.Device.CloseFile()

	n, err := sensor.Device.File().Write([]byte{0xBE})
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New(fmt.Sprintf("Wrong number of bytes [%v] written", n))
	}

	buf := make([]byte, 9)
	n, err = sensor.Device.File().Read(buf)

	if err != nil {
		return err
	}
	if n != 9 {
		return errors.New(fmt.Sprintf("Wrong number of bytes [%v] read", n))
	}

	crc := ds_crc(buf[:8])
	if crc != buf[8] {
		return errors.New(fmt.Sprintf("Wrong CRC [%v] while expected [%v]", buf[8], crc))
	}

	sensor.Raw = int16(buf[1]) * 256 + int16(buf[0])
	cfg := buf[4] & 0x60

	switch cfg {
	case 0x00:
		sensor.Raw &^= 7
	case 0x20:
		sensor.Raw &^= 3
	case 0x40:
		sensor.Raw &^= 1
	}

	return nil
}

func (sensor *DS18B20) ReadTemperature() error {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	err := sensor.measure()
	if err != nil {
		return err
	}
	err = sensor.wait()
	if err != nil {
		return err
	}
	err = sensor.read()
	if err != nil {
		return err
	}
	return nil
}

func (sensor *DS18B20) Celsius() float32 {
	return float32(sensor.Raw) * 0.0625
}

func (sensor *DS18B20) Fahrenheit() float32 {
	return float32(sensor.Raw) * 0.1125 + 32.
}

func (sensor *DS18B20) SetResolution(resolution DS18B20_Resolution) error {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()
	err := sensor.Device.OpenFile()
	if err != nil {
		return err
	}
	defer sensor.Device.CloseFile()

	packet := []byte{0x4E, 0, 0, 0}

	switch resolution {
	case Resolution_9bit:
		packet[3] = 0x1F
	case Resolution_10bit:
		packet[3] = 0x3F
	case Resolution_11bit:
		packet[3] = 0x5F
	case Resolution_12bit:
		packet[3] = 0x7F
	}
	err = sensor.Device.WriteBytes(packet)
	if err != nil {
		return err
	}
	err = sensor.Device.WriteByte(0x48)
	if err != nil {
		return err
	}
	return nil
}

