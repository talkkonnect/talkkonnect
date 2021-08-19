package talkkonnect

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
)

func Speak(text string, destination string) {
	Folder := "audio"
	generatedHashName := generateHashName(text)
	fileName := Folder + "/" + generatedHashName + ".mp3"

	createFolderIfNotExists(Folder)
	downloadIfNotExists(fileName, text)

	if destination == "local" {
		localmediaplayer(fileName)
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
