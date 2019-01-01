package talkkonnect

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func AutoProvision() error {

	if len(TkId) < 8 {
		return errors.New("TkId Configuration Provisioning XML File should be at least 8 characters!")
	}

	if string(TkId[len(TkId)-4]) != ".xml" {
		TkId = TkId + ".xml"
	}

	if string(Url[len(Url)-1]) != "/" {
		Url = Url + "/"
	}

	if string(SaveFilePath[len(SaveFilePath)-1]) != "/" {
		SaveFilePath = SaveFilePath + "/"
	}

	fileUrl := Url + TkId
	log.Println("info: Contacting Provisioning Server to Download XML Config File")
	err := DownloadFile(SaveFilePath, SaveFileName, fileUrl)

	if err != nil {
		return errors.New(fmt.Sprintf("DownloadFile Module Returned an Error: ", err))
	}

	return nil

}

func DownloadFile(SaveFilePath string, SaveFileName string, Url string) error {

	// Get the provisioning file
	resp, err := http.Get(Url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("info: HTTP Provisioning Server Responded With Status 200 OK ")
	} else {
		return errors.New(fmt.Sprintf("error: HTTP Provisioning Server Returned Status ", resp.StatusCode, " ", http.StatusText(resp.StatusCode)))

	}

	// Create the xml config file
	out, err := os.Create(SaveFilePath + SaveFileName)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot Create File Error: ", err))
	}
	defer out.Close()

	// Write the body of fetched xml file to created file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot Copy File Error: ", err))
	}

	return nil
}
