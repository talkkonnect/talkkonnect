// Generic OneWire driver.

package embd

import "sync"

type w1BusFactory func(byte) W1Bus

type w1Driver struct {
	busMap     map[byte]W1Bus
	busMapLock sync.Mutex

	ibf w1BusFactory
}

// NewW1Driver returns a W1Driver interface which allows control
// over the OneWire subsystem.
func NewW1Driver(ibf w1BusFactory) W1Driver {
	return &w1Driver{
		busMap: make(map[byte]W1Bus),
		ibf:    ibf,
	}
}

func (i *w1Driver) Bus(l byte) W1Bus {
	i.busMapLock.Lock()
	defer i.busMapLock.Unlock()

	if b, ok := i.busMap[l]; ok {
		return b
	}

	b := i.ibf(l)
	i.busMap[l] = b
	return b
}

func (i *w1Driver) Close() error {
	for _, b := range i.busMap {
		b.Close()
	}

	return nil
}
