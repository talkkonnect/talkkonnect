/*
MIT License

Copyright (c) 2017 Tibor Hegedüs

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

Modified for talkkonnect by Suvir Kumar <suvir@talkkonnect.com>
*/

package talkkonnect

import (
	"log"
	"time"
)

func (b *Talkkonnect) Speak(text string, destination string, playBackVolume int, duration float32, loop int, language string) {
	generatedHashName := generateHashName(text)
	fileNameWithPath := Config.Global.Software.TTSMessages.TTSSoundDirectory + "/" + generatedHashName + ".mp3"

	createFolderIfNotExists(Config.Global.Software.TTSMessages.TTSSoundDirectory)
	downloadIfNotExists(fileNameWithPath, text, language)

	log.Printf("info: %v, destination=%v playBackVolume=%v duration=%v loop=%v language=%v\n", text, destination, playBackVolume, duration, loop, language)

	if destination == "local" {
		log.Println("debug: Playing TTS Media Locally")
		localMediaPlayer(fileNameWithPath, playBackVolume, true, duration, loop)
	}

	if destination == "intostream" {
		log.Println("debug: Playing TTS Media Into Stream")
		b.BackLightTimer()

		if b.IsTransmitting {
			log.Println("alert: talkkonnect was already transmitting will now stop transmitting and start the stream")
			b.TransmitStop(false)
		}

		IsPlayStream = true
		NowStreaming = IsPlayStream

		log.Println("info: Playing Recieved Text Message Into Stream as ", fileNameWithPath)
		if Config.Global.Software.TTSMessages.TTSTone.ToneEnabled && FileExists(Config.Global.Software.TTSMessages.TTSTone.ToneFile) {
			b.playIntoStream(Config.Global.Software.TTSMessages.TTSTone.ToneFile, float32(Config.Global.Software.TTSMessages.TTSTone.ToneVolume))
		}
		b.playIntoStream(fileNameWithPath, Config.Global.Software.TTSMessages.PlayVolumeIntoStream)
		IsPlayStream = false
		NowStreaming = IsPlayStream
	}

}

func (b *Talkkonnect) TTSPlayerMessage(ttsMessage string, ttsLocalPlay bool, ttsPlayIntoStream bool) {

	if ttsLocalPlay {
		if Config.Global.Software.TTSMessages.GPIO.Enabled {
			GPIOOutPin(Config.Global.Software.TTSMessages.GPIO.Name, "on")
		}
		if Config.Global.Software.TTSMessages.PreDelay.Value.Seconds() > 0 {
			time.Sleep(time.Duration(Config.Global.Software.TTSMessages.PreDelay.Value.Seconds()))
		}
		b.Speak(ttsMessage, "local", Config.Global.Software.TTS.Volumelevel, 0, 1, Config.Global.Software.TTSMessages.TTSLanguage)
		if Config.Global.Software.TTSMessages.PostDelay.Value.Seconds() > 0 {
			time.Sleep(time.Duration(Config.Global.Software.TTSMessages.PostDelay.Value.Seconds()))
		}
		if Config.Global.Software.TTSMessages.GPIO.Enabled {
			GPIOOutPin(Config.Global.Software.TTSMessages.GPIO.Name, "off")
		}
	}

	if ttsPlayIntoStream {
		b.Speak(ttsMessage, "intostream", Config.Global.Software.TTSMessages.SpeakVolumeIntoStream, 0, 1, Config.Global.Software.TTSMessages.TTSLanguage)
	}
}

func (b *Talkkonnect) TTSPlayerAPI(ttsMessage string, ttsLocalPlay bool, ttsPlayIntoStream bool, gpioEnabled bool, gpioName string, preDelay time.Duration, postDelay time.Duration, TTSLanguage string) {

	if ttsLocalPlay {
		if gpioEnabled {
			GPIOOutPin(gpioName, "on")
		}
		if preDelay > 0 {
			time.Sleep(preDelay)
		}
		b.Speak(ttsMessage, "local", Config.Global.Software.TTSMessages.SpeakVolumeIntoStream, 0, 1, TTSLanguage)
		if postDelay > 0 {
			time.Sleep(postDelay)
		}
		if gpioEnabled {
			GPIOOutPin(gpioName, "off")
		}
	}

	if ttsPlayIntoStream {
		b.Speak(ttsMessage, "intostream", Config.Global.Software.TTSMessages.SpeakVolumeIntoStream, 0, 1, TTSLanguage)
	}
}

func TTSEvent(name string) {
	if !Config.Global.Software.TTS.Enabled {
		return
	}

	for _, tts := range Config.Global.Software.TTS.Sound {

		if tts.Action == name {
			if tts.Enabled {
				localMediaPlayer(tts.File, Config.Global.Software.TTS.Volumelevel, tts.Blocking, 0, 1)
				return
			}
		}
	}
}
