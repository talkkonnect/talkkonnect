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
	"regexp"
	"strconv"
	"strings"
	"time"
)

// remoteAPICommandTokenRE allows only safe command tokens: lowercase API names with optional hyphens.
var remoteAPICommandTokenRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

func remoteAPIValidateBuiltinCommand(b *Talkkonnect, cmd string) error {
	c := strings.ToLower(strings.TrimSpace(cmd))
	if c == "" {
		return errors.New("command is empty")
	}
	if !remoteAPICommandTokenRE.MatchString(c) {
		return errors.New("command must be a lowercase letter or digit followed only by letters, digits, or hyphens")
	}
	if _, ok := b.remoteAPICommandHandlers()[c]; !ok {
		return errors.New("command is not in the built-in allow list")
	}
	return nil
}

// remoteAPIQuery holds parameters for remote commands (HTTP query or bottom CLI).
type remoteAPIQuery struct {
	Command              string
	ID                   int
	APIChannel           string
	APIUser              string
	APIVolume            int
	APIVolumeSet         bool
	APIMediaID           string
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
		"setrxvolume":        b.cmdSetRXVolume,
		"listserverchannels": b.cmdListServerChannels,
		"joinchannel":        b.cmdJoinChannel,
		"whisperuser":        b.cmdWhisperUser,
		"whisperclear":       b.cmdWhisperClear,
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
		"announcement":       b.cmdAnnouncement,
		"voicetargetset":     b.cmdSendVoiceTargets,
		"listeningstart":     b.cmdListeningStart,
		"listeningstop":      b.cmdListeningStop,
		"radiotoggle":        b.cmdInternetRadioToggle,
		"radionext":          b.cmdInternetRadioNext,
		"radioprev":          b.cmdInternetRadioPrev,
		"radiovolup":         b.cmdInternetRadioVolUp,
		"radiovoldown":       b.cmdInternetRadioVolDown,
		"multicaston":        b.cmdMulticastOn,
		"multicastoff":       b.cmdMulticastOff,
		"multicasttoggle":    b.cmdMulticastToggle,
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
		case "id":
			q.ID, err = strconv.Atoi(values[0])
			if err != nil {
				return errors.New("voice target id is not a number")
			}
		case "channel":
			q.APIChannel = strings.TrimSpace(values[0])
		case "user":
			q.APIUser = strings.TrimSpace(values[0])
		case "volume":
			q.APIVolume, err = strconv.Atoi(values[0])
			if err != nil {
				return errors.New("volume is not a number")
			}
			q.APIVolumeSet = true
		case "mediaid":
			q.APIMediaID = strings.TrimSpace(values[0])
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

// remoteAPIBadRequest rejects a command whose query parameters are missing or out
// of range, answering an HTTP caller with a 400 and a CLI caller with plain text.
func remoteAPIBadRequest(w io.Writer, hw http.ResponseWriter, isHTTP bool, message string) {
	log.Println("error: " + message)
	if isHTTP {
		http.Error(hw, "400 bad request: "+message, http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "400 bad request: %s\n", message)
}

// remoteAPICallWithParams invokes a command handler that takes query parameters
// and writes either ack or the same 500 the parameterless path writes.
func (b *Talkkonnect) remoteAPICallWithParams(w io.Writer, hw http.ResponseWriter, isHTTP bool, funcs map[string]interface{}, command, ack string, params ...interface{}) {
	if _, err := b.Call(funcs, command, params...); err != nil {
		log.Println("error: Wrong Parameters to Call Function")
		if isHTTP {
			http.Error(hw, fmt.Sprintf("500 internal server error: wrong parameters for command %q", command), http.StatusInternalServerError)
		} else {
			fmt.Fprintf(w, "500 error: wrong parameters for command %q\n", command)
		}
		return
	}
	fmt.Fprint(w, ack)
}

// HandleRemoteAPICommand runs one configured HTTP API command (used by HTTP handler and bottom CLI).
func (b *Talkkonnect) HandleRemoteAPICommand(w io.Writer, q remoteAPIQuery) {
	hw, isHTTP := w.(http.ResponseWriter)

	funcs := b.remoteAPICommandHandlers()
	APICommand := strings.ToLower(strings.TrimSpace(q.Command))

	if err := remoteAPIValidateBuiltinCommand(b, APICommand); err != nil {
		log.Printf("error: remote API command %q rejected: %v\n", APICommand, err)
		if isHTTP {
			http.Error(hw, "400 bad request: "+err.Error(), http.StatusBadRequest)
		} else {
			fmt.Fprintf(w, "400 bad request: %v\n", err)
		}
		return
	}

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
		if isHTTP {
			http.Error(hw, fmt.Sprintf("404 not found: API command %q is not defined in configuration", APICommand), http.StatusNotFound)
		} else {
			fmt.Fprintf(w, "404 error: API Command %v Not A Valid Defined Command\n", APICommand)
		}
		return
	}

	// Command handlers report their human-readable results through sshRemoteReplyF
	// (uptime, the online user list, per-server ping results, radio state). Capture
	// that text for an HTTP caller so a dashboard such as tk-webmonitor can show
	// what the command actually said instead of only the "200 OK" acknowledgement.
	// A bottom CLI caller already has its own writer attached, and listapi prints
	// its listing to w directly, so neither needs capturing here.
	var captured replyCapture
	if isHTTP && APICommand != "listapi" {
		replyID := sshRemoteReplyAttach(&captured)
		defer sshRemoteReplyDetach(replyID)
		defer func() {
			if out := strings.Trim(captured.String(), "\n"); out != "" {
				fmt.Fprintf(w, "%s\n", out)
			}
		}()
	}

	for _, apicommand := range Config.Global.Software.RemoteControl.HTTP.Command {
		if apicommand.Action != APICommand {
			continue
		}
		if len(apicommand.Funcparamname) == 0 {
			_, err := b.Call(funcs, apicommand.Action)
			if err != nil {
				log.Println("error: Wrong Parameters to Call Function")
				if isHTTP {
					http.Error(hw, fmt.Sprintf("500 internal server error: wrong parameters for command %q", APICommand), http.StatusInternalServerError)
				} else {
					fmt.Fprintf(w, "500 error: wrong parameters for command %q\n", APICommand)
				}
			} else {
				fmt.Fprintf(w, "200 OK: http command %v OK \n", APICommand)
			}
		} else {
			if apicommand.Funcparamname != "value" {
				_, err := b.Call(funcs, apicommand.Action, apicommand.Funcparamname)
				if err != nil {
					log.Println("error: Wrong Parameters to Call Function")
					if isHTTP {
						http.Error(hw, fmt.Sprintf("500 internal server error: wrong parameters for command %q", APICommand), http.StatusInternalServerError)
					} else {
						fmt.Fprintf(w, "500 error: wrong parameters for command %q\n", APICommand)
					}
				} else {
					fmt.Fprintf(w, "200 OK: http command %v For %v Control\n", apicommand.Action, apicommand.Message)
				}
			} else {
				switch APICommand {
				case "joinchannel":
					if q.APIChannel == "" {
						remoteAPIBadRequest(w, hw, isHTTP, "joinchannel requires query parameter \"channel\"")
						break
					}
					b.remoteAPICallWithParams(w, hw, isHTTP, funcs, APICommand,
						fmt.Sprintf("200 OK: http command %v for channel %v\n", APICommand, q.APIChannel), q.APIChannel)
				case "whisperuser":
					if q.APIUser == "" {
						remoteAPIBadRequest(w, hw, isHTTP, "whisperuser requires query parameter \"user\"")
						break
					}
					b.remoteAPICallWithParams(w, hw, isHTTP, funcs, APICommand,
						fmt.Sprintf("200 OK: http command %v for user %v\n", APICommand, q.APIUser), q.APIUser)
				case "setrxvolume":
					if !q.APIVolumeSet {
						remoteAPIBadRequest(w, hw, isHTTP, "setrxvolume requires query parameter \"volume\"")
						break
					}
					if q.APIVolume < 0 || q.APIVolume > 100 {
						remoteAPIBadRequest(w, hw, isHTTP, "setrxvolume volume must be between 0 and 100")
						break
					}
					b.remoteAPICallWithParams(w, hw, isHTTP, funcs, APICommand,
						fmt.Sprintf("200 OK: http command %v to %v%%\n", APICommand, q.APIVolume), q.APIVolume)
				case "voicetargetset":
					_, err := b.Call(funcs, apicommand.Action, uint32(q.ID))
					if err != nil {
						log.Println("error: Wrong Parameters to Call Function")
						if isHTTP {
							http.Error(hw, fmt.Sprintf("500 internal server error: wrong parameters for command %q", APICommand), http.StatusInternalServerError)
						} else {
							fmt.Fprintf(w, "500 error: wrong parameters for command %q\n", APICommand)
						}
					} else {
						fmt.Fprintf(w, "200 OK: http command %v OK \n", APICommand)
					}
				case "ttsannouncement":
					_, err := b.Call(funcs, apicommand.Action, q.APITTSMessage, q.APITTSLocalPlay, q.APITTSPlayIntoStream, q.APIGPIOEnabled, q.APIGPIOName, time.Duration(q.APIPreDelay*int(time.Second)), time.Duration(q.APIPostDelay)*time.Second, q.APILanguage)
					if err != nil {
						log.Println("error: Wrong Parameters to Call Function")
						if isHTTP {
							http.Error(hw, fmt.Sprintf("500 internal server error: wrong parameters for command %q", APICommand), http.StatusInternalServerError)
						} else {
							fmt.Fprintf(w, "500 error: wrong parameters for command %q\n", APICommand)
						}
					} else {
						fmt.Fprintf(w, "200 OK: http command %v OK \n", APICommand)
					}
				case "announcement":
					if q.APIMediaID == "" {
						log.Println("error: announcement command requires mediaid query parameter")
						if isHTTP {
							http.Error(hw, "400 bad request: missing required query parameter \"mediaid\"", http.StatusBadRequest)
						} else {
							fmt.Fprintf(w, "400 bad request: missing required query parameter \"mediaid\"\n")
						}
						break
					}
					_, err := b.Call(funcs, apicommand.Action, q.APIMediaID)
					if err != nil {
						log.Println("error: Wrong Parameters to Call Function")
						if isHTTP {
							http.Error(hw, fmt.Sprintf("500 internal server error: wrong parameters for command %q", APICommand), http.StatusInternalServerError)
						} else {
							fmt.Fprintf(w, "500 error: wrong parameters for command %q\n", APICommand)
						}
					} else {
						fmt.Fprintf(w, "200 OK: http command %v started for mediaid %v\n", APICommand, q.APIMediaID)
					}
				}
			}
		}
	}
}

func (b *Talkkonnect) httpAPI(w http.ResponseWriter, r *http.Request) {
	if !remoteControlHTTPClientIPAllowed(r) {
		log.Printf("error: HTTP API request from %q rejected by remote control network ACL\n", r.RemoteAddr)
		http.Error(w, "403 forbidden: client address not allowed by remote control network ACL", http.StatusForbidden)
		return
	}

	APICommands, ok := r.URL.Query()["command"]
	if !ok || len(APICommands) == 0 || strings.TrimSpace(APICommands[0]) == "" {
		log.Println("error: URL Param 'command' is missing example http API commands should be of the format http://a.b.c.d/?command=listapi")
		http.Error(w, "400 bad request: missing required query parameter \"command\" (example: ?command=listapi)", http.StatusBadRequest)
		return
	}

	q := remoteAPIQuery{Command: strings.ToLower(strings.TrimSpace(APICommands[0]))}
	if err := fillHTTPRemoteAPIQueryFromRequest(r, &q); err != nil {
		log.Println("error: " + err.Error())
		http.Error(w, "400 bad request: "+err.Error(), http.StatusBadRequest)
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
		sshRemoteReplyF(msg, "")
	}
}
