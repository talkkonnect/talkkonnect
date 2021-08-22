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
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

func (b *Talkkonnect) Speak(text string, destination string) {
	Folder := "audio"
	generatedHashName := generateHashName(text)
	fileName := Folder + "/" + generatedHashName + ".mp3"

	createFolderIfNotExists(Folder)
	downloadIfNotExists(fileName, text)

	if destination == "local" {
		localmediaplayer(fileName)
	}

	if destination == "intostream" {
		b.BackLightTimer()

		if b.IsTransmitting {
			log.Println("alert: talkkonnect was already transmitting will now stop transmitting and start the stream")
			b.TransmitStop(false)
		}

		IsPlayStream = true
		NowStreaming = IsPlayStream

		log.Println("info: Playing Recieved Text Message Into Stream as ", fileName)
		b.playIntoStream(fileName, StreamSoundVolume)
		IsPlayStream = false
		NowStreaming = IsPlayStream

	}

}

func createFolderIfNotExists(folder string) {
	dir, err := os.Open(folder)
	if os.IsNotExist(err) {
		os.MkdirAll(folder, 0700)
		return
	}

	dir.Close()
}

func downloadIfNotExists(fileName string, text string) {
	f, err := os.Open(fileName)
	if err != nil {
		url := fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&total=1&idx=0&textlen=32&client=tw-ob&q=%s&tl=%s", url.QueryEscape(text), "en")
		response, err := http.Get(url)
		if err != nil {
			return
		}
		defer response.Body.Close()

		output, err := os.Create(fileName)
		if err != nil {
			return
		}

		_, _ = io.Copy(output, response.Body)
	}

	f.Close()
}

func generateHashName(name string) string {
	hash := md5.Sum([]byte(name))
	return hex.EncodeToString(hash[:])
}

func localmediaplayer(fileName string) {
	localplayer := exec.Command("ffplay", "-autoexit", fileName)
	localplayer.Run()
}

func (b *Talkkonnect) TTSPlayer(ttsMessage string, ttsLocalPlay bool, ttsLocalPlayRXLed bool, ttlPlayIntoStream bool) {

	if ttsLocalPlay {
		if ttsLocalPlayRXLed {
			LEDOnFunc(VoiceActivityLED)
		}
		b.Speak(ttsMessage, "local")
		if ttsLocalPlayRXLed {
			LEDOffFunc(VoiceActivityLED)
		}
	}

	if ttlPlayIntoStream {
		b.Speak(ttsMessage, "intostream")
	}
}
