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
	talkkonnectVersion  string = "2.43.02"
	talkkonnectReleased string = "09 Feb 2025"
)

/* Release Notes
Version 2.43.02
fixed voice target clear in clientcommands that was mistakenly commented out causing
users not to be able to clear voicetargets once set
Version 2.43.01
added gpio offset setting for new raspberry pi os support
revamped repeater tone functionality

Version 2.42.01
fixed data structure for repeater tone reading values from xml to talkkonnect config structure bug

Version 2.41.01
cleaned the logic for LED of participants

Version 2.40.02
fixed checking of gpio to allow gpio pins 1 & 2 so it can work with mcp23107 expander
bug report by Monorajan


Version 2.40.01
Stop talkkonnect from crashing on unaccessable channels. If the server doesnt send talkkonnect the
channel permissions on connection then the user will not be able to change channels from the channel
he initially connected


Version 2.39.02
Added currenttxvolume query to all modes of communication including ttykeyboard, usbkeyboard, httpapi, mqtt

Version 2.39.01
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
          <command action="displaymenu" funcparamname="" message="Display Menu" enabled="true"/>
          <command action="channelup" funcparamname="" message="Channel Up" enabled="true"/>
          <command action="channeldown" funcparamname="" message="Channel Down" enabled="true"/>
          <command action="mute-toggle" funcparamname="toggle" message="Mute-Toggle" enabled="true"/>
          <command action="mute" funcparamname="mute" message="Mute" enabled="true"/>
          <command action="unmute" funcparamname="unmute" message="Unute" enabled="true"/>
          <command action="currentrxvolume" funcparamname="" message="Current Volume" enabled="true"/>
          <command action="volumerxup" funcparamname="" message="Volume Up" enabled="true"/>
          <command action="volumerxdown" funcparamname="" message="Volume Down" enabled="true"/>
          <command action="currenttxvolume" funcparamname="" message="Current Volume" enabled="true"/>
          <command action="volumetxup" funcparamname="" message="Volume Up" enabled="true"/>
          <command action="volumetxdown" funcparamname="" message="Volume Down" enabled="true"/>
          <command action="listserverchannels" funcparamname="" message="List Channels" enabled="true"/>
          <command action="starttransmitting" funcparamname="" message="Start Transmitting" enabled="true"/>
          <command action="stoptransmitting" funcparamname="" message="Stop Transmitting" enabled="true"/>
          <command action="listonlineusers" funcparamname="" message="List Users" enabled="true"/>
          <command action="playback" funcparamname="" message="Playback" enabled="true"/>
          <command action="gpsposition" funcparamname="" message="GPS Position" enabled="true"/>
          <command action="sendemail" funcparamname="" message="Send Email" enabled="true"/>
          <command action="previousserver" funcparamname="" message="Previous Server" enabled="true"/>
          <command action="connnextserver" funcparamname="" message="Next Server" enabled="true"/>
          <command action="clearscreen" funcparamname="" message="Clear Screen" enabled="true"/>
          <command action="pingservers" funcparamname="" message="Ping Servers" enabled="true"/>
          <command action="panicsimulation" funcparamname="" message="Panic Simulation" enabled="true"/>
          <command action="repeattxloop" funcparamname="" message="Repeat TX Loop" enabled="true"/>
          <command action="scanchannels" funcparamname="" message="Scan Channels" enabled="true"/>
          <command action="thanks" funcparamname="" message="Thanks" enabled="true"/>
          <command action="showuptime" funcparamname="" message="Show UpTime" enabled="true"/>
          <command action="showversion" funcparamname="" message="Show Version" enabled="true"/>
          <command action="dumpxmlconfig" funcparamname="" message="Dump XML Config" enabled="true"/>
          <command action="ttsannouncement" funcparamname="value" message="TTS Announcement" enabled="true"/>
          <command action="voicetargetset" funcparamname="value" message="Set Voice Target" enabled="true"/>
          <command action="listapi" funcparamname="" message="List API" enabled="true"/>
        </http>

then use browser to call for example http://aaa.bbb.ccc.ddd:8080/?command=volumetxup or http://aaa.bbb.ccc.ddd:8080/?command=volumetxdown fir example


control via mqtt by <mqtt> in the command sections add

          <commands>
            <command action="displaymenu" message="Display Menu" enabled="true"/>
            <command action="channelup" message="Channel Up" enabled="true"/>
            <command action="channeldown" message="Channel Down" enabled="true"/>
            <command action="muteunmute" message="Mute-Toggle" enabled="true"/>
            <command action="currentrxvolume" message="Current Volume" enabled="true"/>
            <command action="volumerxup" message="Volume Up" enabled="true"/>
            <command action="volumerxdown" message="Volume Down" enabled="true"/>
            <command action="currenttxvolume" message="Current Volume" enabled="true"/>
            <command action="volumetxup" message="Volume Up" enabled="true"/>
            <command action="volumetxdown" message="Volume Down" enabled="true"/>
            <command action="listserverchannels" message="List Channels" enabled="true"/>
            <command action="starttransmitting" message="Start Transmitting" enabled="true"/>
            <command action="stoptransmitting" message="Stop Transmitting" enabled="true"/>
            <command action="listonlineusers" message="List Users" enabled="true"/>
            <command action="playback" message="Playback" enabled="true"/>
            <command action="gpsposition" message="GPS Position" enabled="true"/>
            <command action="sendemail" message="Send Email" enabled="true"/>
            <command action="previousserver" message="Previous Server" enabled="true"/>
            <command action="connnextserver" message="Next Server" enabled="true"/>
            <command action="clearscreen" message="Clear Screen" enabled="true"/>
            <command action="pingservers" message="Ping Servers" enabled="true"/>
            <command action="panicsimulation" message="Panic Simulation" enabled="true"/>
            <command action="repeattxloop" message="Repeat TX Loop" enabled="true"/>
            <command action="scanchannels" message="Scan Channels" enabled="true"/>
            <command action="thanks" message="Thanks" enabled="true"/>
            <command action="showuptime" message="Show UpTime" enabled="true"/>
            <command action="dumpxmlconfig" message="Dump XML Config" enabled="true"/>
            <command action="ttsannouncement" message="TTS Announcement" enabled="true"/>
            <command action="voicetargetset" message="Set Voice Target" enabled="true"/>
            <command action="attention" message="Attention LED" enabled="true"/>
            <command action="relay" message="RelayControl" enabled="true"/>
          </commands>
*/
