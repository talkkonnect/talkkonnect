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
 * talKKonnectContributor(s):
 *
 * Suvir Kumar <suvir@talkkonnect.com>
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * MQTT License Details Copyright (c) 2013 IBM Corp.
 *
 * This project is dual licensed under the Eclipse Public License 1.0 and the
 * Eclipse Distribution License 1.0 as described in the epl-v10 and edl-v10 files.
 * The EDL is copied below in order to pass the pkg.go.dev license check (https://pkg.go.dev/license-policy).
 * Eclipse Distribution License - v 1.0
 * Copyright (c) 2007, Eclipse Foundation, Inc. and its licensors.
 * All rights reserved.
 * Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
 *
 * Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
 * Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
 * Neither the name of the Eclipse Foundation, Inc. nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 * All rights reserved. This program and the accompanying materials are made available under the terms of the Eclipse Public License v1.0
 * which accompanies this distribution, and is available at http://www.eclipse.org/legal/epl-v10.html
 *
 */

package talkkonnect

import (
	"crypto/tls"
	"log"
	"strconv"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var MQTTPublishPayload MQTT.Token
var MQTTClient MQTT.Client

func (b *Talkkonnect) mqttsubscribe() {
	if Config.Global.Software.RemoteControl.MQTT.Enabled {
		log.Printf("info: MQTT Subscription Information")
		log.Printf("info: MQTT Broker      : %s\n", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTBroker)
		log.Printf("debug: MQTT clientid    : %s\n", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTId)
		log.Printf("debug: MQTT user        : %s\n", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTUser)
		log.Printf("debug: MQTT password    : %s\n", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTPassword)
		log.Printf("info: Subscribed topic : %s\n", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTSubTopic)

		connOpts := MQTT.NewClientOptions().AddBroker(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTBroker).SetClientID(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTId).SetCleanSession(true)
		if Config.Global.Software.RemoteControl.MQTT.Settings.MQTTUser != "" {
			connOpts.SetUsername(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTUser)
			if Config.Global.Software.RemoteControl.MQTT.Settings.MQTTPassword != "" {
				connOpts.SetPassword(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTPassword)
			}
		}
		tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
		connOpts.SetTLSConfig(tlsConfig)

		connOpts.OnConnect = func(c MQTT.Client) {
			if token := c.Subscribe(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTSubTopic, byte(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTQos), b.onMessageReceived); token.Wait() && token.Error() != nil {
				log.Println("error: MQTT Token Error!")
				return
			}
		}

		MQTTClient = MQTT.NewClient(connOpts)
		if token := MQTTClient.Connect(); token.Wait() && token.Error() != nil {
			log.Println("error: MQTT Token Error!")
			return
		} else {
			log.Printf("info: Connected to     : %s\n", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTBroker)
		}
	}
}

func MQTTPublish(mqttPayload string) {
	MQTTPublishPayload = MQTTClient.Publish(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTPubTopic, Config.Global.Software.RemoteControl.MQTT.Settings.MQTTQos, Config.Global.Software.RemoteControl.MQTT.Settings.MQTTRetained, mqttPayload)
	go func() {
		<-MQTTPublishPayload.Done()
		if MQTTPublishPayload.Error() != nil {
			log.Println("error: ", MQTTPublishPayload.Error())
		} else {
			log.Printf("info: Successfully Published MQTT Topic %v Payload %v\n", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTPubTopic, mqttPayload)
			return
		}
	}()
}

func (b *Talkkonnect) onMessageReceived(client MQTT.Client, message MQTT.Message) {

	var (
		CommandDefined bool
		PayLoad        string
	)

	funcs := map[string]interface{}{
		"displaymenu":        b.cmdDisplayMenu,
		"channelup":          b.cmdChannelUp,
		"channeldown":        b.cmdChannelDown,
		"muteunmute":         b.cmdMuteUnmute,
		"currentrxvolume":    b.cmdCurrentRXVolume,
		"volumerxup":         b.cmdVolumeRXUp,
		"volumerxdown":       b.cmdVolumeRXDown,
		"currenttxvolume":    b.cmdCurrentTXVolume,
		"volumetxup":         b.cmdVolumeTXUp,
		"volumetxdown":       b.cmdVolumeTXDown,
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
		"dumpxmlconfig":      b.cmdDumpXMLConfig,
		"voicetargetset":     b.cmdSendVoiceTargets,
		"listeningstart":     b.cmdListeningStart,
		"listeningstop":      b.cmdListeningStop,
		"attention":          attention,
		"relay":              relay}

	PayLoad = strings.ToLower(string(message.Payload()))
	log.Printf("info: Received MQTT message on topic: %s Payload: %s\n", message.Topic(), PayLoad)

	byteCommand := strings.Split(strings.ToLower(PayLoad), ":")
	stringCommand := strings.Join(byteCommand[:], "")
	Command := strings.Split(stringCommand, " ")

	for _, mqttcommand := range Config.Global.Software.RemoteControl.MQTT.Commands.Command {
		if Command[0] == strings.ToLower(mqttcommand.Action) {
			CommandDefined = true
			break
		}
	}

	if !CommandDefined {
		log.Printf("error: MQTT Command %v Not Defined\n", Command[0])
		return
	}

	for _, mqttcommand := range Config.Global.Software.RemoteControl.MQTT.Commands.Command {
		if strings.Contains(Command[0], mqttcommand.Action) {
			if mqttcommand.Enabled {
				log.Print("alert : mqttcommand ", mqttcommand)
				var Err error
				switch Command[0] {
				case "muteunmute":
					if len(Command) == 2 {
						if Command[1] == "toggle" {
							_, Err = b.Call(funcs, mqttcommand.Action, "mute-toggle")
						}
						if Command[1] == "mute" {
							_, Err = b.Call(funcs, mqttcommand.Action, "mute")
						}
						if Command[1] == "unmute" {
							_, Err = b.Call(funcs, mqttcommand.Action, "unmute")
						}
					} else {
						log.Println("error: Malformed MQTT Command")
					}
				case "attention":
					if len(Command) == 2 {
						if Command[1] == "blink" {
							_, Err = b.Call(funcs, mqttcommand.Action, "blink")
						}
						if Command[1] == "on" {
							_, Err = b.Call(funcs, mqttcommand.Action, "on")
						}
						if Command[1] == "off" {
							_, Err = b.Call(funcs, mqttcommand.Action, "off")
						}
					} else {
						log.Println("error: Malformed MQTT Command")
					}
				case "relay":
					if len(Command) == 3 {
						if Command[2] == "pulse" {
							_, Err = b.Call(funcs, mqttcommand.Action, "pulse", Command[3])
						}
						if Command[2] == "on" {
							_, Err = b.Call(funcs, mqttcommand.Action, "on", Command[3])
						}
						if Command[2] == "off" {
							_, Err = b.Call(funcs, mqttcommand.Action, "off", Command[3])
						}
					} else {
						log.Println("error: Malformed MQTT Command")
					}
				case "voicetargetset":
					if len(Command) == 2 {
						id, err := strconv.Atoi(Command[1])
						if err != nil {
							return
						}
						if id < 32 {
							_, Err = b.Call(funcs, mqttcommand.Action, uint32(id))
						} else {
							log.Println("error: Value of Target ID Not In Range (0-31) ", id)
						}
					} else {
						log.Println("error: Malformed MQTT Command")
					}
				default:
					if len(Command) == 1 {
						_, Err = b.Call(funcs, mqttcommand.Action)
					}
				}

				if Err == nil {
					log.Printf("MQTT Command %v Processed", Command)
				} else {
					log.Printf("error: MQTT Command %v Failed", Command)
				}
			}
		}
	}
}

func attention(command string) {
	switch command {
	case "blink":
		for i := 0; i < Config.Global.Software.RemoteControl.MQTT.Settings.MQTTAttentionBlinkTimes; i++ {
			GPIOOutPin("attention", "on")
			time.Sleep(time.Duration(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTAttentionBlinkmsecs) * time.Millisecond)
			GPIOOutPin("attention", "off")
			time.Sleep(time.Duration(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTAttentionBlinkmsecs) * time.Millisecond)
		}
	case "on":
		GPIOOutPin("attention", "on")
	case "off":
		GPIOOutPin("attention", "off")
	}
}

func relay(command string, no string) {
	number := no
	checkno, err := strconv.Atoi(no)
	if err != nil || checkno == 0 || checkno > 2 {
		return
	}
	switch command {
	case "pulse":
		GPIOOutPin("relay"+number, "pulse")
	case "on":
		GPIOOutPin("relay"+number, "on")
	case "off":
		GPIOOutPin("relay"+number, "off")
	}
}
