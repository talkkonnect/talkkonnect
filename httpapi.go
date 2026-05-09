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
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// remoteAPIQuery holds parameters for remote commands (HTTP query or bottom CLI).
type remoteAPIQuery struct {
	Command              string
	ID                   int
	APITTSMessage        string
	APITTSLocalPlay      bool
	APITTSPlayIntoStream bool
	APIGPIOEnabled       bool
	APIGPIOName          string
	APIPreDelay          int
	APIPostDelay         int
	APILanguage          string
}

func (b *Talkkonnect) remoteAPICommandHandlers() map[string]interface{} {
	return map[string]interface{}{
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
		"radiotoggle":        b.cmdInternetRadioToggle,
		"radionext":          b.cmdInternetRadioNext,
		"radioprev":          b.cmdInternetRadioPrev,
		"radiovolup":         b.cmdInternetRadioVolUp,
		"radiovoldown":       b.cmdInternetRadioVolDown,
		"listapi":            listAPI,
	}
}

func fillHTTPRemoteAPIQueryFromRequest(r *http.Request, q *remoteAPIQuery) error {
	var err error
	for key, values := range r.URL.Query() {
		if len(values) == 0 {
			continue
		}
		switch strings.ToLower(key) {
		case "command":
			q.Command = strings.ToLower(strings.TrimSpace(values[0]))
		case "id":
			q.ID, err = strconv.Atoi(values[0])
			if err != nil {
				return errors.New("voice target id is not a number")
			}
		case "ttsmessage":
			q.APITTSMessage = values[0]
		case "ttslocalplay":
			switch strings.ToLower(values[0]) {
			case "true":
				q.APITTSLocalPlay = true
			case "false":
				q.APITTSLocalPlay = false
			}
		case "ttsplayintostream":
			switch strings.ToLower(values[0]) {
			case "true":
				q.APITTSPlayIntoStream = true
			case "false":
				q.APITTSPlayIntoStream = false
			}
		case "gpioenabled":
			switch strings.ToLower(values[0]) {
			case "true":
				q.APIGPIOEnabled = true
			case "false":
				q.APIGPIOEnabled = false
			}
		case "gpioname":
			q.APIGPIOName = values[0]
		case "predelay":
			q.APIPreDelay, err = strconv.Atoi(values[0])
			if err != nil {
				return errors.New("predelay is not a number")
			}
		case "postdelay":
			q.APIPostDelay, err = strconv.Atoi(values[0])
			if err != nil {
				return errors.New("postdelay is not a number")
			}
		case "language":
			q.APILanguage = values[0]
		}
	}
	return nil
}

// HandleRemoteAPICommand runs one configured HTTP API command (used by HTTP handler and bottom CLI).
func (b *Talkkonnect) HandleRemoteAPICommand(w io.Writer, q remoteAPIQuery) {
	funcs := b.remoteAPICommandHandlers()
	APICommand := strings.ToLower(strings.TrimSpace(q.Command))

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

	for _, apicommand := range Config.Global.Software.RemoteControl.HTTP.Command {
		if apicommand.Action != APICommand {
			continue
		}
		if len(apicommand.Funcparamname) == 0 {
			_, err := b.Call(funcs, apicommand.Action)
			if err != nil {
				log.Println("error: Wrong Parameters to Call Function")
				fmt.Fprintf(w, "500 error: wrong parameters for command %q\n", APICommand)
			} else {
				fmt.Fprintf(w, "200 OK: http command %v OK \n", APICommand)
			}
		} else {
			if apicommand.Funcparamname != "value" {
				_, err := b.Call(funcs, apicommand.Action, apicommand.Funcparamname)
				if err != nil {
					log.Println("error: Wrong Parameters to Call Function")
					fmt.Fprintf(w, "500 error: wrong parameters for command %q\n", APICommand)
				} else {
					fmt.Fprintf(w, "200 OK: http command %v For %v Control\n", apicommand.Action, apicommand.Message)
				}
			} else {
				switch APICommand {
				case "voicetargetset":
					_, err := b.Call(funcs, apicommand.Action, uint32(q.ID))
					if err != nil {
						log.Println("error: Wrong Parameters to Call Function")
						fmt.Fprintf(w, "500 error: wrong parameters for command %q\n", APICommand)
					} else {
						fmt.Fprintf(w, "200 OK: http command %v OK \n", APICommand)
					}
				case "ttsannouncement":
					_, err := b.Call(funcs, apicommand.Action, q.APITTSMessage, q.APITTSLocalPlay, q.APITTSPlayIntoStream, q.APIGPIOEnabled, q.APIGPIOName, time.Duration(q.APIPreDelay*int(time.Second)), time.Duration(q.APIPostDelay)*time.Second, q.APILanguage)
					if err != nil {
						log.Println("error: Wrong Parameters to Call Function")
						fmt.Fprintf(w, "500 error: wrong parameters for command %q\n", APICommand)
					} else {
						fmt.Fprintf(w, "200 OK: http command %v OK \n", APICommand)
					}
				}
			}
		}
	}
}

func (b *Talkkonnect) httpAPI(w http.ResponseWriter, r *http.Request) {
	APICommands, ok := r.URL.Query()["command"]
	if !ok {
		log.Println("error: URL Param 'command' is missing example http API commands should be of the format http://a.b.c.d/?command=listapi")
		fmt.Fprintf(w, "error: API should be of the format http://a.b.c.d:"+Config.Global.Software.RemoteControl.HTTP.ListenPort+"/?command=StartTransmitting or of the format http://a.b.c.d:"+Config.Global.Software.RemoteControl.HTTP.ListenPort+"?command=setvoicetarget&id=0\n")
		return
	}

	q := remoteAPIQuery{Command: strings.ToLower(strings.TrimSpace(APICommands[0]))}
	APIDefined := false
	for _, apicommand := range Config.Global.Software.RemoteControl.HTTP.Command {
		if apicommand.Action == q.Command {
			APIDefined = true
			break
		}
	}
	if !APIDefined {
		log.Printf("error: API Command %v Not A Valid Defined Command\n", q.Command)
		fmt.Fprintf(w, "404 error: API Command %v Not A Valid Defined Command\n", q.Command)
		return
	}

	if err := fillHTTPRemoteAPIQueryFromRequest(r, &q); err != nil {
		log.Println("error: " + err.Error())
		fmt.Fprintf(w, "404 error: API %v\n", err.Error())
		return
	}

	b.HandleRemoteAPICommand(w, q)
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
		msg := fmt.Sprintf("info: API Command %v for %v Control Available\n", apicommand.Action, apicommand.Message)
		log.Print(msg)
		sshRemoteReplyF(msg)
	}
}
