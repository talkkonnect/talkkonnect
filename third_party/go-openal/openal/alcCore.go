// Forked by Tim Shannon 2012
// Copyright 2009 Peter H. Froehlich. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// C-level binding for OpenAL's "alc" API.
//
// Please consider using the Go-level binding instead.
//
// Note that "alc" introduces the exact same types as "al"
// but with different names. Check the documentation of
// openal/al for more information about the mapping to
// Go types.
//
// XXX: Sadly we have to returns pointers for both Device
// and Context to avoid problems with implicit assignments
// in clients. It's sad because it makes the overhead a
// lot higher, each of those calls triggers an allocation.
package openal

//#cgo linux LDFLAGS: -lopenal
//#cgo darwin LDFLAGS: -framework OpenAL
//#include <stdlib.h>
//#include "local.h"
/*
ALCdevice *walcOpenDevice(const char *devicename) {
	if (devicename != NULL && devicename[0] == '\0') {
		return alcOpenDevice(NULL);
	}
	return alcOpenDevice(devicename);
}
const ALCchar *walcGetString(ALCdevice *device, ALCenum param) {
	return alcGetString(device, param);
}
ALCboolean walcIsExtensionPresent(ALCdevice *device, const ALCchar *extname) {
	return alcIsExtensionPresent(device, extname);
}
void walcGetIntegerv(ALCdevice *device, ALCenum param, ALCsizei size, void *data) {
	alcGetIntegerv(device, param, size, data);
}
ALCdevice *walcCaptureOpenDevice(const char *devicename, ALCuint frequency, ALCenum format, ALCsizei buffersize) {
	if (devicename != NULL && devicename[0] == '\0') {
		return alcCaptureOpenDevice(NULL, frequency, format, buffersize);
	}
	return alcCaptureOpenDevice(devicename, frequency, format, buffersize);
}
ALCint walcGetInteger(ALCdevice *device, ALCenum param) {
	ALCint result;
	alcGetIntegerv(device, param, 1, &result);
	return result;
}
*/
import "C"
import "unsafe"

const (
	Frequency     = 0x1007 // int Hz
	Refresh       = 0x1008 // int Hz
	Sync          = 0x1009 // bool
	MonoSources   = 0x1010 // int
	StereoSources = 0x1011 // int
)

// The Specifier string for default device?
const (
	DefaultDeviceSpecifier = 0x1004
	DeviceSpecifier        = 0x1005
	Extensions             = 0x1006
)

// ?
const (
	MajorVersion = 0x1000
	MinorVersion = 0x1001
)

// ?
const (
	AttributesSize = 0x1002
	AllAttributes  = 0x1003
)

// Capture extension
const (
	CaptureDeviceSpecifier        = 0x310
	CaptureDefaultDeviceSpecifier = 0x311
	CaptureSamples                = 0x312
)

const ExtensionEnumerateAll = "ALC_ENUMERATE_ALL_EXT"

const (
	AllDevicesSpecifier        = 0x1013
	DefaultAllDevicesSpecifier = 0x1014
)

type Device struct {
	handle *C.ALCdevice
}

func (self *Device) getError() uint32 {
	return uint32(C.alcGetError(self.handle))
}

// Err() returns the most recent error generated
// in the AL state machine.
func (self *Device) Err() error {
	switch code := self.getError(); code {
	case 0x0000:
		return nil
	case 0xA001:
		return ErrInvalidDevice
	case 0xA002:
		return ErrInvalidContext
	case 0xA003:
		return ErrInvalidEnum
	case 0xA004:
		return ErrInvalidValue
	case 0xA005:
		return ErrOutOfMemory
	default:
		return ErrorCode(code)
	}
}

func OpenDevice(name string) *Device {
	p := cStringOrEmpty(name)
	h := C.walcOpenDevice(p)
	freeCString(p, name)
	return &Device{h}
}

// OpenDeviceChecked opens a playback device and validates the handle.
func OpenDeviceChecked(name string) (*Device, error) {
	device := OpenDevice(name)
	if err := device.Valid(); err != nil {
		return nil, err
	}
	return device, nil
}

// GetDeviceString returns an OpenAL device property string.
// Pass device=nil to query global enumerations.
func GetDeviceString(device *Device, param uint32) string {
	var handle *C.ALCdevice
	if device != nil {
		handle = device.handle
	}
	s := C.walcGetString(handle, C.ALCenum(param))
	if s == nil {
		return ""
	}
	return C.GoString(s)
}

// IsALCExtensionPresent reports whether an OpenAL ALC extension is available.
func IsALCExtensionPresent(ext string) bool {
	p := C.CString(ext)
	defer C.free(unsafe.Pointer(p))
	return C.walcIsExtensionPresent(nil, p) != 0
}

func (self *Device) Valid() error {
	if self == nil || self.handle == nil {
		return ErrInvalidDevice
	}
	return self.Err()
}

func (self *Device) CloseDevice() bool {
	//TODO: really a method? or not?
	return C.alcCloseDevice(self.handle) != 0
}

