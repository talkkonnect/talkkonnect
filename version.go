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
 * The Initial Developer of the Original Code is Suvir Kumar <suvir@talkkonnect.com>
 *
 * Portions created by the Initial Developer are Copyright (C) Suvir Kumar. All Rights Reserved.
 *
 * Contributor(s):
 *
 * Suvir Kumar <suvir@talkkonnect.com>
 * Zoran Dimitrijevic
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * xmlparser.go -> talkkonnect functionality to read from XML file and populate global variables
 */

package talkkonnect

const (
	talkkonnectVersion  string = "2.07.21"
	talkkonnectReleased string = "Jan 12 2022"
)

/* Release Notes
1. Sanity Check for voiceactivitytimersecs if less that 200 msecs set to 200 msecs
2. Removed redundant definition of ttyKBStruct and usbStruct and combined into one struct
3. Removed unused definition of InputEventSoundStruct
4. Added Tighter GPIO Checks for Pins 2,3 If I2C Device enabled these pins should be avoided warning
5. Added Tighter GPIO Checks for Pins 7,8,9,10,11 if SPI Device Enabled these pins should be avoided warning
6. Added Tighter GPIO Checks for Pins in the rage 2-27 only
7. Added Debugging Verbosity to currently unhandled mumble events
8. Fixed Command Arguements for PaPlay
9. Changed version checking function instead of saying newer version available say different version available
*/
