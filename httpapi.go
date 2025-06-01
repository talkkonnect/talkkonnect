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
 *
 */

package talkkonnect

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func (b *Talkkonnect) httpAPI(w http.ResponseWriter, r *http.Request) {
	funcs := map[string]interface{}{
		"displaymenu":        b.cmdDisplayMenu,
		"channelup":          b.cmdChannelUp,
		"channeldown":        b.cmdChannelDown,
		"mute-toggle":        b.cmdMuteUnmute,
		"mute":               b.cmdMuteUnmute,
		"unmute":             b.cmdMuteUnmute,
		"currentrxvolume":    b.cmdCurrentRXVolume,
		"volumerxup":         b.cmdVolumeRXUp,
		"volumerxdown":       b.cmdVolumeRXDown,
		"volumetxup":         b.cmdVolumeTXUp,
		"volumetxdown":       b.cmdVolumeTXDown,
		"currenttxvolume":    b.cmdCurrentTXVolume,
		"listserverchannels": b.cmdListServerChannels,
		"starttransmitting":  b.cmdStartTransmitting,
		"stoptransmitting":   b.cmdStopTransmitting,
		"listonlineusers":    b.cmdListOnlineUsers,
		"playback":           b.cmdPlayback,
		"gpsposition":        b.cmdGPSPosition,
		"sendemail":          b.cmdSendEmail,
		"previousserver":     b.cmdConnPreviousServer,
		"connnextserver":     b.cmdConnNextServer,
		"clearscreen":        b.cmdClearScreen,
		"pingservers":        b.cmdPingServers,
		"panicsimulation":    b.cmdPanicSimulation,
		"repeattxloop":       b.cmdRepeatTxLoop,
		"scanchannels":       b.cmdScanChannels,
		"thanks":             cmdThanks,
		"showuptime":         b.cmdShowUptime,
		"showversion":        b.cmdDisplayVersion,
		"dumpxmlconfig":      b.cmdDumpXMLConfig,
		"ttsannouncement":    b.TTSPlayerAPI,
		"voicetargetset":     b.cmdSendVoiceTargets,
		"listeningstart":     b.cmdListeningStart,
		"listeningstop":      b.cmdListeningStop,
		"listapi":            listAPI}

	APICommands, ok := r.URL.Query()["command"]

	if !ok  {
		log.Println("error: URL Param 'command' is missing example http API commands should be of the format http://a.b.c.d/?command=listapi")
		fmt.Fprintf(w, "error: API should be of the format http://a.b.c.d:"+Config.Global.Software.RemoteControl.HTTP.ListenPort+"/?command=StartTransmitting or of the format http://a.b.c.d:"+Config.Global.Software.RemoteControl.HTTP.ListenPort+"?command=setvoicetarget&id=0\n")
		return
	}

	var APIID int
	var APITTSMessage string
	var APITTSLocalPlay bool
	var APITTSPlayIntoStream bool
	var APIGPIOEnabled bool
	var APIGPIOName string
	var APIPreDelay int
	var APIPostDelay int
	var APILanguage string
	var err error

	APICommand := strings.ToLower(APICommands[0])
	APIDefined := false
	for _, apicommand := range Config.Global.Software.RemoteControl.HTTP.Command {
		if APICommand == "listapi" && apicommand.Enabled {
			fmt.Fprintf(w, "200 OK: API Command %v for %v Control Available\n", apicommand.Action, apicommand.Message)
		}
		if apicommand.Action == APICommand {
			APIDefined = true
		}
	}

	if !APIDefined {
		log.Printf("error: API Command %v Not A Valid Defined Command\n", APICommand)
		fmt.Fprintf(w, "404 error: API Command %v Not A Valid Defined Command\n", APICommand)
		return
	}

	for key, values := range r.URL.Query() {
		if strings.ToLower(key) == "command" {
			APICommand = values[0]
		}

		if strings.ToLower(key) == "id" {
			APIID, err = strconv.Atoi(values[0])
			if err != nil {
				log.Println("error: Target ID is not Number")
				fmt.Fprintf(w, "404 error: API VoiceTarget ID is not Number\n")
				return
			}
		}

		if strings.ToLower(key) == "ttsmessage" {
			APITTSMessage = values[0]
		}

		if strings.ToLower(key) == "ttslocalplay" {
			var temp string = values[0]
			if strings.ToLower(temp) == "true" {
				APITTSLocalPlay = true
			}
			if strings.ToLower(temp) == "false" {
				APITTSLocalPlay = false
			}
		}

		if strings.ToLower(key) == "ttsplayintostream" {
			var temp string = values[0]
			if strings.ToLower(temp) == "true" {
				APITTSPlayIntoStream = true
			}
			if strings.ToLower(temp) == "false" {
				APITTSPlayIntoStream = false
			}
		}

		if strings.ToLower(key) == "gpioenabled" {
			var temp string = values[0]
			if strings.ToLower(temp) == "true" {
				APIGPIOEnabled = true
			}
			if strings.ToLower(temp) == "false" {
				APIGPIOEnabled = false
			}
		}

		if strings.ToLower(key) == "gpioname" {
			APIGPIOName = values[0]
		}

		if strings.ToLower(key) == "predelay" {
			APIPreDelay, err = strconv.Atoi(values[0])
			if err != nil {
				log.Println("error: PreDelay is not Number")
				fmt.Fprintf(w, "404 error: API PreDelay is not Number\n")
				return
			}
		}

		if strings.ToLower(key) == "postdelay" {
			APIPostDelay, err = strconv.Atoi(values[0])
			if err != nil {
				log.Println("error: PostDelay is not Number")
				fmt.Fprintf(w, "404 error: API PostDelay is not Number\n")
				return
			}
		}

		if strings.ToLower(key) == "language" {
			APILanguage = values[0]
		}

	}

	for _, apicommand := range Config.Global.Software.RemoteControl.HTTP.Command {
		if apicommand.Action == APICommand {
			if len(apicommand.Funcparamname) == 0 {
				_, err := b.Call(funcs, apicommand.Action)
				if err != nil {
					log.Println("error: Wrong Parameters to Call Function")
				} else {
					fmt.Fprintf(w, "200 OK: http command %v OK \n", APICommand)
				}
			} else {
				if apicommand.Funcparamname != "value" {
					_, err := b.Call(funcs, apicommand.Action, apicommand.Funcparamname)
					if err != nil {
						log.Println("error: Wrong Parameters to Call Function")
					} else {
						fmt.Fprintf(w, "200 OK: http command %v For %v Control\n", apicommand.Action, apicommand.Message)
					}
				} else {
					switch APICommand {
					case "voicetargetset":
						_, err := b.Call(funcs, apicommand.Action, uint32(APIID))
						if err != nil {
							log.Println("error: Wrong Parameters to Call Function")
						} else {
							fmt.Fprintf(w, "200 OK: http command %v OK \n", APICommand)
						}
					case "ttsannouncement":
						_, err := b.Call(funcs, apicommand.Action, APITTSMessage, APITTSLocalPlay, APITTSPlayIntoStream, APIGPIOEnabled, APIGPIOName, time.Duration(APIPreDelay*int(time.Second)), time.Duration(APIPostDelay)*time.Second, APILanguage)
						if err != nil {
							log.Println("error: Wrong Parameters to Call Function")
						} else {
							fmt.Fprintf(w, "200 OK: http command %v OK \n", APICommand)
						}
					}
				}
			}
		}
	}
}

func (b *Talkkonnect) Call(m map[string]interface{}, name string, params ...interface{}) (result []reflect.Value, err error) {
	f := reflect.ValueOf(m[name])
	if len(params) != f.Type().NumIn() {
		err = errors.New("the number of params is not adapted")
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}

func listAPI() {
	for _, apicommand := range Config.Global.Software.RemoteControl.HTTP.Command {
		log.Printf("info: API Command %v for %v Control Available\n", apicommand.Action, apicommand.Message)
	}
}
