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
import (
	"fmt"
	"unsafe"
)

type State int32

func (s State) String() string {
	switch s {
	case Initial:
		return "Initial"
	case Playing:
		return "Playing"
	case Paused:
		return "Paused"
	case Stopped:
		return "Stopped"
	default:
		return fmt.Sprintf("%x", int32(s))
	}
}

// Results from Source.State() query.
const (
	Initial State = 0x1011
	Playing State = 0x1012
	Paused  State = 0x1013
	Stopped State = 0x1014
)

// Results from Source.Type() query.
const (
	Static       = 0x1028
	Streaming    = 0x1029
	Undetermined = 0x1030
)

// TODO: Source properties.
// Regardless of what your al.h header may claim, Pitch
// only applies to Sources, not to Listeners. And I got
// that from Chris Robinson himself.
const (
	AlSourceRelative    = 0x202
	AlConeInnerAngle    = 0x1001
	AlConeOuterAngle    = 0x1002
	AlPitch             = 0x1003
	AlDirection         = 0x1005
	AlLooping           = 0x1007
	AlBuffer            = 0x1009
	AlMinGain           = 0x100D
	AlMaxGain           = 0x100E
	AlReferenceDistance = 0x1020
	AlRolloffFactor     = 0x1021
	AlConeOuterGain     = 0x1022
	AlMaxDistance       = 0x1023
	AlSecOffset         = 0x1024
	AlSampleOffset      = 0x1025
	AlByteOffset        = 0x1026
)

// Sources represent sound emitters in 3d space.
type Source uint32

type Sources []Source

// NewSources() creates n sources.
// Renamed, was GenSources.
func NewSources(n int) (sources Sources) {
	sources = make(Sources, n)
	C.walGenSources(C.ALsizei(n), unsafe.Pointer(&sources[0]))
	return
}

// Delete deletes the sources.
func (self Sources) Delete() {
	n := len(self)
	C.walDeleteSources(C.ALsizei(n), unsafe.Pointer(&self[0]))
}

// Renamed, was SourcePlayv.
func (self Sources) Play() {
	C.walSourcePlayv(C.ALsizei(len(self)), unsafe.Pointer(&self[0]))
}

// Renamed, was SourceStopv.
func (self Sources) Stop() {
	C.walSourceStopv(C.ALsizei(len(self)), unsafe.Pointer(&self[0]))
}

// Renamed, was SourceRewindv.
func (self Sources) Rewind() {
	C.walSourceRewindv(C.ALsizei(len(self)), unsafe.Pointer(&self[0]))
}

// Renamed, was SourcePausev.
func (self Sources) Pause() {
	C.walSourcePausev(C.ALsizei(len(self)), unsafe.Pointer(&self[0]))
}

// Renamed, was Sourcef.
func (self Source) Setf(param int32, value float32) {
	C.alSourcef(C.ALuint(self), C.ALenum(param), C.ALfloat(value))
}

// Renamed, was Source3f.
func (self Source) Set3f(param int32, value1, value2, value3 float32) {
	C.alSource3f(C.ALuint(self), C.ALenum(param), C.ALfloat(value1), C.ALfloat(value2), C.ALfloat(value3))
}

// Renamed, was Sourcefv.
func (self Source) Setfv(param int32, values []float32) {
	C.walSourcefv(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&values[0]))
}

// Renamed, was Sourcei.
func (self Source) Seti(param int32, value int32) {
	C.alSourcei(C.ALuint(self), C.ALenum(param), C.ALint(value))
}

// Renamed, was Source3i.
func (self Source) Set3i(param int32, value1, value2, value3 int32) {
	C.alSource3i(C.ALuint(self), C.ALenum(param), C.ALint(value1), C.ALint(value2), C.ALint(value3))
}

// Renamed, was Sourceiv.
func (self Source) Setiv(param int32, values []int32) {
	C.walSourceiv(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&values[0]))
}

// Renamed, was GetSourcef.
func (self Source) Getf(param int32) float32 {
	return float32(C.walGetSourcef(C.ALuint(self), C.ALenum(param)))
}

// Renamed, was GetSource3f.
func (self Source) Get3f(param int32) (v1, v2, v3 float32) {
	C.walGetSource3f(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&v1),
		unsafe.Pointer(&v2), unsafe.Pointer(&v3))
	return
}

// Renamed, was GetSourcefv.
func (self Source) Getfv(param int32, values []float32) {
	C.walGetSourcefv(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&values[0]))
}

