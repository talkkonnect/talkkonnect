package openal_test

import (
	"fmt"
	"github.com/timshannon/go-openal/openal"
	"io/ioutil"
	"time"
)

func ExamplePlay() {
	device := openal.OpenDevice("")
	defer device.CloseDevice()

	context := device.CreateContext()
	defer context.Destroy()
	context.Activate()

	source := openal.NewSource()
	defer source.Pause()
	source.SetLooping(false)

	buffer := openal.NewBuffer()

	if err := openal.Err(); err != nil {
		fmt.Println(err)
		return
	}

	data, err := ioutil.ReadFile("data/welcome.wav")
	if err != nil {
		fmt.Println(err)
		return
	}

	buffer.SetData(openal.FormatMono16, data, 44100)

	source.SetBuffer(buffer)
	source.Play()
	for source.State() == openal.Playing {
		// loop long enough to let the wave file finish
		time.Sleep(time.Millisecond * 100)
	}
	source.Delete()
	fmt.Println("sound played")
	// Output: sound played
}

func ExampleMonitor() {
	const (
		frequency    = 44100
		format       = openal.FormatStereo16
		captureSize  = 512
		buffersCount = 10
	)
	mic := openal.CaptureOpenDevice("", frequency, format, frequency*2)
	mic.CaptureStart()
	defer mic.CloseDevice()

	device := openal.OpenDevice("")
	defer device.CloseDevice()

	context := device.CreateContext()
	context.Activate()
	defer context.Destroy()

	source := openal.NewSource()
	source.SetLooping(false)
	defer source.Stop()

	buffers := openal.NewBuffers(buffersCount)
	samples := make([]byte, captureSize*format.SampleSize())

	start := time.Now()
	for time.Since(start) < time.Second { // play for 1 second
		if err := openal.Err(); err != nil {
			fmt.Println("error:", err)
			return
		}
		// Get any free buffers
		if prcessed := source.BuffersProcessed(); prcessed > 0 {
			buffersNew := make(openal.Buffers, prcessed)
			source.UnqueueBuffers(buffersNew)
			buffers = append(buffers, buffersNew...)
		}
		if len(buffers) == 0 {
			continue
		}

		if mic.CapturedSamples() >= captureSize {
			mic.CaptureTo(samples)
			buffer := buffers[len(buffers)-1]
			buffers = buffers[:len(buffers)-1]
			buffer.SetData(format, samples, frequency)
			source.QueueBuffer(buffer)

			// If we have enough buffers, start playing
			if source.State() != openal.Playing {
				if source.BuffersQueued() > 2 {
					source.Play()
				}
			}
		}
	}
	fmt.Println(source.State())
	// Output: Playing
}
