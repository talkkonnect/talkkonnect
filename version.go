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
	talkkonnectVersion  string = "2.25.01"
	talkkonnectReleased string = "11 June  2023"
)

/* Release Notes
1. Voice Target for 5 GPIO Voice Target Buttons

XML Config Added for this feature

in accounts section Define Users that you want in the voice target ID can be one user or more user to each id
  <voicetargets>
        <id value="1">
            <user>suvir-demo</user>
        </id>
        <id value="2">
          <users>
            <user>suvir-ubunbtu</user>
          </users>
        </id>
        <id value="3">
          <users>
            <user>suvir-demo</user>
          </users>
        </id>
  <voicetargets>

added section in software just under memory channels to map the buttons to voicetarget id
  <presetvoicetargets enabled="true">
        <voicetargetset gpioname="presetvoicetarget1" id="0" enabled="true"/> <!-- ID = 0 Clears Voice Targets ->>
        <voicetargetset gpioname="presetvoicetarget2" id="1" enabled="true"/>
        <voicetargetset gpioname="presetvoicetarget3" id="2" enabled="true"/>
        <voicetargetset gpioname="presetvoicetarget4" id="3" enabled="true"/>
        <voicetargetset gpioname="presetvoicetarget5" id="4" enabled="true"/>
     </presetvoicetargets>

Added in GPIO
   <hardware targetboard="rpi">
      <io>
        <gpioexpander enabled="false">
        </gpioexpander>
        <max7219 max7219cascaded="1" spibus="0" spidevice="0" brightness="7" enabled="false"/>
        <pins>
          <pin direction="input"  device="pushbutton"    name="presetvoicetarget1" pinno="xx" type="gpio" chipid="0" enabled="true"/>
          <pin direction="input"  device="pushbutton"    name="presetvoicetarget2" pinno="xx" type="gpio" chipid="0" enabled="true"/>
          <pin direction="input"  device="pushbutton"    name="presetvoicetarget3" pinno="xx" type="gpio" chipid="0" enabled="true"/>
          <pin direction="input"  device="pushbutton"    name="presetvoicetarget4" pinno="xx" type="gpio" chipid="0" enabled="true"/>
          <pin direction="input"  device="pushbutton"    name="presetvoicetarget5" pinno="xx" type="gpio" chipid="0" enabled="true"/>
        </pins>


*/
