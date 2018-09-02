//OneWire support.

package embd

import (
	"log"
	"os"
)
// W1Bus interface is used to interact with the OneWire bus.
type W1Bus interface {

	// List devices on the bus
	ListDevices() (devices []string, err error)

	// Open a device
	Open(address string) (device W1Device, err error)

	// Close releases the resources associated with the bus.
	Close() error
}

// W1Device interface is user to interact with the OneWire device.
type W1Device interface {
	// Get file
	File() *os.File
	// Open file
	OpenFile() error
	// Close file
	CloseFile() error
	// ReadByte reads a byte from the device.
	ReadByte() (value byte, err error)
	// ReadByte number of bytes from the device.
	ReadBytes(number int) (value []byte, err error)
	// WriteByte writes a byte to the device.
	WriteByte(value byte) error
	// WriteBytes writes a slice bytes to the device.
	WriteBytes(value []byte) error

	// Close releases the resources associated with the device.
	Close() error
}

// W1Driver interface interacts with the host descriptors to allow us
// control of OneWire communication.
type W1Driver interface {
	Bus(l byte) W1Bus

	// Close releases the resources associated with the driver.
	Close() error
}

var w1DriverInitialized bool
var w1DriverInstance W1Driver

// InitW1 initializes the W1 driver.
func InitW1() error {
	if w1DriverInitialized {
		return nil
	}

	desc, err := DescribeHost()
	if err != nil {
		return err
	}

	if desc.W1Driver == nil {
		return ErrFeatureNotSupported
	}

	w1DriverInstance = desc.W1Driver()
	w1DriverInitialized = true

	return nil
}

// CloseW1 releases resources associated with the OneWire driver.
func CloseW1() error {
	log.Println("Closing w1 driver")
	return w1DriverInstance.Close()
}

// NewW1Bus returns a W1Bus.
func NewW1Bus(l byte) W1Bus {
	if err := InitW1(); err != nil {
		panic(err)
	}

	return w1DriverInstance.Bus(l)
}
