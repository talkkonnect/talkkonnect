// Forked by Tim Shannon 2012
// Copyright 2009 Peter H. Froehlich. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Go binding for OpenAL's "al" API.
//
// See http://connect.creativelabs.com/openal/Documentation/OpenAL%201.1%20Specification.htm
// for details about OpenAL not described here.
//
// OpenAL types are (in principle) mapped to Go types as
// follows:
//
//	ALboolean	bool	(al.h says char, but Go's bool should be compatible)
//	ALchar		uint8	(although al.h suggests int8, Go's uint8 (aka byte) seems better)
//	ALbyte		int8	(al.h says char, implying that char is signed)
//	ALubyte		uint8	(al.h says unsigned char)
//	ALshort		int16
//	ALushort	uint16
//	ALint		int32
//	ALuint		uint32
//	ALsizei		int32	(although that's strange, it's what OpenAL wants)
//	ALenum		int32	(although that's strange, it's what OpenAL wants)
//	ALfloat		float32
//	ALdouble	float64
//	ALvoid		not applicable (but see below)
//
// We also stick to these (not mentioned explicitly in
// OpenAL):
//
//	ALvoid*		unsafe.Pointer (but never exported)
//	ALchar*		string
//
// Finally, in places where OpenAL expects pointers to
// C-style arrays, we use Go slices if appropriate:
//
//	ALboolean*	[]bool
//	ALvoid*		[]byte (see Buffer.SetData() for example)
//	ALint*		[]int32
//	ALuint*		[]uint32 []Source []Buffer
//	ALfloat*	[]float32
//	ALdouble*	[]float64
//
// Overall, the correspondence of types hopefully feels
// natural enough. Note that many of these types do not
// actually occur in the API.
//
// The names of OpenAL constants follow the established
// Go conventions: instead of AL_FORMAT_MONO16 we use
// FormatMono16 for example.
//
// Conversion to Go's camel case notation does however
// lead to name clashes between constants and functions.
// For example, AL_DISTANCE_MODEL becomes DistanceModel
// which collides with the OpenAL function of the same
// name used to set the current distance model. We have
// to rename either the constant or the function, and
// since the function name seems to be at fault (it's a
// setter but doesn't make that obvious), we rename the
// function.
//
// In fact, we renamed plenty of functions, not just the
// ones where collisions with constants were the driving
// force. For example, instead of the Sourcef/GetSourcef
// abomination, we use Getf/Setf methods on a Source type.
// Everything should still be easily recognizable for
// OpenAL hackers, but this structure is a lot more
// sensible (and reveals that the OpenAL API is actually
// not such a bad design).
//
// There are a few cases where constants would collide
// with the names of types we introduced here. Since the
// types serve a much more important function, we renamed
// the constants in those cases. For example AL_BUFFER
// would collide with the type Buffer so it's name is now
// Buffer_ instead. Not pretty, but in many cases you
// don't need the constants anyway as the functionality
// they represent is probably available through one of
// the convenience functions we introduced as well. For
// example consider the task of attaching a buffer to a
// source. In C, you'd say alSourcei(sid, AL_BUFFER, bid).
// In Go, you can say sid.Seti(Buffer_, bid) as well, but
// you probably want to say sid.SetBuffer(bid) instead.
//
// TODO: Decide on the final API design; the current state
// has only specialized methods, none of the generic ones
// anymore; it exposes everything (except stuff we can't
// do) but I am not sure whether this is the right API for
// the level we operate on. Not yet anyway. Anyone?
package openal

/*
#cgo linux LDFLAGS: -lopenal
#cgo windows LDFLAGS: -lopenal32
#cgo darwin LDFLAGS: -framework OpenAL
#include <stdlib.h>
#include "local.h"
#include "wrapper.h"
*/
import "C"
import "unsafe"

// General purpose constants. None can be used with SetDistanceModel()
// to disable distance attenuation. None can be used with Source.SetBuffer()
// to clear a Source of buffers.
const (
	None    = 0
	alFalse = 0
	alTrue  = 1
)

// GetInteger() queries.
const (
	alDistanceModel = 0xD000
)

