// I²C support.

package rpi

import (
	"fmt"
	"os"
	"sync"
	"log"
	"io/ioutil"

	"github.com/zlowred/embd"
)

type w1Bus struct {
	l           byte
	busMap      map[string]embd.W1Device
	Mu          sync.Mutex

	initialized bool
}

type w1Device struct {
	file        *os.File
	addr        string
	initialized bool
	bus         *w1Bus
}

func NewW1Bus(l byte) embd.W1Bus {
	return &w1Bus{l: l, busMap: make(map[string]embd.W1Device)}
}

func (b *w1Bus) init() error {
	if b.initialized {
		return nil
	}

	var err error
	if _, err = os.Stat("/sys/bus/w1"); os.IsNotExist(err) {
		return err
	}

	log.Printf("onewire: bus %v initialized\n", b.l)

	b.initialized = true

	return nil
}

func (d *w1Device) File() *os.File {
	return d.file
}

func (d *w1Device) OpenFile() error {
	if err := d.init(); err != nil {
		return err
	}
	var err error
	d.file, err = os.OpenFile(fmt.Sprintf("/sys/bus/w1/devices/%s/rw", d.addr), os.O_RDWR | os.O_SYNC, os.ModeDevice | os.ModeExclusive)
	return err
}

func (d *w1Device) CloseFile() error {
	if err := d.init(); err != nil {
		return err
	}
	return d.file.Close()
}

func (d *w1Device) init() error {
	if d.initialized {
		return nil
	}
	var err error
	if d.file, err = os.OpenFile(fmt.Sprintf("/sys/bus/w1/devices/%s/rw", d.addr), os.O_RDWR | os.O_SYNC, os.ModeDevice | os.ModeExclusive); err != nil {
		return err
	}
	defer d.file.Close()

	log.Printf("onewire: device %s initialized\n", d.addr)

	d.initialized = true

	return nil
}

func (d *w1Device) ReadByte() (byte, error) {
	d.bus.Mu.Lock()
	defer d.bus.Mu.Unlock()

	if err := d.init(); err != nil {
		return 0, err
	}

	bytes := make([]byte, 1)
	n, _ := d.file.Read(bytes)

	if n != 1 {
		return 0, fmt.Errorf("onewire: Unexpected number (%v) of bytes read in ReadByte", n)
	}

	return bytes[0], nil
}

func (d *w1Device) WriteByte(value byte) error {
	d.bus.Mu.Lock()
	defer d.bus.Mu.Unlock()

	if err := d.init(); err != nil {
		return err
	}

	n, err := d.file.Write([]byte{value})

	if n != 1 {
		err = fmt.Errorf("onewire: Unexpected number (%v) of bytes written in WriteByte", n)
	}

	return err
}

func (d *w1Device) WriteBytes(value []byte) error {
	d.bus.Mu.Lock()
	defer d.bus.Mu.Unlock()

	if err := d.init(); err != nil {
		return err
	}

	for i := range value {
		n, err := d.file.Write([]byte{value[i]})

		if n != 1 {
			return fmt.Errorf("onewire: Unexpected number (%v) of bytes written in WriteBytes", n)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *w1Device) ReadBytes(number int) (value []byte, err error) {
	d.bus.Mu.Lock()
	defer d.bus.Mu.Unlock()

	if err := d.init(); err != nil {
		return nil, err
	}

	bytes := make([]byte, number)
	n, _ := d.file.Read(bytes)

	if n != number {
		return nil, fmt.Errorf("onewire: Unexpected number (%v) of bytes read in ReadBytes", n)
	}

	return bytes, nil
}

func (b *w1Bus) ListDevices() (devices []string, err error) {
	dir, err := ioutil.ReadDir("/sys/bus/w1/devices/")
	if err != nil {
		return nil, err
	}
	devs := make([]string, len(dir))

	for index, element := range dir {
		devs[index] = element.Name()
	}

	return devs, nil
}

func (b *w1Bus) Open(address string) (device embd.W1Device, err error) {
	b.Mu.Lock()
	defer b.Mu.Unlock()

	if d, ok := b.busMap[address]; ok {
		return d, nil
	}

	d := &w1Device{addr: address, bus: b}
	b.busMap[address] = d
	return d, nil
}

func (b *w1Bus) Close() error {
	log.Println("Closing w1 bus")
	b.Mu.Lock()
	defer b.Mu.Unlock()

	for _, b := range b.busMap {
		b.Close()
	}

	b.busMap = make(map[string]embd.W1Device)
	return nil
}

func (d *w1Device) Close() error {
	log.Printf("Closing w1 device [%v]\n", d.addr)
	if !d.initialized {
		log.Println("not initialized")
	}

	return nil
	return d.file.Close()
}