// Renamed, was GetSourcei.
func (self Source) Geti(param int32) int32 {
	return int32(C.walGetSourcei(C.ALuint(self), C.ALenum(param)))
}

// Renamed, was GetSource3i.
func (self Source) Get3i(param int32) (v1, v2, v3 int32) {
	C.walGetSource3i(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&v1),
		unsafe.Pointer(&v2), unsafe.Pointer(&v3))
	return
}

// Renamed, was GetSourceiv.
func (self Source) Getiv(param int32, values []int32) {
	C.walGetSourceiv(C.ALuint(self), C.ALenum(param), unsafe.Pointer(&values[0]))
}

// Delete deletes the source.
// Convenience function, see DeleteSources().
func (self Source) Delete() {
	C.walDeleteSource(C.ALuint(self))
}

// Renamed, was SourcePlay.
func (self Source) Play() {
	C.alSourcePlay(C.ALuint(self))
}

// Renamed, was SourceStop.
func (self Source) Stop() {
	C.alSourceStop(C.ALuint(self))
}

// Renamed, was SourceRewind.
func (self Source) Rewind() {
	C.alSourceRewind(C.ALuint(self))
}

// Renamed, was SourcePause.
func (self Source) Pause() {
	C.alSourcePause(C.ALuint(self))
}

// Renamed, was SourceQueueBuffers.
func (self Source) QueueBuffers(buffers Buffers) {
	C.walSourceQueueBuffers(C.ALuint(self), C.ALsizei(len(buffers)), unsafe.Pointer(&buffers[0]))
}

// Renamed, was SourceUnqueueBuffers.
func (self Source) UnqueueBuffers(buffers Buffers) {
	C.walSourceUnqueueBuffers(C.ALuint(self), C.ALsizei(len(buffers)), unsafe.Pointer(&buffers[0]))
}

///// Convenience ////////////////////////////////////////////////////

// NewSource() creates a single source.
// Convenience function, see NewSources().
func NewSource() Source {
	return Source(C.walGenSource())
}

// Convenience method, see Source.QueueBuffers().
func (self Source) QueueBuffer(buffer Buffer) {
	C.walSourceQueueBuffer(C.ALuint(self), C.ALuint(buffer))
}

// Convenience method, see Source.QueueBuffers().
func (self Source) UnqueueBuffer() Buffer {
	return Buffer(C.walSourceUnqueueBuffer(C.ALuint(self)))
}

// Source queries.
// TODO: SourceType isn't documented as a query in the
// al.h header, but it is documented that way in
// the OpenAL 1.1 specification.
const (
	AlSourceState      = 0x1010
	AlBuffersQueued    = 0x1015
	AlBuffersProcessed = 0x1016
	AlSourceType       = 0x1027
)

// Convenience method, see Source.Geti().
func (self Source) BuffersQueued() int32 {
	return self.Geti(AlBuffersQueued)
}

// Convenience method, see Source.Geti().
func (self Source) BuffersProcessed() int32 {
	return self.Geti(AlBuffersProcessed)
}

// Convenience method, see Source.Geti().
func (self Source) State() State {
	return State(self.Geti(AlSourceState))
}

// Convenience method, see Source.Geti().
func (self Source) Type() int32 {
	return self.Geti(AlSourceType)
}

// Convenience method, see Source.Getf().
func (self Source) GetGain() (gain float32) {
	return self.Getf(AlGain)
}

// Convenience method, see Source.Setf().
func (self Source) SetGain(gain float32) {
	self.Setf(AlGain, gain)
}

// Convenience method, see Source.Getf().
func (self Source) GetMinGain() (gain float32) {
	return self.Getf(AlMinGain)
}

// Convenience method, see Source.Setf().
func (self Source) SetMinGain(gain float32) {
	self.Setf(AlMinGain, gain)
}

// Convenience method, see Source.Getf().
func (self Source) GetMaxGain() (gain float32) {
	return self.Getf(AlMaxGain)
}

// Convenience method, see Source.Setf().
func (self Source) SetMaxGain(gain float32) {
	self.Setf(AlMaxGain, gain)
}

// Convenience method, see Source.Getf().
func (self Source) GetReferenceDistance() (distance float32) {
	return self.Getf(AlReferenceDistance)
}

// Convenience method, see Source.Setf().
func (self Source) SetReferenceDistance(distance float32) {
	self.Setf(AlReferenceDistance, distance)
}

// Convenience method, see Source.Getf().
func (self Source) GetMaxDistance() (distance float32) {
	return self.Getf(AlMaxDistance)
}

