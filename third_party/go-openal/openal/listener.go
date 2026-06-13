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

// Listener properties.
const (
	AlOrientation = 0x100F
)

// Listener represents the singleton receiver of
// sound in 3d space.
//
// We "fake" this type so we can provide OpenAL
// listener calls as methods. This is convenient
// and makes all those calls consistent with the
// way they work for Source and Buffer. You can't
// make new listeners, there's only one!
type Listener struct{}

// Renamed, was Listenerf.
func (self Listener) Setf(param int32, value float32) {
	C.alListenerf(C.ALenum(param), C.ALfloat(value))
}

// Renamed, was Listener3f.
func (self Listener) Set3f(param int32, value1, value2, value3 float32) {
	C.alListener3f(C.ALenum(param), C.ALfloat(value1), C.ALfloat(value2), C.ALfloat(value3))
}

// Renamed, was Listenerfv.
func (self Listener) Setfv(param int32, values []float32) {
	C.walListenerfv(C.ALenum(param), unsafe.Pointer(&values[0]))
}

// Renamed, was Listeneri.
func (self Listener) Seti(param int32, value int32) {
	C.alListeneri(C.ALenum(param), C.ALint(value))
}

// Renamed, was Listener3i.
func (self Listener) Set3i(param int32, value1, value2, value3 int32) {
	C.alListener3i(C.ALenum(param), C.ALint(value1), C.ALint(value2), C.ALint(value3))
}

// Renamed, was Listeneriv.
func (self Listener) Setiv(param int32, values []int32) {
	C.walListeneriv(C.ALenum(param), unsafe.Pointer(&values[0]))
}

// Renamed, was GetListenerf.
func (self Listener) Getf(param int32) float32 {
	return float32(C.walGetListenerf(C.ALenum(param)))
}

// Renamed, was GetListener3f.
func (self Listener) Get3f(param int32) (v1, v2, v3 float32) {
	C.walGetListener3f(C.ALenum(param), unsafe.Pointer(&v1),
		unsafe.Pointer(&v2), unsafe.Pointer(&v3))
	return
}

// Renamed, was GetListenerfv.
func (self Listener) Getfv(param int32, values []float32) {
	C.walGetListenerfv(C.ALenum(param), unsafe.Pointer(&values[0]))
	return
}

// Renamed, was GetListeneri.
func (self Listener) Geti(param int32) int32 {
	return int32(C.walGetListeneri(C.ALenum(param)))
}

// Renamed, was GetListener3i.
func (self Listener) Get3i(param int32) (v1, v2, v3 int32) {
	C.walGetListener3i(C.ALenum(param), unsafe.Pointer(&v1),
		unsafe.Pointer(&v2), unsafe.Pointer(&v3))
	return
}

// Renamed, was GetListeneriv.
func (self Listener) Getiv(param int32, values []int32) {
	C.walGetListeneriv(C.ALenum(param), unsafe.Pointer(&values[0]))
}

///// Convenience ////////////////////////////////////////////////////

// Convenience method, see Listener.Setf().
func (self Listener) SetGain(gain float32) {
	self.Setf(AlGain, gain)
}

// Convenience method, see Listener.Getf().
func (self Listener) GetGain() (gain float32) {
	return self.Getf(AlGain)
}

// Convenience method, see Listener.Setfv().
func (self Listener) SetPosition(vector *Vector) {
	self.Set3f(AlPosition, vector[x], vector[y], vector[z])
}

// Convenience method, see Listener.Getfv().
func (self Listener) GetPosition(result *Vector) {
	result[x], result[y], result[z] = self.Get3f(AlPosition)
}

// Convenience method, see Listener.Setfv().
func (self Listener) SetVelocity(vector *Vector) {
	self.Set3f(AlVelocity, vector[x], vector[y], vector[z])
}

// Convenience method, see Listener.Getfv().
func (self Listener) GetVelocity(result *Vector) {
	result[x], result[y], result[z] = self.Get3f(AlVelocity)
}

// Convenience method, see Listener.Setfv().
func (self Listener) SetOrientation(at *Vector, up *Vector) {
	tempSlice[0] = at[x]
	tempSlice[1] = at[y]
	tempSlice[2] = at[z]
	tempSlice[3] = up[x]
	tempSlice[4] = up[y]
	tempSlice[5] = up[z]
	self.Setfv(AlOrientation, tempSlice)
}

// Convenience method, see Listener.Getfv().
func (self Listener) GetOrientation(resultAt, resultUp *Vector) {
	self.Getfv(AlOrientation, tempSlice)
	resultAt[x] = tempSlice[0]
	resultAt[y] = tempSlice[1]
	resultAt[z] = tempSlice[2]
	resultUp[x] = tempSlice[3]
	resultUp[y] = tempSlice[4]
	resultUp[z] = tempSlice[5]
}
