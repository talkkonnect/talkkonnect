package talkkonnect

import (
	"errors"
	"fmt"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	"github.com/talkkonnect/volume-go"
	"log"
	"os/exec"
	"time"
)

var stream *gumbleffmpeg.Stream

func (b *Talkkonnect) PlayIntoStream(filepath string, vol float32) {

	if stream != nil && stream.State() == gumbleffmpeg.StatePlaying {
		log.Println(fmt.Sprintf("info: Streaming Stopped", filepath))
		stream.Stop()
		b.LEDOff(b.TransmitLED)
		return
	}

	if ChimesSoundEnabled {
		if stream != nil && stream.State() == gumbleffmpeg.StatePlaying {
			time.Sleep(100 * time.Millisecond)
			return
		}

		b.LEDOn(b.TransmitLED)
		stream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile(filepath), vol)

		if err := stream.Play(); err != nil {
			log.Println(fmt.Sprintf("alert: Can't play %s error %s", filepath, err))
		} else {
			log.Println(fmt.Sprintf("info: File %s Playing!", filepath))
			stream.Wait()
			stream.Stop()
			b.LEDOff(b.TransmitLED)
		}
	} else {
		log.Println(fmt.Sprintf("alert: Sound Disabled by Config"))
	}
	return
}

func (b *Talkkonnect) RogerBeep(filepath string, vol float32) error {
	if RogerBeepSoundEnabled {
		if stream != nil && stream.State() == gumbleffmpeg.StatePlaying {
			time.Sleep(100 * time.Millisecond)
			return nil
		}
		stream = gumbleffmpeg.New(b.Client, gumbleffmpeg.SourceFile(filepath), vol)
		if err := stream.Play(); err != nil {
			return errors.New(fmt.Sprintf("alert: Can't Play Roger beep File %s error %s", filepath, err))
		} else {
			log.Println("info: Roger Beep File " + filepath + " Playing!")
		}
	} else {
		log.Println(fmt.Sprintf("alert: Roger Beep Sound Disabled by Config"))
	}
	return nil
}

func PlayWavLocal(filepath string, playbackvolume int) error {
	origVolume, _ = volume.GetVolume(OutputDevice)
	cmd := exec.Command("/usr/bin/aplay", filepath)
	err := volume.SetVolume(playbackvolume, OutputDevice)
	if err != nil {
		return errors.New(fmt.Sprintf("alert: set volume failed: %+v", err))
	}
	_, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New(fmt.Sprintf("alert: cmd.Run() for aplay failed with %s\n", err))
	}
	err = volume.SetVolume(origVolume, OutputDevice)
	if err != nil {
		return errors.New(fmt.Sprintf("alert: set volume failed: %+v", err))
	}
	return nil
}