func (self *Device) CreateContext() *Context {
	// TODO: really a method?
	// TODO: attrlist support
	return &Context{C.alcCreateContext(self.handle, nil)}
}

func (self *Device) GetIntegerv(param uint32, size uint32) (result []int32) {
	result = make([]int32, size)
	C.walcGetIntegerv(self.handle, C.ALCenum(param), C.ALCsizei(size), unsafe.Pointer(&result[0]))
	return
}

func (self *Device) GetInteger(param uint32) int32 {
	return int32(C.walcGetInteger(self.handle, C.ALCenum(param)))
}

type CaptureDevice struct {
	Device
	sampleSize uint32
}

func CaptureOpenDevice(name string, freq uint32, format Format, size uint32) *CaptureDevice {
	p := cStringOrEmpty(name)
	h := C.walcCaptureOpenDevice(p, C.ALCuint(freq), C.ALCenum(format), C.ALCsizei(size))
	freeCString(p, name)
	return &CaptureDevice{Device{h}, uint32(format.SampleSize())}
}

// CaptureOpenDeviceChecked opens a capture device and validates the handle.
func CaptureOpenDeviceChecked(name string, freq uint32, format Format, size uint32) (*CaptureDevice, error) {
	device := CaptureOpenDevice(name, freq, format, size)
	if err := device.Valid(); err != nil {
		return nil, err
	}
	return device, nil
}

func cStringOrEmpty(name string) *C.char {
	if name == "" {
		return (*C.char)(unsafe.Pointer(nil))
	}
	return C.CString(name)
}

func freeCString(p *C.char, name string) {
	if name != "" {
		C.free(unsafe.Pointer(p))
	}
}

// XXX: Override Device.CloseDevice to make sure the correct
// C function is called even if someone decides to use this
// behind an interface.
func (self *CaptureDevice) CloseDevice() bool {
	return C.alcCaptureCloseDevice(self.handle) != 0
}

func (self *CaptureDevice) CaptureCloseDevice() bool {
	return self.CloseDevice()
}

func (self *CaptureDevice) CaptureStart() {
	C.alcCaptureStart(self.handle)
}

func (self *CaptureDevice) CaptureStop() {
	C.alcCaptureStop(self.handle)
}

func (self *CaptureDevice) CaptureTo(data []byte) {
	C.alcCaptureSamples(self.handle, unsafe.Pointer(&data[0]), C.ALCsizei(uint32(len(data))/self.sampleSize))
}

func (self *CaptureDevice) CaptureToInt16(data []int16) {
	C.alcCaptureSamples(self.handle, unsafe.Pointer(&data[0]), C.ALCsizei(uint32(len(data))*2/self.sampleSize))
}

func (self *CaptureDevice) CaptureMono8To(data []byte) {
	self.CaptureTo(data)
}

func (self *CaptureDevice) CaptureMono16To(data []int16) {
	self.CaptureToInt16(data)
}

func (self *CaptureDevice) CaptureStereo8To(data [][2]byte) {
	C.alcCaptureSamples(self.handle, unsafe.Pointer(&data[0]), C.ALCsizei(uint32(len(data))*2/self.sampleSize))
}

func (self *CaptureDevice) CaptureStereo16To(data [][2]int16) {
	C.alcCaptureSamples(self.handle, unsafe.Pointer(&data[0]), C.ALCsizei(uint32(len(data))*4/self.sampleSize))
}

func (self *CaptureDevice) CaptureSamples(size uint32) (data []byte) {
	data = make([]byte, size*self.sampleSize)
	self.CaptureTo(data)
	return
}

func (self *CaptureDevice) CaptureSamplesInt16(size uint32) (data []int16) {
	data = make([]int16, size*self.sampleSize/2)
	self.CaptureToInt16(data)
	return
}

func (self *CaptureDevice) CapturedSamples() (size uint32) {
	return uint32(self.GetInteger(CaptureSamples))
}

///// Context ///////////////////////////////////////////////////////

// Context encapsulates the state of a given instance
// of the OpenAL state machine. Only one context can
// be active in a given process.
type Context struct {
	handle *C.ALCcontext
}

// A context that doesn't exist, useful for certain
// context operations (see OpenAL documentation for
// details).
var NullContext Context

// Renamed, was MakeContextCurrent.
func (self *Context) Activate() bool {
	return C.alcMakeContextCurrent(self.handle) != alFalse
}

// Renamed, was ProcessContext.
func (self *Context) Process() {
	C.alcProcessContext(self.handle)
}

// Renamed, was SuspendContext.
func (self *Context) Suspend() {
	C.alcSuspendContext(self.handle)
}

// Renamed, was DestroyContext.
func (self *Context) Destroy() {
	C.alcDestroyContext(self.handle)
	self.handle = nil
}

// Renamed, was GetContextsDevice.
func (self *Context) GetDevice() *Device {
	return &Device{C.alcGetContextsDevice(self.handle)}
}

// Renamed, was GetCurrentContext.
func CurrentContext() *Context {
	return &Context{C.alcGetCurrentContext()}
}
