// Package npa700 allows interfacing with GE NPA-700 pressure sensor. This sensor
// has the ability to provide compensated temperature and pressure readings.
package npa700

import (
	"github.com/zlowred/embd"
	"sync"
)

// NPA700 represents a Bosch BMP180 barometric sensor.
type NPA700 struct {
	Bus  embd.I2CBus

	RawTemperature int16
	RawPressure int16
	mu sync.Mutex
	addr byte
}

// New returns a handle to a BMP180 sensor.
func New(bus embd.I2CBus, addr byte) *NPA700 {
	return &NPA700{Bus: bus, addr: addr}
}

func (sensor *NPA700) Read() error {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	data, err := sensor.Bus.ReadBytes(sensor.addr, 4)

	if err != nil {
		return err
	}

	sensor.RawPressure = (int16(data[0]) << 8) + int16(data[1])
	sensor.RawTemperature = (int16(data[2]) << 3) + (int16(data[3]) >> 5)

	return nil
}

func (sensor *NPA700) Celsius() float32 {
	return float32(sensor.RawTemperature) * 200. / 2048. - 50.
}

func (sensor *NPA700) Fahrenheit() float32 {
	return (float32(sensor.RawTemperature) * 200. / 2048. - 50.) * 1.8 + 32.
}

func (sensor *NPA700) Pascals(offset float32, minValue float32, maxValue float32, minPressure float32, maxPressure float32) float32 {
	return minPressure + (float32(sensor.RawPressure) + offset - minValue) / (maxValue - minValue) * (maxPressure - minPressure)
}
