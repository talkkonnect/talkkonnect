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
	talkkonnectVersion  string = "2.39.01"
	talkkonnectReleased string = "18 Feb 2024"
)

/* Release Notes
Mic Volume Remote Control Available via Keyboard, HTTPAPI and MQTT Feature requested by Kekstar
to use it

via keyboard to map volumetxup to key 1 and volumetxdown to key 2 under the <keyboard> section of talkkonnect.xml

       <command action="volumetxup" paramname="" paramvalue="" enabled="true">
          <ttykeyboard scanid="49" keylabel="1" enabled="true"/>
          <usbkeyboard scanid="49" keylabel="1" enabled="true"/>
        </command>
        <command action="volumetxdown" paramname="" paramvalue="" enabled="true">
          <ttykeyboard scanid="50" keylabel="2" enabled="true"/>
          <usbkeyboard scanid="50" keylabel="2" enabled="true"/>
        </command>

control via httpapi <remotecontrol> section for talkkonnect
       <http listenport="8080" enabled="true">
          <command action="volumerxup" funcparamname="" message="Volume Up" enabled="true"/>
          <command action="volumerxdown" funcparamname="" message="Volume Down" enabled="true"/>
          <command action="volumetxup" funcparamname="" message="Volume Up" enabled="true"/>
          <command action="volumetxdown" funcparamname="" message="Volume Down" enabled="true"/>


		  then use browser to call for example http://aaa.bbb.ccc.ddd:8080/?command=volumetxup or http://aaa.bbb.ccc.ddd:8080/?command=volumetxdown


control via mqtt by <mqtt> in the command sections add

         <commands>
            <command action="volumerxup" message="Volume Up" enabled="true"/>
            <command action="volumerxdown" message="Volume Down" enabled="true"/>
            <command action="volumetxup" message="Volume Up" enabled="true"/>
            <command action="volumetxdown" message="Volume Down" enabled="true"/>
*/
