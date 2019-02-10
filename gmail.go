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
 * gmail.go -> talkkonnect funtion used to send email via GMAIL SMTP
 */

package talkkonnect

import (
	"errors"
	"fmt"
	hd44780 "github.com/talkkonnect/go-hd44780"
	"github.com/xackery/gomail"
)

func sendviagmail(username string, password string, receiver string, subject string, message string) error {

	err := gomail.Send(username, password, []string{receiver}, subject, message)
	if err != nil {
		return errors.New(fmt.Sprintf("Sending Email Via GMAIL Error: ", err.Error()))
	}

	go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin,LCDInterfaceType, LCDI2CAddress)

	return nil
}
