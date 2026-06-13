// Copyright 2009 Peter H. Froehlich. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package openal

/*
#include <stdlib.h>
#cgo darwin LDFLAGS: -framework OpenAL
#include "local.h"
#include "wrapper.h"
*/
import "C"
import "unsafe"

// Buffers are storage space for sample data.
type Buffer uint32

// Attributes that can be queried with Buffer.Geti().
const (
	alFrequency = 0x2001
	alBits      = 0x2002
	alChannels  = 0x2003
	alSize      = 0x2004
)

type Buffers []Buffer

// NewBuffers() creates n fresh buffers.
// Renamed, was GenBuffers.
func NewBuffers(n int) (buffers Buffers) {
	buffers = make(Buffers, n)
	C.walGenBuffers(C.ALsizei(n), unsafe.Pointer(&buffers[0]))
	return
}

// Delete() deletes the given buffers.
func (self Buffers) Delete() {
	n := len(self)
	C.walDeleteBuffers(C.ALsizei(n), unsafe.Pointer(&self[0]))
}

// Renamed, was Bufferf.
func (self Buffer) setf(param int32, value float32) {
	C.alBufferf(C.ALuint(self), C.ALenum(param), C.ALfloat(value))
}

// Renamed, was Buffer3f.
func (self Buffer) set3f(param int32, value1, value2, value3 float32) {
	C.alBuffer3f(C.ALuint(self), C.ALenum(param), C.ALfloat(value1), C.ALfloat(value2), C.ALfloat(value3))
}

// Renamed, was Bufferfv.
func (self Buffer) setfv(param int32, values []float32) {
	C.walBufferfv(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&values[0]))
}

// Renamed, was Bufferi.
func (self Buffer) seti(param int32, value int32) {
	C.alBufferi(C.ALuint(self), C.ALenum(param), C.ALint(value))
}

// Renamed, was Buffer3i.
func (self Buffer) set3i(param int32, value1, value2, value3 int32) {
	C.alBuffer3i(C.ALuint(self), C.ALenum(param), C.ALint(value1), C.ALint(value2), C.ALint(value3))
}

// Renamed, was Bufferiv.
func (self Buffer) setiv(param int32, values []int32) {
	C.walBufferiv(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&values[0]))
}

// Renamed, was GetBufferf.
func (self Buffer) getf(param int32) float32 {
	return float32(C.walGetBufferf(C.ALuint(self), C.ALenum(param)))
}

// Renamed, was GetBuffer3f.
func (self Buffer) get3f(param int32) (value1, value2, value3 float32) {
	var v1, v2, v3 float32
	C.walGetBuffer3f(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&v1),
		unsafe.Pointer(&v2), unsafe.Pointer(&v3))
	value1, value2, value3 = v1, v2, v3
	return
}

// Renamed, was GetBufferfv.
func (self Buffer) getfv(param int32, values []float32) {
	C.walGetBufferfv(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&values[0]))
	return
}

// Renamed, was GetBufferi.
func (self Buffer) geti(param int32) int32 {
	return int32(C.walGetBufferi(C.ALuint(self), C.ALenum(param)))
}

// Renamed, was GetBuffer3i.
func (self Buffer) get3i(param int32) (value1, value2, value3 int32) {
	var v1, v2, v3 int32
	C.walGetBuffer3i(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&v1),
		unsafe.Pointer(&v2), unsafe.Pointer(&v3))
	value1, value2, value3 = v1, v2, v3
	return
}

// Renamed, was GetBufferiv.
func (self Buffer) getiv(param int32, values []int32) {
	C.walGetBufferiv(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&values[0]))
}

type Format uint32

func (f Format) SampleSize() int {
	switch f {
	case FormatMono8:
		return 1
	case FormatMono16:
		return 2
	case FormatStereo8:
		return 2
	case FormatStereo16:
		return 4
	default:
		return 1
	}
}

// Format of sound samples passed to Buffer.SetData().
const (
	FormatMono8    Format = 0x1100
	FormatMono16   Format = 0x1101
	FormatStereo8  Format = 0x1102
	FormatStereo16 Format = 0x1103
)

// SetData() specifies the sample data the buffer should use.
// For FormatMono16 and FormatStereo8 the data slice must be a
// multiple of two bytes long; for FormatStereo16 the data slice
// must be a multiple of four bytes long. The frequency is given
// in Hz.
// Renamed, was BufferData.
func (self Buffer) SetData(format Format, data []byte, frequency int32) {
	C.alBufferData(C.ALuint(self), C.ALenum(format), unsafe.Pointer(&data[0]),
		C.ALsizei(len(data)), C.ALsizei(frequency))
}

func (self Buffer) SetDataInt16(format Format, data []int16, frequency int32) {
	C.alBufferData(C.ALuint(self), C.ALenum(format), unsafe.Pointer(&data[0]),
		C.ALsizei(len(data)*2), C.ALsizei(frequency))
}

func (self Buffer) SetDataMono8(data []byte, frequency int32) {
	C.alBufferData(C.ALuint(self), C.ALenum(FormatMono8), unsafe.Pointer(&data[0]),
		C.ALsizei(len(data)), C.ALsizei(frequency))
}

func (self Buffer) SetDataMono16(data []int16, frequency int32) {
	C.alBufferData(C.ALuint(self), C.ALenum(FormatMono16), unsafe.Pointer(&data[0]),
		C.ALsizei(len(data)*2), C.ALsizei(frequency))
}

func (self Buffer) SetDataStereo8(data [][2]byte, frequency int32) {
	C.alBufferData(C.ALuint(self), C.ALenum(FormatStereo8), unsafe.Pointer(&data[0]),
		C.ALsizei(len(data)*2), C.ALsizei(frequency))
}

func (self Buffer) SetDataStereo16(data [][2]int16, frequency int32) {
	C.alBufferData(C.ALuint(self), C.ALenum(FormatStereo16), unsafe.Pointer(&data[0]),
		C.ALsizei(len(data)*4), C.ALsizei(frequency))
}

// NewBuffer() creates a single buffer.
// Convenience function, see NewBuffers().
func NewBuffer() Buffer {
	return Buffer(C.walGenBuffer())
}

// Delete() deletes a single buffer.
// Convenience function, see DeleteBuffers().
func (self Buffer) Delete() {
	C.walDeleteSource(C.ALuint(self))
}

// GetFrequency() returns the frequency, in Hz, of the buffer's sample data.
// Convenience method.
func (self Buffer) GetFrequency() uint32 {
	return uint32(self.geti(alFrequency))
}

// GetBits() returns the resolution, either 8 or 16 bits, of the buffer's sample data.
// Convenience method.
func (self Buffer) GetBits() uint32 {
	return uint32(self.geti(alBits))
}

// GetChannels() returns the number of channels, either 1 or 2, of the buffer's sample data.
// Convenience method.
func (self Buffer) GetChannels() uint32 {
	return uint32(self.geti(alChannels))
}

// GetSize() returns the size, in bytes, of the buffer's sample data.
// Convenience method.
func (self Buffer) GetSize() uint32 {
	return uint32(self.geti(alSize))
}
