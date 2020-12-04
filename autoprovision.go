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

func autoProvision() error {

	if len(TkID) < 8 {
		return errors.New("TkID Configuration Provisioning XML File should be at least 8 characters!")
	}

	if string(TkID[len(TkID)-4]) != ".xml" {
		TkID = TkID + ".xml"
	}

	if string(URL[len(URL)-1]) != "/" {
		URL = URL + "/"
	}

	if string(SaveFilePath[len(SaveFilePath)-1]) != "/" {
		SaveFilePath = SaveFilePath + "/"
	}

	fileURL := URL + TkID
	log.Println("debug: Contacting Provisioning Server to Download XML Config File")
	err := DownloadFile(SaveFilePath, SaveFilename, fileURL)

	if err != nil {
		return errors.New(fmt.Sprintf("error: DownloadFile Module Returned an Error: ", err))
	}

	return nil

}

func DownloadFile(SaveFilePath string, SaveFilename string, URL string) error {

	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("debug: HTTP Provisioning Server Responded With Status 200 OK ")
	} else {
		return errors.New(fmt.Sprintf("error: HTTP Provisioning Server Returned Status ", resp.StatusCode, " ", http.StatusText(resp.StatusCode)))

	}

	out, err := os.Create(SaveFilePath + SaveFilename)
	if err != nil {
		return errors.New(fmt.Sprintf("error: Cannot Create File Error: ", err))
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("error: Cannot Copy File Error: ", err))
	}

	return nil
}
