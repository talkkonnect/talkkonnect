package openal_test

import (
	"github.com/timshannon/go-openal/openal"
	"testing"
)

func TestGetVendor(t *testing.T) {
	device := openal.OpenDevice("")
	defer device.CloseDevice()

	context := device.CreateContext()
	defer context.Destroy()
	context.Activate()

	vendor := openal.GetVendor()

	if err := openal.Err(); err != nil {
		t.Fatal(err)
	} else if vendor == "" {
		t.Fatal("empty vendor returned")
	}
}