// Convenience method, see Source.Setf().
func (self Source) SetMaxDistance(distance float32) {
	self.Setf(AlMaxDistance, distance)
}

// Convenience method, see Source.Getf().
func (self Source) GetPitch() float32 {
	return self.Getf(AlPitch)
}

// Convenience method, see Source.Setf().
func (self Source) SetPitch(pitch float32) {
	self.Setf(AlPitch, pitch)
}

// Convenience method, see Source.Getf().
func (self Source) GetRolloffFactor() (gain float32) {
	return self.Getf(AlRolloffFactor)
}

// Convenience method, see Source.Setf().
func (self Source) SetRolloffFactor(gain float32) {
	self.Setf(AlRolloffFactor, gain)
}

// Convenience method, see Source.Geti().
func (self Source) GetLooping() bool {
	return self.Geti(AlLooping) != alFalse
}

var bool2al map[bool]int32 = map[bool]int32{true: alTrue, false: alFalse}

// Convenience method, see Source.Seti().
func (self Source) SetLooping(yes bool) {
	self.Seti(AlLooping, bool2al[yes])
}

// Convenience method, see Source.Geti().
func (self Source) GetSourceRelative() bool {
	return self.Geti(AlSourceRelative) != alFalse
}

// Convenience method, see Source.Seti().
func (self Source) SetSourceRelative(yes bool) {
	self.Seti(AlSourceRelative, bool2al[yes])
}

// Convenience method, see Source.Setfv().
func (self Source) SetPosition(vector *Vector) {
	self.Set3f(AlPosition, vector[x], vector[y], vector[z])
}

// Convenience method, see Source.Getfv().
func (self Source) GetPosition(result *Vector) {
	result[x], result[y], result[z] = self.Get3f(AlPosition)
}

// Convenience method, see Source.Setfv().
func (self Source) SetDirection(vector *Vector) {
	self.Set3f(AlDirection, vector[x], vector[y], vector[z])
}

// Convenience method, see Source.Getfv().
func (self Source) GetDirection(result *Vector) {
	result[x], result[y], result[z] = self.Get3f(AlDirection)
}

// Convenience method, see Source.Setfv().
func (self Source) SetVelocity(vector *Vector) {
	self.Set3f(AlVelocity, vector[x], vector[y], vector[z])
}

// Convenience method, see Source.Getfv().
func (self Source) GetVelocity(result *Vector) {
	result[x], result[y], result[z] = self.Get3f(AlVelocity)
}

// Convenience method, see Source.Getf().
func (self Source) GetOffsetSeconds() float32 {
	return self.Getf(AlSecOffset)
}

// Convenience method, see Source.Setf().
func (self Source) SetOffsetSeconds(offset float32) {
	self.Setf(AlSecOffset, offset)
}

// Convenience method, see Source.Geti().
func (self Source) GetOffsetSamples() int32 {
	return self.Geti(AlSampleOffset)
}

// Convenience method, see Source.Seti().
func (self Source) SetOffsetSamples(offset int32) {
	self.Seti(AlSampleOffset, offset)
}

// Convenience method, see Source.Geti().
func (self Source) GetOffsetBytes() int32 {
	return self.Geti(AlByteOffset)
}

// Convenience method, see Source.Seti().
func (self Source) SetOffsetBytes(offset int32) {
	self.Seti(AlByteOffset, offset)
}

// Convenience method, see Source.Getf().
func (self Source) GetInnerAngle() float32 {
	return self.Getf(AlConeInnerAngle)
}

// Convenience method, see Source.Setf().
func (self Source) SetInnerAngle(offset float32) {
	self.Setf(AlConeInnerAngle, offset)
}

// Convenience method, see Source.Getf().
func (self Source) GetOuterAngle() float32 {
	return self.Getf(AlConeOuterAngle)
}

// Convenience method, see Source.Setf().
func (self Source) SetOuterAngle(offset float32) {
	self.Setf(AlConeOuterAngle, offset)
}

// Convenience method, see Source.Getf().
func (self Source) GetOuterGain() float32 {
	return self.Getf(AlConeOuterGain)
}

// Convenience method, see Source.Setf().
func (self Source) SetOuterGain(offset float32) {
	self.Setf(AlConeOuterGain, offset)
}

// Convenience method, see Source.Geti().
func (self Source) SetBuffer(buffer Buffer) {
	self.Seti(AlBuffer, int32(buffer))
}

// Convenience method, see Source.Geti().
func (self Source) GetBuffer() (buffer Buffer) {
	return Buffer(self.Geti(AlBuffer))
}
