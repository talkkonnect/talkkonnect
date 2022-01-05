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
	"regexp"
)

func autoProvision() error {
	if len(Config.Global.Software.AutoProvisioning.TkID) < 8 {
		var err error
		var macaddress []string
		macaddress, err = getMacAddr()
		if err != nil {
			return errors.New("TkID Configuration Provisioning XML Filenaame Not Found And Cannot Get Mac Address ")
		}
		for _, a := range macaddress {
			re, err := regexp.Compile(`(:)`)
			if err != nil {
				FatalCleanUp(err.Error())
			}
			Config.Global.Software.AutoProvisioning.TkID = re.ReplaceAllString(a, "")
		}
	}

	if string(Config.Global.Software.AutoProvisioning.TkID[len(Config.Global.Software.AutoProvisioning.TkID)-4]) != ".xml" {
		Config.Global.Software.AutoProvisioning.TkID = Config.Global.Software.AutoProvisioning.TkID + ".xml"
	}

	if string(Config.Global.Software.AutoProvisioning.URL[len(Config.Global.Software.AutoProvisioning.URL)-1]) != "/" {
		Config.Global.Software.AutoProvisioning.URL = Config.Global.Software.AutoProvisioning.URL + "/"
	}

	if string(Config.Global.Software.AutoProvisioning.SaveFilePath[len(Config.Global.Software.AutoProvisioning.SaveFilePath)-1]) != "/" {
		Config.Global.Software.AutoProvisioning.SaveFilePath = Config.Global.Software.AutoProvisioning.SaveFilePath + "/"
	}

	fileURL := Config.Global.Software.AutoProvisioning.URL + Config.Global.Software.AutoProvisioning.TkID
	log.Println("info: Trying to Autoprovision with URL: ", fileURL)
	err := downloadFile(Config.Global.Software.AutoProvisioning.SaveFilePath, Config.Global.Software.AutoProvisioning.SaveFilename, fileURL)
	if err != nil {
		return fmt.Errorf("error: DownloadFile Module Returned an Error: %q", err.Error())
	}

	return nil

}

func downloadFile(SaveFilePath string, SaveFilename string, URL string) error {
	resp, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("debug: HTTP Provisioning Server Responded With Status 200 OK ")
	} else {
		return fmt.Errorf("error: HTTP Provisioning Server Returned Status %q %q", resp.StatusCode, http.StatusText(resp.StatusCode))

	}

	out, err := os.Create(SaveFilePath + SaveFilename)
	if err != nil {
		return fmt.Errorf("error: Cannot Create File Error: %q", err.Error())
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error: Cannot Copy File Error: %q", err.Error())
	}

	return nil
}
