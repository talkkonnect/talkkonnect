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
 * oleddisplay.go modified to work with talkkonnect
 */

package talkkonnect

import (
	"log"
	"strings"
	"sync"
)

var mutex = &sync.Mutex{}

func oledDisplay(OledClear bool, OledRow int, OledColumn int, OledText string) {
	mutex.Lock()
	defer mutex.Unlock()

	if OLEDEnabled == false {
		log.Println("error: OLED Function Called in Error!")
		return
	}

	if OLEDInterfacetype != "i2c" {
		log.Println("error: Only i2c OLED Screens Supported Now!")
		return
	}

	if OledClear == false && len(OledText) > 0 && LCDIsDark == true {
		Oled.DisplayOn()
	}

	if OledClear == true {
		Oled.Clear()
		log.Println("debug: OLED Clearing Screen")
		if len(OledText) == 0 {
			return
		}
	}

	Oled.SetCursor(OledRow, 0)

	var rpadding = int(OLEDDisplayColumns)

	if len(OledText) <= int(OLEDDisplayColumns) {
		rpadding = int(OLEDDisplayColumns) - len(OledText)
	}

	var text string = OledText + strings.Repeat(" ", rpadding)

	Oled.SetCursor(OledRow, OLEDStartColumn)

	if len(OledText) >= int(OLEDDisplayColumns) {
		Oled.Write(OledText[:OLEDDisplayColumns])
	} else {
		Oled.Write(text)
	}
}
