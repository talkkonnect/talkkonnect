package openal

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidName      = errors.New("openal: invalid name")
	ErrInvalidEnum      = errors.New("openal: invalid enum")
	ErrInvalidValue     = errors.New("openal: invalid value")
	ErrInvalidOperation = errors.New("openal: invalid operation")

	ErrInvalidContext = errors.New("openal: invalid context")
	ErrInvalidDevice  = errors.New("openal: invalid device")
	ErrOutOfMemory    = errors.New("openal: out of memory")
)

type ErrorCode uint32

func (e ErrorCode) Error() string {
	return fmt.Sprintf("openal: error code %x", uint32(e))
}

// Err() returns the most recent error generated
// in the AL state machine.
func Err() error {
	switch code := getError(); code {
	case 0x0000:
		return nil
	case 0xA001:
		return ErrInvalidName
	case 0xA002:
		return ErrInvalidEnum
	case 0xA003:
		return ErrInvalidValue
	case 0xA004:
		return ErrInvalidOperation
	default:
		return ErrorCode(code)
	}
}
