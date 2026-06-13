// Copyright 2009 Peter H. Froehlich. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Convenience functions in pure Go.
//
// Not all convenience functions are here: those that need
// to call C code have to be in core.go instead due to cgo
// limitations, while those that are methods have to be in
// core.go due to language limitations. They should all be
// here of course, at least conceptually.

package openal

import "strings"

// Convenience Interface.
type Vector [3]float32

var tempSlice = make([]float32, 6)

const (
	x = iota
	y
	z
)

// Convenience function, see GetInteger().
func GetDistanceModel() int32 {
	return getInteger(alDistanceModel)
}

// Convenience function, see GetFloat().
func GetDopplerFactor() float32 {
	return getFloat(alDopplerFactor)
}

// Convenience function, see GetFloat().
func GetDopplerVelocity() float32 {
	return getFloat(alDopplerVelocity)
}

// Convenience function, see GetFloat().
func GetSpeedOfSound() float32 {
	return getFloat(alSpeedOfSound)
}

// Convenience function, see GetString().
func GetVendor() string {
	return GetString(alVendor)
}

// Convenience function, see GetString().
func GetVersion() string {
	return GetString(alVersion)
}

// Convenience function, see GetString().
func GetRenderer() string {
	return GetString(alRenderer)
}

// Convenience function, see GetString().
func GetExtensions() string {
	return GetString(alExtensions)
}

func GetExtensionsSlice() []string {
	return strings.Split(GetExtensions(), " ")
}

func IsExtensionPresent(ext string) bool {
	return strings.Index(GetExtensions(), ext) >= 0
}