// GetFloat() queries.
const (
	alDopplerFactor   = 0xC000
	alDopplerVelocity = 0xC001
	alSpeedOfSound    = 0xC003
)

// GetString() queries.
const (
	alVendor     = 0xB001
	alVersion    = 0xB002
	alRenderer   = 0xB003
	alExtensions = 0xB004
)

// Shared Source/Listener properties.
const (
	AlPosition = 0x1004
	AlVelocity = 0x1006
	AlGain     = 0x100A
)

func GetString(param int32) string {
	return C.GoString(C.walGetString(C.ALenum(param)))
}

func getBoolean(param int32) bool {
	return C.alGetBoolean(C.ALenum(param)) != alFalse
}

func getInteger(param int32) int32 {
	return int32(C.alGetInteger(C.ALenum(param)))
}

func getFloat(param int32) float32 {
	return float32(C.alGetFloat(C.ALenum(param)))
}

func getDouble(param int32) float64 {
	return float64(C.alGetDouble(C.ALenum(param)))
}

// Renamed, was GetBooleanv.
func getBooleans(param int32, data []bool) {
	C.walGetBooleanv(C.ALenum(param), unsafe.Pointer(&data[0]))
}

// Renamed, was GetIntegerv.
func getIntegers(param int32, data []int32) {
	C.walGetIntegerv(C.ALenum(param), unsafe.Pointer(&data[0]))
}

// Renamed, was GetFloatv.
func getFloats(param int32, data []float32) {
	C.walGetFloatv(C.ALenum(param), unsafe.Pointer(&data[0]))
}

// Renamed, was GetDoublev.
func getDoubles(param int32, data []float64) {
	C.walGetDoublev(C.ALenum(param), unsafe.Pointer(&data[0]))
}

// GetError() returns the most recent error generated
// in the AL state machine.
func getError() uint32 {
	return uint32(C.alGetError())
}

// Renamed, was DopplerFactor.
func SetDopplerFactor(value float32) {
	C.alDopplerFactor(C.ALfloat(value))
}

// Renamed, was DopplerVelocity.
func SetDopplerVelocity(value float32) {
	C.alDopplerVelocity(C.ALfloat(value))
}

// Renamed, was SpeedOfSound.
func SetSpeedOfSound(value float32) {
	C.alSpeedOfSound(C.ALfloat(value))
}

// Distance models for SetDistanceModel() and GetDistanceModel().
const (
	InverseDistance         = 0xD001
	InverseDistanceClamped  = 0xD002
	LinearDistance          = 0xD003
	LinearDistanceClamped   = 0xD004
	ExponentDistance        = 0xD005
	ExponentDistanceClamped = 0xD006
)

// SetDistanceModel() changes the current distance model.
// Pass "None" to disable distance attenuation.
// Renamed, was DistanceModel.
func SetDistanceModel(model int32) {
	C.alDistanceModel(C.ALenum(model))
}

///// Crap ///////////////////////////////////////////////////////////

// These functions are wrapped and should work fine, but they
// have no purpose: There are *no* capabilities in OpenAL 1.1
// which is the latest specification. So we removed from from
// the API for now, it's complicated enough without them.
//
//func Enable(capability int32) {
//	C.alEnable(C.ALenum(capability));
//}
//
//func Disable(capability int32) {
//	C.alDisable(C.ALenum(capability));
//}
//
//func IsEnabled(capability int32) bool {
//	return C.alIsEnabled(C.ALenum(capability)) != alFalse;
//}

// These constants are documented as "not yet exposed". We
// keep them here in case they ever become valid. They are
// buffer states.
//
//const (
//	Unused = 0x2010;
//	Pending = 0x2011;
//	Processed = 0x2012;
//)

// These functions would work fine, but they are not very
// useful since we have distinct Source and Buffer types.
// Leaving them out reduces API complexity, a good thing.
//
//func IsSource(id uint32) bool {
//	return C.alIsSource(C.ALuint(id)) != alFalse;
//}
//
//func IsBuffer(id uint32) bool {
//	return C.alIsBuffer(C.ALuint(id)) != alFalse;
//}
