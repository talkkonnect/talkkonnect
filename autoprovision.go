/*
 * talkkonnect headless mumble client/gateway with lcd screen and channel control
 * Copyright (C) 2018-2019, Suvir Kumar <suvir@talkkonnect.com>
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * Software distributed under the License is distributed on an "AS IS" basis,
 * WITHOUT WARRANTY OF ANY KIND, either express or implied. See the License
 * for the specific language governing rights and limitations under the
 * License.
 *
 * talkkonnect is the based on talkiepi and barnard by Daniel Chote and Tim Cooper
 *
 * The Initial Developer of the Original Code is
 * Suvir Kumar <suvir@talkkonnect.com>
 * Portions created by the Initial Developer are Copyright (C) Suvir Kumar. All Rights Reserved.
 *
 * Contributor(s):
 *
 * Suvir Kumar <suvir@talkkonnect.com>
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * autoprovision.go -- Autoprovisioning function for talkkonnect
 *
 */

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

	out, err := os.Create(SaveFilePath + SaveFileName)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot Create File Error: ", err))
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot Copy File Error: ", err))
	}

	return nil
}
