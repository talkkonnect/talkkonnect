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

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/talkkonnect/colog"
	goled "github.com/talkkonnect/go-oled-i2c"
	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	"github.com/talkkonnect/sa818"
	"golang.org/x/sys/unix"
)

type ConfigStruct struct {
	XMLName  xml.Name `xml:"document"`
	Accounts struct {
		Account []struct {
			Name             string `xml:"name,attr"`
			Default          bool   `xml:"default,attr"`
			ServerAndPort    string `xml:"serverandport"`
			UserName         string `xml:"username"`
			Password         string `xml:"password"`
			Insecure         bool   `xml:"insecure"`
			Register         bool   `xml:"register"`
			Certificate      string `xml:"certificate"`
			Channel          string `xml:"channel"`
			Ident            string `xml:"ident"`
			Listentochannels struct {
				ChannelNames []string `xml:"channel"`
			} `xml:"listentochannels"`
			TokensEnabled bool `xml:"enabled,attr"`
			Tokens        struct {
				Token []string `xml:"token"`
			} `xml:"tokens"`
			Voicetargets struct {
				ID []struct {
					Value     uint32 `xml:"value,attr"`
					IsCurrent bool   `xml:"iscurrent"`
					Users     struct {
						User []string `xml:"user"`
					} `xml:"users"`
					Channels struct {
						Channel []struct {
							Name      string `xml:"name"`
							Recursive bool   `xml:"recursive"`
							Links     bool   `xml:"links"`
							Group     string `xml:"group"`
						} `xml:"channel"`
					} `xml:"channels"`
				} `xml:"id"`
			} `xml:"voicetargets"`
		} `xml:"account"`
	} `xml:"accounts"`
	Global struct {
		Software struct {
			Settings struct {
				SingleInstance          bool          `xml:"singleinstance"`
				OutputDevice            string        `xml:"outputdevice"`
				OutputDeviceShort       string        `xml:"outputdeviceshort"`
				OutputVolControlDevice  string        `xml:"outputvolcontroldevice"`
				OutputMuteControlDevice string        `xml:"outputmutecontroldevice"`
				InputDevice             string        `xml:"inputdevice"`
				LogFilenameAndPath      string        `xml:"logfilenameandpath"`
				Logging                 string        `xml:"logging"`
				Loglevel                string        `xml:"loglevel"`
				CancellableStream       bool          `xml:"cancellablestream"`
				StreamOnStart           bool          `xml:"streamonstart"`
				StreamOnStartAfter      time.Duration `xml:"streamonstartafter"`
				StreamSendMessage       bool          `xml:"streamsendmessage"`
				TXOnStart               bool          `xml:"txonstart"`
				TXOnStartAfter          time.Duration `xml:"txonstartafter"`
				RepeatTXTimes           int           `xml:"repeattxtimes"`
				RepeatTXDelay           time.Duration `xml:"repeattxdelay"`
				SimplexWithMute         bool          `xml:"simplexwithmute"`
				TxCounter               bool          `xml:"txcounter"`
				NextServerIndex         int           `xml:"nextserverindex"`
				TXLockOut               bool          `xml:"txlockout"`
				ListenToChannelsOnStart bool          `xml:"listentochannelsonstart"`
			} `xml:"settings"`
			RemoteSSHConsole struct {
				Enabled   bool   `xml:"enabled,attr"`
				Username  string `xml:"username"`
				Password  string `xml:"password"`
				IDRSAFile string `xml:"idrsafile"`
				Listen    string `xml:"listen"`
			} `xml:"remotesshconsole"`
			AutoProvisioning struct {
				Enabled      bool   `xml:"enabled,attr"`
				TkID         string `xml:"tkid"`
				URL          string `xml:"url"`
				SaveFilePath string `xml:"savefilepath"`
				SaveFilename string `xml:"savefilename"`
			} `xml:"autoprovisioning"`
			Beacon struct {
				Enabled                bool    `xml:"enabled,attr"`
				BeaconTimerSecs        int     `xml:"beacontimersecs"`
				BeaconFileAndPath      string  `xml:"beaconfileandpath"`
				LocalPlay              bool    `xml:"localplay"`
				LocalVolume            int     `xml:"localvolume"`
				GPIOEnabled            bool    `xml:"gpioenabled"`
				GPIOName               string  `xml:"gpioname"`
				Playintostream         bool    `xml:"playintostream"`
				BeaconVolumeIntoStream float32 `xml:"beaconvolumeintostream"`
			} `xml:"beacon"`
			TTS struct {
				Enabled     bool   `xml:"enabled,attr"`
				Volumelevel int    `xml:"volumelevel"`
				Language    string `xml:"language,attr"`
				Sound       []struct {
					Action   string `xml:"action,attr"`
					File     string `xml:"file,attr"`
					Blocking bool   `xml:"blocking,attr"`
					Enabled  bool   `xml:"enabled,attr"`
				} `xml:"sound"`
			} `xml:"tts"`
			SMTP struct {
				Enabled       bool   `xml:"enabled,attr"`
				Username      string `xml:"username"`
				Password      string `xml:"password"`
				Receiver      string `xml:"receiver"`
				Subject       string `xml:"subject"`
				Message       string `xml:"message"`
				GpsDateTime   bool   `xml:"gpsdatetime"`
				GpsLatLong    bool   `xml:"gpslatlong"`
				GoogleMapsURL bool   `xml:"googlemapsurl"`
			} `xml:"smtp"`
			Sounds struct {
				Sound []struct {
					Event    string `xml:"event,attr"`
					File     string `xml:"file,attr"`
					Volume   string `xml:"volume,attr"`
					Blocking bool   `xml:"blocking,attr"`
					Enabled  bool   `xml:"enabled,attr"`
				} `xml:"sound"`
				Input struct {
					Enabled bool `xml:"enabled,attr"`
					Sound   []struct {
						Event   string `xml:"event,attr"`
						File    string `xml:"file,attr"`
						Enabled bool   `xml:"enabled,attr"`
					} `xml:"sound"`
				} `xml:"input"`
				RepeaterTone struct {
					Enabled         bool    `xml:"enabled,attr"`
					ToneFrequencyHz int     `xml:"tonefrequencyhz"`
					ToneDurationSec float32 `xml:"tonedurationsec"`
				} `xml:"repeatertone"`
			} `xml:"sounds"`
			RemoteControl struct {
				XMLName xml.Name `xml:"remotecontrol"`
				HTTP    struct {
					Enabled    bool   `xml:"enabled,attr"`
					ListenPort string `xml:"listenport,attr"`
					Command    []struct {
						Action        string `xml:"action,attr"`
						Funcname      string `xml:"funcname,attr"`
						Funcparamname string `xml:"funcparamname,attr"`
						Message       string `xml:"message,attr"`
						Enabled       bool   `xml:"enabled,attr"`
					} `xml:"command"`
				} `xml:"http"`
				MQTT struct {
					Enabled  bool `xml:"enabled,attr"`
					Settings struct {
						MQTTEnabled             bool   `xml:"enabled,attr"`
						MQTTSubTopic            string `xml:"mqttsubtopic"`
						MQTTPubTopic            string `xml:"mqttpubtopic"`
						MQTTBroker              string `xml:"mqttbroker"`
						MQTTPassword            string `xml:"mqttpassword"`
						MQTTUser                string `xml:"mqttuser"`
						MQTTId                  string `xml:"mqttid"`
						MQTTCleansess           bool   `xml:"cleansess"`
						MQTTQos                 byte   `xml:"qos"`
						MQTTNum                 int    `xml:"num"`
						MQTTPayload             string `xml:"payload"`
						MQTTAction              string `xml:"action"`
						MQTTStore               string `xml:"store"`
						MQTTRetained            bool   `xml:"retained"`
						MQTTAttentionBlinkTimes int    `xml:"attentionblinktimes"`
						MQTTAttentionBlinkmsecs int    `xml:"attentionblinkmsecs"`
						Pubpayload              struct {
							Mqtt []struct {
								Item    string `xml:"item,attr"`
								Payload string `xml:"payload,attr"`
								Enabled bool   `xml:"enabled,attr"`
							} `xml:"mqtt"`
						} `xml:"pubpayload"`
					} `xml:"settings"`
					Commands struct {
						Command []struct {
							Action  string `xml:"action,attr"`
							Message string `xml:"message,attr"`
							Enabled bool   `xml:"enabled,attr"`
						} `xml:"command"`
					} `xml:"commands"`
				} `xml:"mqtt"`
			}
			PrintVariables struct {
				PrintAccount            bool   `xml:"printaccount"`
				PrintSystemSettings     bool   `xml:"printsystemsettings"`
				PrintRemoteSSHConsole   bool   `xml:"printremotesshconsole"`
				PrintProvisioning       bool   `xml:"printprovisioning"`
				PrintBeacon             bool   `xml:"printbeacon"`
				PrintTTS                bool   `xml:"printtts"`
				PrintSMTP               bool   `xml:"printsmtp"`
				PrintSounds             bool   `xml:"printsounds"`
				PrintHTTPAPI            bool   `xml:"printhttpapi"`
				PrintMQTT               bool   `xml:"printmqtt"`
				PrintTTSMessages        bool   `xml:"printttsmessages"`
				PrintIgnoreUser         bool   `xml:"printignoreuser"`
				PrintHardware           bool   `xml:"printhardware"`
				PrintGPIOExpander       bool   `xml:"printgpioexpander"`
				PrintMax7219            bool   `xml:"printmax7219"`
				PrintPins               bool   `xml:"printpins"`
				PrintRotary             bool   `xml:"printrotary"`
				PrintPulse              bool   `xml:"printpulse"`
				PrintVolumeButtonStep   bool   `xml:"printvolumebuttonstep"`
				PrintHeartBeat          bool   `xml:"printheartbeat"`
				PrintComment            bool   `xml:"printcomment"`
				PrintLCD                bool   `xml:"printlcd"`
				PrintOLED               bool   `xml:"printoled"`
				PrintGPS                bool   `xml:"printgps"`
				PrintTraccar            bool   `xml:"printtraccar"`
				PrintPanic              bool   `xml:"printpanic"`
				PrintUSBKeyboard        bool   `xml:"printusbkeyboard"`
				PrintAudioRecord        bool   `xml:"printaudiorecord"`
				PrintKeyboardMap        bool   `xml:"printkeyboardmap"`
				PrintRadioModule        bool   `xml:"printradiomodule"`
				PrintMultimedia         bool   `xml:"printmultimedia"`
				Printlistentochannels   string `xml:"printlistentochannels"`
				PrintMemoryChannels     bool   `xml:"printmemorychannels"`
				PrintPresetVoiceTargets bool   `xml:"printpresetvoicetargets"`
			} `xml:"printvariables"`
			TTSMessages struct {
				Enabled           bool   `xml:"enabled,attr"`
				TTSLanguage       string `xml:"ttslanguage"`
				TTSMessageFromTag bool   `xml:"ttsmessagefromtag"`
				TTSTone           struct {
					ToneEnabled bool   `xml:"enabled,attr"`
					ToneFile    string `xml:"file,attr"`
					ToneVolume  int    `xml:"volume,attr"`
				} `xml:"ttstone"`
				Blocking              bool    `xml:"localblocking"`
				TTSSoundDirectory     string  `xml:"ttssounddirectory"`
				LocalPlay             bool    `xml:"localplay"`
				PlayIntoStream        bool    `xml:"playintostream"`
				SpeakVolumeIntoStream int     `xml:"speakvolumeintostream"`
				PlayVolumeIntoStream  float32 `xml:"playvolumeintostream"`
				GPIO                  struct {
					Name    string `xml:"name,attr"`
					Enabled bool   `xml:"enabled,attr"`
				} `xml:"gpio"`
				PreDelay struct {
					Value   time.Duration `xml:"value,attr"`
					Enabled bool          `xml:"enabled,attr"`
				} `xml:"predelay"`
				PostDelay struct {
					Value   time.Duration `xml:"value,attr"`
					Enabled bool          `xml:"enabled,attr"`
				} `xml:"postdelay"`
			} `xml:"ttsmessages"`
			IgnoreUser struct {
				IgnoreUserEnabled bool   `xml:"enabled,attr"`
				IgnoreUserRegex   string `xml:"ignoreuserregex"`
			} `xml:"ignoreuser"`
			MemoryChannels struct {
				Enabled bool `xml:"enabled,attr"`
				Channel []struct {
					GPIOName    string `xml:"gpioname,attr"`
					ChannelName string `xml:"channelname,attr"`
					Enabled     bool   `xml:"enabled,attr"`
				} `xml:"channel"`
			} `xml:"memorychannels"`
			PresetVoiceTargets struct {
				Enabled        bool `xml:"enabled,attr"`
				VoiceTargetSet []struct {
					GPIOName string `xml:"gpioname,attr"`
					ID       uint32 `xml:"id,attr"`
					Enabled  bool   `xml:"enabled,attr"`
				} `xml:"voicetargetset"`
			} `xml:"presetvoicetargets"`
		} `xml:"software"`
		Hardware struct {
			TargetBoard             string        `xml:"targetboard,attr"`
			LedStripEnabled         bool          `xml:"ledstripenabled"`
			VoiceActivityTimermsecs time.Duration `xml:"voiceactivitytimermsecs"`
			IO                      struct {
				GPIOExpander struct {
					Enabled bool `xml:"enabled,attr"`
					Chip    []struct {
						ID             int   `xml:"id,attr"`
						I2Cbus         uint8 `xml:"i2cbus,attr"`
						MCP23017Device uint8 `xml:"mcp23017device,attr"`
						Enabled        bool  `xml:"enabled,attr"`
					} `xml:"chip"`
				} `xml:"gpioexpander"`
				Max7219 struct {
					Enabled         bool `xml:"enabled,attr"`
					Max7219Cascaded int  `xml:"max7219cascaded,attr"`
					SPIBus          int  `xml:"spibus,attr"`
					SPIDevice       int  `xml:"spidevice,attr"`
					Brightness      byte `xml:"brightness,attr"`
				} `xml:"max7219"`
				Pins struct {
					Pin []struct {
						Direction string `xml:"direction,attr"`
						Device    string `xml:"device,attr"`
						Name      string `xml:"name,attr"`
						PinNo     uint   `xml:"pinno,attr"`
						Type      string `xml:"type,attr"`
						ID        int    `xml:"chipid,attr"`
						Inverted  bool   `xml:"inverted,attr"`
						Enabled   bool   `xml:"enabled,attr"`
					} `xml:"pin"`
				} `xml:"pins"`
				RotaryEncoder struct {
					Enabled bool `xml:"enabled,attr"`
					Control []struct {
						Function string `xml:"function,attr"`
						Enabled  bool   `xml:"enabled,attr"`
					} `xml:"control"`
				} `xml:"rotaryencoder"`
				Pulse struct {
					Leading  time.Duration `xml:"leadingmsecs,attr"`
					Pulse    time.Duration `xml:"pulsemsecs,attr"`
					Trailing time.Duration `xml:"trailingmsecs,attr"`
				} `xml:"pulse"`
				VolumeButtonStep struct {
					VolUpStep   int `xml:"volupstep"`
					VolDownStep int `xml:"voldownstep"`
				} `xml:"volumebuttonstep"`
			} `xml:"io"`
			HeartBeat struct {
				Enabled     bool   `xml:"enabled,attr"`
				LEDPin      string `xml:"heartbeatledpin"`
				Periodmsecs int    `xml:"periodmsecs"`
				LEDOnmsecs  int    `xml:"ledonmsecs"`
				LEDOffmsecs int    `xml:"ledoffmsecs"`
			} `xml:"heartbeat"`
			Comment struct {
				CommentButtonPin  string `xml:"commentbuttonpin"`
				CommentMessageOff string `xml:"commentmessageoff"`
				CommentMessageOn  string `xml:"commentmessageon"`
			} `xml:"comment"`
			Listening struct {
				Enabled            bool   `xml:"enabled,attr"`
				ListeningButtonPin string `xml:"listeningbuttonpin"`
			} `xml:"listening"`
			LCD struct {
				Enabled               bool   `xml:"enabled,attr"`
				InterfaceType         string `xml:"lcdinterfacetype"`
				I2CAddress            uint8  `xml:"lcdi2caddress"`
				BacklightTimerEnabled bool   `xml:"lcdbacklighttimerenabled"`
				BackLightTimeoutSecs  int    `xml:"lcdbacklighttimeoutsecs"`
				BackLightLEDPin       string `xml:"lcdbacklightpin"`
				RsPin                 int    `xml:"lcdrspin"`
				EPin                  int    `xml:"lcdepin"`
				D4Pin                 int    `xml:"lcdd4pin"`
				D5Pin                 int    `xml:"lcdd5pin"`
				D6Pin                 int    `xml:"lcdd6pin"`
				D7Pin                 int    `xml:"lcdd7pin"`
			} `xml:"lcd"`
			OLED struct {
				Enabled                 bool   `xml:"enabled,attr"`
				InterfaceType           string `xml:"oledinterfacetype"`
				DisplayRows             int    `xml:"oleddisplayrows"`
				DisplayColumns          uint8  `xml:"oleddisplaycolumns"`
				DefaultI2CBus           int    `xml:"oleddefaulti2cbus"`
				DefaultI2CAddress       uint8  `xml:"oleddefaulti2caddress"`
				ScreenWidth             int    `xml:"oledscreenwidth"`
				ScreenHeight            int    `xml:"oledscreenheight"`
				CommandColumnAddressing int    `xml:"oledcommandcolumnaddressing"`
				AddressBasePageStart    int    `xml:"oledaddressbasepagestart"`
				CharLength              int    `xml:"oledcharlength"`
				StartColumn             int    `xml:"oledstartcolumn"`
			} `xml:"oled"`
			GPS struct {
				Enabled             bool   `xml:"enabled,attr"`
				Port                string `xml:"port"`
				Baud                uint   `xml:"baud"`
				TxData              string `xml:"txdata"`
				Even                bool   `xml:"even"`
				Odd                 bool   `xml:"odd"`
				Rs485               bool   `xml:"rs485"`
				Rs485HighDuringSend bool   `xml:"rs485highduringsend"`
				Rs485HighAfterSend  bool   `xml:"rs485highaftersend"`
				StopBits            uint   `xml:"stopbits"`
				DataBits            uint   `xml:"databits"`
				CharTimeOut         uint   `xml:"chartimeout"`
				MinRead             uint   `xml:"minread"`
				Rx                  bool   `xml:"rx"`
				GpsInfoVerbose      bool   `xml:"gpsinfoverbose"`
				GpsDiagSounds       bool   `xml:"gpsdiagsounds"`
				GpsDisplayShow      bool   `xml:"gpsdisplayshow"`
			} `xml:"gps"`
			Traccar struct {
				Enabled             bool   `xml:"enabled,attr"`
				Track               bool   `xml:"track"`
				ClientId            string `xml:"clientid"`
				DeviceScreenEnabled bool   `xml:"devicescreenenabled"`
				TraccarDiagSounds   bool   `xml:"traccardiagsounds"`
				TraccarDisplayShow  bool   `xml:"traccardispayshow"`
				Protocol            struct {
					Name   string `xml:"name,attr"`
					Osmand struct {
						Port      string `xml:"port,attr"`
						ServerURL string `xml:"serverurl"`
					} `xml:"osmand"`
					T55 struct {
						Port     string `xml:"port,attr"`
						ServerIP string `xml:"serverip"`
					} `xml:"t55"`
					Opengts struct {
						Port      string `xml:"port,attr"`
						ServerURL string `xml:"serverurl"`
					} `xml:"opengts"`
				} `xml:"protocol"`
			} `xml:"traccar"`
			PanicFunction struct {
				Enabled              bool    `xml:"enabled,attr"`
				FilenameAndPath      string  `xml:"filenameandpath"`
				Volume               float32 `xml:"volume"`
				Blocking             bool    `xml:"blocking,attr"`
				SendIdent            bool    `xml:"sendident"`
				Message              string  `xml:"panicmessage"`
				PMailEnabled         bool    `xml:"panicemail"`
				PEavesdropEnabled    bool    `xml:"eavesdrop"`
				RecursiveSendMessage bool    `xml:"recursivesendmessage"`
				SendGpsLocation      bool    `xml:"sendgpslocation"`
				TxLockEnabled        bool    `xml:"txlockenabled"`
				TxLockTimeOutSecs    uint    `xml:"txlocktimeoutsecs"`
				PLowProfile          bool    `xml:"lowprofile"`
			} `xml:"panicfunction"`
			USBKeyboard struct {
				Enabled         bool   `xml:"enabled,attr"`
				USBKeyboardPath string `xml:"usbkeyboarddevpath"`
				NumlockScanID   rune   `xml:"numlockscanid"`
			} `xml:"usbkeyboard"`
			AudioRecordFunction struct {
				Enabled           bool   `xml:"enabled,attr"`
				RecordOnStart     bool   `xml:"recordonstart"`
				RecordSystem      string `xml:"recordsystem"`
				RecordMode        string `xml:"recordmode"`
				RecordTimeout     int64  `xml:"recordtimeout"`
				RecordFromOutput  string `xml:"recordfromoutput"`
				RecordFromInput   string `xml:"recordfrominput"`
				RecordMicTimeout  int64  `xml:"recordmictimeout"`
				RecordSoft        string `xml:"recordsoft"`
				RecordSavePath    string `xml:"recordsavepath"`
				RecordArchivePath string `xml:"recordarchivepath"`
				RecordProfile     string `xml:"recordprofile"`
				RecordFileFormat  string `xml:"recordfileformat"`
				RecordChunkSize   string `xml:"recordchunksize"`
			} `xml:"audiorecordfunction"`
			Keyboard struct {
				Command []struct {
					Action      string `xml:"action,attr"`
					ParamName   string `xml:"paramname,attr"`
					Paramvalue  string `xml:"paramvalue,attr"`
					Enabled     bool   `xml:"enabled,attr"`
					Ttykeyboard struct {
						Scanid   rune   `xml:"scanid,attr"`
						Keylabel uint32 `xml:"keylabel,attr"`
						Enabled  bool   `xml:"enabled,attr"`
					} `xml:"ttykeyboard"`
					Usbkeyboard struct {
						Scanid   rune   `xml:"scanid,attr"`
						Keylabel uint32 `xml:"keylabel,attr"`
						Enabled  bool   `xml:"enabled,attr"`
					} `xml:"usbkeyboard"`
				} `xml:"command"`
			} `xml:"keyboard"`
			Radio struct {
				XMLName          xml.Name `xml:"radio"`
				Enabled          bool     `xml:"enabled,attr"`
				ConnectChannelID string   `xml:"connectchannelid"`
				Sa818            struct {
					Enabled   bool `xml:"enabled,attr"`
					PDEnabled bool `xml:"enabled"`
					Serial    struct {
						Enabled  bool   `xml:"enabled,attr"`
						Port     string `xml:"port"`
						Baud     uint   `xml:"baud"`
						Stopbits uint   `xml:"stopbits"`
						Databits uint   `xml:"databits"`
					} `xml:"serial"`
					Channels struct {
						Channel []struct {
							ID         string  `xml:"id,attr"`
							Name       string  `xml:"name,attr"`
							Enabled    bool    `xml:"enabled,attr"`
							ItemInList int     `xml:""`
							Bandwidth  int     `xml:"bandwidth"`
							Rxfreq     float32 `xml:"rxfreq"`
							Txfreq     float32 `xml:"txfreq"`
							Squelch    int     `xml:"squelch"`
							Ctcsstone  int     `xml:"ctcsstone"`
							Dcstone    int     `xml:"dcstone"`
							Predeemph  int     `xml:"predeemph"`
							Highpass   int     `xml:"highpass"`
							Lowpass    int     `xml:"lowpass"`
							Volume     int     `xml:"volume"`
							TXPower    string  `xml:"txpower"`
						} `xml:"channel"`
					} `xml:"channels"`
				} `xml:"sa818"`
			}
			AnalogRelays struct {
				Enabled bool `xml:"enabled,attr"`
				Zones   struct {
					Zone []struct {
						Enabled       bool   `xml:"enabled,attr"`
						Name          string `xml:"name,attr"`
						ListenChannel string `xml:"listenchannel,attr"`
						Pins          struct {
							Name []string `xml:"name"`
						} `xml:"pins"`
					} `xml:"zone"`
				} `xml:"zones"`
			} `xml:"analogrelays"`
		} `xml:"hardware"`
		Multimedia struct {
			ID []struct {
				Value   string `xml:"value,attr"`
				Enabled bool   `xml:"enabled,attr"`
				Params  struct {
					Announcementtone struct {
						File     string `xml:"file,attr"`
						Volume   int    `xml:"volume,attr"`
						Blocking bool   `xml:"blocking"`
						Enabled  bool   `xml:"enabled,attr"`
					} `xml:"announcementtone"`
					Localplay bool `xml:"localplay"`
					GPIO      struct {
						Name    string `xml:"name,attr"`
						Enabled bool   `xml:"enabled,attr"`
					} `xml:"gpio"`
					Predelay struct {
						Value   time.Duration `xml:"value,attr"`
						Enabled bool          `xml:"enabled,attr"`
					} `xml:"predelay"`
					Postdelay struct {
						Value   time.Duration `xml:"value,attr"`
						Enabled bool          `xml:"enabled,attr"`
					} `xml:"postdelay"`
					Playintostream bool `xml:"playintostream"`
					Voicetarget    bool `xml:"voicetarget"`
				} `xml:"params"`
				Media struct {
					Source []struct {
						Name     string  `xml:"name,attr"`
						File     string  `xml:"file,attr"`
						Volume   int     `xml:"volume,attr"`
						Duration float32 `xml:"duration,attr"`
						Offset   float32 `xml:"offset,attr"`
						Loop     int     `xml:"loop,attr"`
						Blocking bool    `xml:"blocking"`
						Enabled  bool    `xml:"enabled,attr"`
					} `xml:"source"`
				} `xml:"media"`
			} `xml:"id"`
		} `xml:"multimedia"`
	} `xml:"global"`
}

type VTStruct struct {
	ID []struct {
		Value     uint32
		IsCurrent bool
		Users     struct {
			User []string
		}
		Channels struct {
			Channel []struct {
				Name      string
				Recursive bool
				Links     bool
				Group     string
			}
		}
	}
}

type MemoryChannelStruct struct {
	Enabled     bool
	ChannelName string
}

type VoiceTargetStruct struct {
	Enabled bool
	ID      uint32
}

type KBStruct struct {
	Enabled    bool
	KeyLabel   uint32
	Command    string
	ParamName  string
	ParamValue string
}

type EventSoundStruct struct {
	Enabled  bool
	FileName string
	Volume   string
	Blocking bool
}

type InputEventSoundFileStruct struct {
	Event   string
	File    string
	Enabled bool
}

type streamTrackerStruct struct {
	UserID      uint32
	UserName    string
	UserSession uint32
	C           <-chan *gumble.AudioPacket
}

type talkingStruct struct {
	IsTalking  bool
	WhoTalking string
	OnChannel  string
}

type mqttPubButtonStruct struct {
	Item    string
	Payload string
	Enabled bool
}

type radioChannelsStruct struct {
	ID         string
	Name       string
	ItemInList int
	Bandwidth  int
	Rxfreq     float32
	Txfreq     float32
	Squelch    int
	Ctcsstone  int
	Dcstone    int
	Predeemph  int
	Highpass   int
	Lowpass    int
	Volume     int
}

type rotaryFunctionsStruct struct {
	Item     int
	Function string
}

// type analogZoneStruct struct {
// 	oneShot     bool
// 	lastChannel string
// }

// Generic Global Config Variables
var Config ConfigStruct
var ConfigXMLFile string
var radioChannels []radioChannelsStruct
var RotaryFunctions []rotaryFunctionsStruct

// Generic Global State Variables
var (
	KillHeartBeat           bool
	IsPlayStream            bool
	IsConnected             bool
	Streaming               bool
	HTTPServRunning         bool
	NowStreaming            bool
	InStreamTalking         bool
	InStreamSource          bool
	LCDIsDark               bool
	GPSDataChannelReceivers int
	TXLockOut               bool
	RootChannel             *gumble.Channel
	TopChannel              *gumble.Channel
	TopChannelID            uint32
	IsPlaying               bool
)

// Generic Global Counter Variables
var (
	AccountCount    int
	ConnectAttempts int
	AccountIndex    int
	GenericCounter  int
	CurrentIndex    int
	ChannelAction   string
)

// Generic Global Timer Variables
var (
	BackLightTime    = time.NewTicker(5 * time.Second)
	BackLightTimePtr = &BackLightTime
	StartTime        = time.Now()
	LastTime         = now.Unix()
	TalkedTicker     = time.NewTicker(time.Millisecond * 200)
	Talking          = make(chan talkingStruct, 10)
	BeaconTime       = time.NewTicker(100 * time.Second)
	BeaconTimePtr    = &BeaconTime
)

var (
	LcdText = [4]string{"nil", "nil", "nil", "nil"}
	//	MyLedStrip *LedStrip
	TTYKeyMap            = make(map[rune]KBStruct)
	USBKeyMap            = make(map[rune]KBStruct)
	GPIOMemoryMap        = make(map[string]MemoryChannelStruct)
	GPIOVoiceTargetMap   = make(map[string]VoiceTargetStruct)
	AccessableChannelMap = make(map[int]string)
)

// Mumble Account Settings Global Variables
var (
	Default               []bool
	Name                  []string
	Server                []string
	Username              []string
	Password              []string
	Insecure              []bool
	Register              []bool
	Certificate           []string
	Channel               []string
	Ident                 []string
	Tokens                []gumble.AccessTokens
	VT                    []VTStruct
	ChannelsList          []ChannelsListStruct
	ListenChannelNameList []string
	Accounts              int
)

// HD44780 LCD Screen Settings Golbal Variables
var (
	LCDEnabled               bool
	LCDInterfaceType         string
	LCDI2CAddress            uint8
	LCDBackLightTimerEnabled bool
	LCDBackLightTimeout      time.Duration
	LCDRSPin                 int
	LCDEPin                  int
	LCDD4Pin                 int
	LCDD5Pin                 int
	LCDD6Pin                 int
	LCDD7Pin                 int
)

// OLED Screen Settings Golbal Variables
var (
	OLEDEnabled                 bool
	OLEDInterfacetype           string
	OLEDDefaultI2cAddress       uint8
	OLEDDefaultI2cBus           int
	OLEDScreenWidth             int
	OLEDScreenHeight            int
	OLEDDisplayRows             int
	OLEDDisplayColumns          uint8
	OLEDStartColumn             int
	OLEDCharLength              int
	OLEDCommandColumnAddressing int
	OLEDAddressBasePageStart    int
	Oled                        *goled.Oled
)

// Generic Local Variables
var (
	txcounter      int
	isTx           bool
	pstream        *gumbleffmpeg.Stream
	LastSpeaker    string = ""
	RotaryFunction rotaryFunctionsStruct
)

var StreamTracker = map[uint32]streamTrackerStruct{}
var DMOSetup sa818.DMOSetupStruct

func readxmlconfig(file string, reloadxml bool) error {
	var ReConfig ConfigStruct

	xmlFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	log.Println("info: Successfully Read file " + filepath.Base(file))
	defer xmlFile.Close()

	byteValue, _ := io.ReadAll(xmlFile)

	if !reloadxml {
		err = xml.Unmarshal(byteValue, &Config)
		if err != nil {
			return fmt.Errorf(filepath.Base(file) + " " + err.Error())
		}
	} else {
		err = xml.Unmarshal(byteValue, &ReConfig)
		if err != nil {
			return fmt.Errorf(filepath.Base(file) + " " + err.Error())
		}
	}
	CheckConfigSanity(reloadxml)

	if !reloadxml {
		for _, account := range Config.Accounts.Account {
			if account.Default {
				Name = append(Name, account.Name)
				Server = append(Server, account.ServerAndPort)
				Username = append(Username, account.UserName)
				Password = append(Password, account.Password)
				Insecure = append(Insecure, account.Insecure)
				Register = append(Register, account.Register)
				Certificate = append(Certificate, account.Certificate)
				Channel = append(Channel, account.Channel)
				Ident = append(Ident, account.Ident)
				Tokens = append(Tokens, account.Tokens.Token)
				VT = append(VT, VTStruct(account.Voicetargets))
				//ListenChannelNameList = append(ListenChannelNameList, account.Listentochannels.ChannelNames...)
				AccountCount++
			}
		}
	}
	for _, kMainCommands := range Config.Global.Hardware.Keyboard.Command {
		if kMainCommands.Enabled {
			if kMainCommands.Ttykeyboard.Enabled {
				TTYKeyMap[kMainCommands.Ttykeyboard.Scanid] = KBStruct{kMainCommands.Ttykeyboard.Enabled, kMainCommands.Ttykeyboard.Keylabel, kMainCommands.Action, kMainCommands.ParamName, kMainCommands.Paramvalue}
			}
			if kMainCommands.Usbkeyboard.Enabled {
				USBKeyMap[kMainCommands.Usbkeyboard.Scanid] = KBStruct{kMainCommands.Usbkeyboard.Enabled, kMainCommands.Usbkeyboard.Keylabel, kMainCommands.Action, kMainCommands.ParamName, kMainCommands.Paramvalue}
			}

		}
	}

	for _, memoryButtonCommands := range Config.Global.Software.MemoryChannels.Channel {
		if memoryButtonCommands.Enabled {
			log.Printf("debug: Populating %v With Channel %v\n", memoryButtonCommands.GPIOName, memoryButtonCommands.ChannelName)
			GPIOMemoryMap[memoryButtonCommands.GPIOName] = MemoryChannelStruct{memoryButtonCommands.Enabled, memoryButtonCommands.ChannelName}
		}
	}

	for _, voicetargetButtonCommands := range Config.Global.Software.PresetVoiceTargets.VoiceTargetSet {
		if voicetargetButtonCommands.Enabled {
			log.Printf("debug: Populating %v With VoiceTarget ID %v\n", voicetargetButtonCommands.GPIOName, voicetargetButtonCommands.ID)
			GPIOVoiceTargetMap[voicetargetButtonCommands.GPIOName] = VoiceTargetStruct{voicetargetButtonCommands.Enabled, voicetargetButtonCommands.ID}
		}
	}

	exec, err := os.Executable()

	if err != nil {
		exec = "./talkkonnect" //Hardcode our default name
	}

	// Set our default config file path (for autoprovision)
	defaultConfPath, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		FatalCleanUp("Unable to get path for config file " + err.Error())
	}

	// Set our default logging path
	//This section is pretty unix specific.. sorry if you like windows support.
	defaultLogPath := "/tmp/" + filepath.Base(exec) + ".log" // Safe assumption as it should be writable for everyone
	// First see if we can write in our CWD and use it over /tmp
	cwd, err := os.Getwd()
	if err == nil {
		cwd, err := filepath.Abs(cwd)
		if err == nil {
			if unix.Access(cwd, unix.W_OK) == nil {
				defaultLogPath = cwd + "/" + filepath.Base(exec) + ".log"
			}
		}
	}

	// Next try a file in our config path and favor it over CWD
	if unix.Access(defaultConfPath, unix.W_OK) == nil {
		defaultLogPath = defaultConfPath + "/" + filepath.Base(exec) + ".log"
	}

	// Last, see if the system talkkonnect log exists and is writeable and do that over CWD, HOME and /tmp
	if _, err := os.Stat("/var/log/" + filepath.Base(exec) + ".log"); err == nil {
		f, err := os.OpenFile("/var/log/"+filepath.Base(exec)+".log", os.O_WRONLY, 0664)
		if err == nil {
			defaultLogPath = "/var/log/" + filepath.Base(exec) + ".log"
		}
		f.Close()
	}

	if len(Config.Global.Software.Settings.OutputDeviceShort) == 0 {
		Config.Global.Software.Settings.OutputDeviceShort = Config.Global.Software.Settings.OutputDevice
	}

	if len(Config.Global.Software.Settings.OutputVolControlDevice) == 0 {
		Config.Global.Software.Settings.OutputVolControlDevice = Config.Global.Software.Settings.OutputDevice
	}
	if len(Config.Global.Software.Settings.OutputMuteControlDevice) == 0 {
		Config.Global.Software.Settings.OutputMuteControlDevice = Config.Global.Software.Settings.OutputDevice
	}

	if strings.ToLower(Config.Global.Software.Settings.Logging) != "screen" && Config.Global.Software.Settings.LogFilenameAndPath == "" {
		Config.Global.Software.Settings.LogFilenameAndPath = defaultLogPath
	}

	if !reloadxml {
		LCDEnabled = Config.Global.Hardware.LCD.Enabled
		LCDInterfaceType = Config.Global.Hardware.LCD.InterfaceType
		LCDI2CAddress = Config.Global.Hardware.LCD.I2CAddress
		LCDBackLightTimerEnabled = Config.Global.Hardware.LCD.Enabled
		LCDBackLightTimeout = time.Duration(Config.Global.Hardware.LCD.BackLightTimeoutSecs)
		LCDRSPin = Config.Global.Hardware.LCD.RsPin
		LCDEPin = Config.Global.Hardware.LCD.EPin
		LCDD4Pin = Config.Global.Hardware.LCD.D4Pin
		LCDD5Pin = Config.Global.Hardware.LCD.D5Pin
		LCDD6Pin = Config.Global.Hardware.LCD.D6Pin
		LCDD7Pin = Config.Global.Hardware.LCD.D7Pin

		OLEDEnabled = Config.Global.Hardware.OLED.Enabled
		OLEDInterfacetype = Config.Global.Hardware.OLED.InterfaceType
		OLEDDisplayRows = Config.Global.Hardware.OLED.DisplayRows
		OLEDDisplayColumns = Config.Global.Hardware.OLED.DisplayColumns
		OLEDDefaultI2cBus = Config.Global.Hardware.OLED.DefaultI2CBus
		OLEDDefaultI2cAddress = Config.Global.Hardware.OLED.DefaultI2CAddress
		OLEDScreenWidth = Config.Global.Hardware.OLED.ScreenWidth
		OLEDScreenHeight = Config.Global.Hardware.OLED.ScreenHeight
		OLEDCommandColumnAddressing = Config.Global.Hardware.OLED.CommandColumnAddressing
		OLEDAddressBasePageStart = Config.Global.Hardware.OLED.AddressBasePageStart
		OLEDCharLength = Config.Global.Hardware.OLED.CharLength
		OLEDStartColumn = Config.Global.Hardware.OLED.StartColumn

		if Config.Global.Hardware.TargetBoard != "rpi" {
			LCDBackLightTimerEnabled = false
		}

		if Config.Global.Hardware.VoiceActivityTimermsecs == 0 {
			Config.Global.Hardware.VoiceActivityTimermsecs = 200
		}

		if Config.Global.Hardware.IO.VolumeButtonStep.VolUpStep == 0 {
			Config.Global.Hardware.IO.VolumeButtonStep.VolUpStep = +1
		}

		if Config.Global.Hardware.IO.VolumeButtonStep.VolDownStep == 0 {
			Config.Global.Hardware.IO.VolumeButtonStep.VolDownStep = -1
		}

		if OLEDEnabled {
			Oled, err = goled.BeginOled(OLEDDefaultI2cAddress, OLEDDefaultI2cBus, OLEDScreenWidth, OLEDScreenHeight, OLEDDisplayRows, OLEDDisplayColumns, OLEDStartColumn, OLEDCharLength, OLEDCommandColumnAddressing, OLEDAddressBasePageStart)
			if err != nil {
				log.Println("error: Cannot Communicate with OLED")
				OLEDEnabled = false
			}
		}
	}
	log.Println("info: Successfully loaded XML configuration file into memory")

	// Add Allowed Mutable Settings For talkkonnect upon live reloadxml config to the list below omit all other variables
	if reloadxml {
		if Config.Global.Software.Settings.Loglevel != ReConfig.Global.Software.Settings.Loglevel {
			Config.Global.Software.Settings.Loglevel = ReConfig.Global.Software.Settings.Loglevel
			switch Config.Global.Software.Settings.Loglevel {
			case "trace":
				colog.SetMinLevel(colog.LTrace)
				log.Println("info: Loglevel Set to Trace")
			case "debug":
				colog.SetMinLevel(colog.LDebug)
				log.Println("info: Loglevel Set to Debug")
			case "info":
				colog.SetMinLevel(colog.LInfo)
				log.Println("info: Loglevel Set to Info")
			case "warning":
				colog.SetMinLevel(colog.LWarning)
				log.Println("info: Loglevel Set to Warning")
			case "error":
				colog.SetMinLevel(colog.LError)
				log.Println("info: Loglevel Set to Error")
			case "alert":
				colog.SetMinLevel(colog.LAlert)
				log.Println("info: Loglevel Set to Alert")
			default:
				colog.SetMinLevel(colog.LInfo)
				log.Println("info: Default Loglevel unset in XML config automatically loglevel to Info")
			}
		}

		Config.Global.Software.Settings.CancellableStream = ReConfig.Global.Software.Settings.CancellableStream
		Config.Global.Software.Settings.StreamSendMessage = ReConfig.Global.Software.Settings.StreamSendMessage
		Config.Global.Software.Settings.RepeatTXTimes = ReConfig.Global.Software.Settings.RepeatTXTimes
		Config.Global.Software.Settings.RepeatTXDelay = ReConfig.Global.Software.Settings.RepeatTXDelay
		Config.Global.Software.Settings.SimplexWithMute = ReConfig.Global.Software.Settings.SimplexWithMute
		Config.Global.Software.Beacon = ReConfig.Global.Software.Beacon
		Config.Global.Software.TTS = ReConfig.Global.Software.TTS
		Config.Global.Software.Sounds = ReConfig.Global.Software.Sounds
		Config.Global.Software.RemoteControl.HTTP.Enabled = ReConfig.Global.Software.RemoteControl.HTTP.Enabled
		Config.Global.Software.RemoteControl.HTTP.Command = ReConfig.Global.Software.RemoteControl.HTTP.Command
		Config.Global.Software.RemoteControl.MQTT.Commands.Command = ReConfig.Global.Software.RemoteControl.MQTT.Commands.Command
		Config.Global.Software.PrintVariables = ReConfig.Global.Software.PrintVariables
		Config.Global.Software.TTSMessages = ReConfig.Global.Software.TTSMessages
		Config.Global.Software.IgnoreUser = ReConfig.Global.Software.IgnoreUser
		Config.Global.Hardware.PanicFunction = ReConfig.Global.Hardware.PanicFunction
		Config.Global.Hardware.Keyboard.Command = ReConfig.Global.Hardware.Keyboard.Command
		Config.Global.Multimedia = ReConfig.Global.Multimedia
		//ReConfig.Accounts.Account[0].Listentochannels

	}
	return nil
}

func printxmlconfig() {

	if !Config.Global.Software.PrintVariables.PrintAccount {
		log.Println("info: ---------- Account Information -------- SKIPPED ")
	} else {
		log.Println("info: ---------- Account Info ---------- ")
		for index, account := range Config.Accounts.Account {
			if account.Default {
				var AcctIsDefault string = "x"
				if Server[AccountIndex] == account.ServerAndPort && Username[AccountIndex] == account.UserName {
					AcctIsDefault = "✓"
				}
				log.Printf("info: %v Account %v Name %v Enabled %v \n", AcctIsDefault, index, account.Name, account.Default)
				log.Printf("info: %v Server:Port     %v \n", AcctIsDefault, account.ServerAndPort)
				log.Printf("info: %v Username %v Password %v \n", AcctIsDefault, account.UserName, account.Password)
				log.Printf("info: %v Insecure %v Register %v \n", AcctIsDefault, account.Insecure, account.Register)
				log.Printf("info: %v Certificate      %v \n", AcctIsDefault, account.Certificate)
				log.Printf("info: %v Channel          %v \n", AcctIsDefault, account.Channel)
				log.Printf("info: %v Ident            %v \n", AcctIsDefault, account.Ident)
				log.Printf("info: %v ListentoChannels %v \n", AcctIsDefault, account.Listentochannels)
				log.Printf("info: %v Tokens           %v \n", AcctIsDefault, account.Tokens)
				log.Printf("info: %v VoiceTargets     %v \n", AcctIsDefault, account.Voicetargets)
			}
		}
	}

	if !Config.Global.Software.PrintVariables.PrintSystemSettings {
		log.Println("info: -------- System Settings -------- SKIPPED ")
	} else {
		log.Println("info: -------- System Settings -------- ")
		log.Println("info: Single Instance                  ", Config.Global.Software.Settings.SingleInstance)
		log.Println("info: Output Device                    ", Config.Global.Software.Settings.OutputDevice)
		log.Println("info: Output Device(Short)             ", Config.Global.Software.Settings.OutputDeviceShort)
		log.Println("info: Output Vol Control Device        ", Config.Global.Software.Settings.OutputVolControlDevice)
		log.Println("info: Output Mute Control Device       ", Config.Global.Software.Settings.OutputMuteControlDevice)
		log.Println("info: Input Device                     ", Config.Global.Software.Settings.InputDevice)
		log.Println("info: Log File                         ", Config.Global.Software.Settings.LogFilenameAndPath)
		log.Println("info: Logging                          ", Config.Global.Software.Settings.Logging)
		log.Println("info: Loglevel                         ", Config.Global.Software.Settings.Loglevel)
		log.Println("info: CancellableStream                ", fmt.Sprintf("%t", Config.Global.Software.Settings.CancellableStream))
		log.Println("info: StreamOnStart                    ", fmt.Sprintf("%t", Config.Global.Software.Settings.StreamOnStart))
		log.Println("info: StreamOnStartAfter               ", fmt.Sprintf("%v", Config.Global.Software.Settings.StreamOnStartAfter))
		log.Println("info: TXOnStart                        ", fmt.Sprintf("%t", Config.Global.Software.Settings.TXOnStart))
		log.Println("info: TXOnStartAfter                   ", fmt.Sprintf("%v", Config.Global.Software.Settings.TXOnStartAfter))
		log.Println("info: RepeatTXTimes                    ", fmt.Sprintf("%v", Config.Global.Software.Settings.RepeatTXTimes))
		log.Println("info: RepeatTXDelay                    ", fmt.Sprintf("%v", Config.Global.Software.Settings.RepeatTXDelay))
		log.Println("info: SimplexWithMute                  ", fmt.Sprintf("%t", Config.Global.Software.Settings.SimplexWithMute))
		log.Println("info: TxCounter                        ", fmt.Sprintf("%t", Config.Global.Software.Settings.TxCounter))
		log.Println("info: NextServerIndex                  ", fmt.Sprintf("%v", Config.Global.Software.Settings.NextServerIndex))
		log.Println("info: TXLockOut                        ", fmt.Sprintf("%t", Config.Global.Software.Settings.TXLockOut))
		log.Println("info: ListenToChannelOnStart           ", fmt.Sprintf("%t", Config.Global.Software.Settings.ListenToChannelsOnStart))
	}

	if !Config.Global.Software.PrintVariables.PrintRemoteSSHConsole {
		log.Println("info: -------- Remote SSH Console Settings -------- SKIPPED ")
	} else {
		log.Println("info: -------- Remote SSH Console Settings -------- ")
		log.Println("info: Enabled      ", Config.Global.Software.RemoteSSHConsole.Enabled)
		log.Println("info: Username     ", Config.Global.Software.RemoteSSHConsole.Username)
		log.Println("info: Password     ", Config.Global.Software.RemoteSSHConsole.Password)
		log.Println("info: IDRSAFile    ", Config.Global.Software.RemoteSSHConsole.IDRSAFile)
		log.Println("info: Listen       ", Config.Global.Software.RemoteSSHConsole.Listen)
	}

	if !Config.Global.Software.PrintVariables.PrintProvisioning {
		log.Println("info: --------   AutoProvisioning   --------- SKIPPED ")
	} else {
		log.Println("info: --------   AutoProvisioning   --------- ")
		log.Println("info: AutoProvisioning Enabled    " + fmt.Sprintf("%t", Config.Global.Software.AutoProvisioning.Enabled))
		log.Println("info: Talkkonned ID (tkid)        " + Config.Global.Software.AutoProvisioning.TkID)
		log.Println("info: AutoProvisioning Server URL " + Config.Global.Software.AutoProvisioning.URL)
		log.Println("info: Config Local Path           " + Config.Global.Software.AutoProvisioning.SaveFilePath)
		log.Println("info: Config Local Filename       " + Config.Global.Software.AutoProvisioning.SaveFilename)
	}

	if !Config.Global.Software.PrintVariables.PrintBeacon {
		log.Println("info: --------   Beacon   --------- SKIPPED ")
	} else {
		log.Println("info: --------  Beacon   --------- ")
		log.Println("info: Beacon Enabled            " + fmt.Sprintf("%t", Config.Global.Software.Beacon.Enabled))
		log.Println("info: Beacon Filename & Path    " + Config.Global.Software.Beacon.BeaconFileAndPath)
		log.Println("info: Beacon Time (Secs)        " + fmt.Sprintf("%v", Config.Global.Software.Beacon.BeaconTimerSecs))
		log.Println("info: Beacon Volume Into Stream " + fmt.Sprintf("%v", Config.Global.Software.Beacon.BeaconVolumeIntoStream))
		log.Println("info: Beacon GPIOName           " + fmt.Sprintf("%v", Config.Global.Software.Beacon.GPIOName))
		log.Println("info: Local Volume              " + fmt.Sprintf("%v", Config.Global.Software.Beacon.LocalPlay))
		log.Println("info: Beacon Play Into Stream   " + fmt.Sprintf("%v", Config.Global.Software.Beacon.Playintostream))
	}

	if !Config.Global.Software.PrintVariables.PrintTTS {
		log.Println("info: --------   TTS  -------- SKIPPED ")
	} else {
		log.Println("info: -------- TTS  -------- ")
		log.Println("info: Enabled      " + fmt.Sprintf("%t", Config.Global.Software.TTS.Enabled))
		log.Println("info: Volume Level ", Config.Global.Software.TTS.Volumelevel)
		log.Println("info: Language     ", Config.Global.Software.TTS.Language)
		for _, tts := range Config.Global.Software.TTS.Sound {
			log.Printf("%+v\n", tts)
		}
	}

	if !Config.Global.Software.PrintVariables.PrintSMTP {
		log.Println("info: --------   Gmail SMTP Settings  -------- SKIPPED ")
	} else {
		log.Println("info: --------  Gmail SMTP Settings  -------- ")
		log.Println("info: Email Enabled   " + fmt.Sprintf("%t", Config.Global.Software.SMTP.Enabled))
		log.Println("info: Username        " + Config.Global.Software.SMTP.Username)
		log.Println("info: Password        " + Config.Global.Software.SMTP.Password)
		log.Println("info: Receiver        " + Config.Global.Software.SMTP.Receiver)
		log.Println("info: Subject         " + Config.Global.Software.SMTP.Subject)
		log.Println("info: Message         " + Config.Global.Software.SMTP.Message)
		log.Println("info: GPS Date/Time   " + fmt.Sprintf("%t", Config.Global.Software.SMTP.GpsDateTime))
		log.Println("info: GPS Lat/Long    " + fmt.Sprintf("%t", Config.Global.Software.SMTP.GpsLatLong))
		log.Println("info: Google Maps URL " + fmt.Sprintf("%t", Config.Global.Software.SMTP.GoogleMapsURL))
	}

	if !Config.Global.Software.PrintVariables.PrintSounds {
		log.Println("info: ------------ Sounds  ------------------ SKIPPED ")

	} else {
		log.Println("info: ------------- Sounds  ------------------ ")
		for _, sounds := range Config.Global.Software.Sounds.Sound {
			log.Printf("info: |Event=%v |File=%v |Volume=%v |Blocking=%v |Enabled=%v\n", sounds.Event, sounds.File, sounds.Volume, sounds.Blocking, sounds.Enabled)
		}
		log.Println("info: Input Enabled         " + fmt.Sprintf("%t", Config.Global.Software.Sounds.Input.Enabled))
		for _, input := range Config.Global.Software.Sounds.Input.Sound {
			log.Printf("info: |Event=%v |File=%v |Enabled=%v\n", input.Event, input.File, input.Enabled)
		}
		log.Println("info: Repeater Tone Enabled      " + fmt.Sprintf("%t", Config.Global.Software.Sounds.RepeaterTone.Enabled))
		log.Println("info: Repeater Tone Freq (Hz)    ", Config.Global.Software.Sounds.RepeaterTone.ToneFrequencyHz)
		log.Println("info: Repeater Tone Duration (s) ", Config.Global.Software.Sounds.RepeaterTone.ToneDurationSec)
	}

	if !Config.Global.Software.PrintVariables.PrintHTTPAPI {
		log.Println("info: ------------ HTTP API  ----------------- SKIPPED ")
	} else {
		log.Println("info: ------------ HTTP API  ----------------- ")
		log.Println("info: HTTP API Enabled ", Config.Global.Software.RemoteControl.HTTP.Enabled)
		log.Println("info: HTTP API Listen Port ", Config.Global.Software.RemoteControl.HTTP.ListenPort)
		for _, command := range Config.Global.Software.RemoteControl.HTTP.Command {
			log.Printf("info: Enabled=%v Action=%v Name=%v Param=%v Message=%v\n", command.Enabled, command.Action, command.Funcname, command.Funcparamname, command.Message)
		}
	}

	if !Config.Global.Software.PrintVariables.PrintMQTT {
		log.Println("info: ------------ MQTT Function ------- SKIPPED ")
	} else {
		log.Println("info: ------------ MQTT Function -------------- ")
		log.Println("info: Enabled             " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Enabled))
		log.Println("info: Subscibe Topic      " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTSubTopic))
		log.Println("info: Publish  Topic      " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTPubTopic))
		log.Println("info: Broker              " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTBroker))
		log.Println("info: Password            " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTPassword))
		log.Println("info: Id                  " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTId))
		log.Println("info: Cleansess           " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTCleansess))
		log.Println("info: Qos                 " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTQos))
		log.Println("info: Num                 " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTNum))
		log.Println("info: Payload             " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTPayload))
		log.Println("info: Action              " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTAction))
		log.Println("info: Store               " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTStore))
		log.Println("info: AttentionBlinkTimes " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTAttentionBlinkTimes))
		log.Println("info: AttentionBlinkmsecs " + fmt.Sprintf("%v", Config.Global.Software.RemoteControl.MQTT.Settings.MQTTAttentionBlinkmsecs))
		for _, command := range Config.Global.Software.RemoteControl.MQTT.Commands.Command {
			log.Printf("info: Enabled=%v Action=%v Message=%v\n", command.Enabled, command.Action, command.Message)
		}
	}

	if !Config.Global.Software.PrintVariables.PrintTTSMessages {
		log.Println("info: ------------ TTSMessages Function ------- SKIPPED ")
	} else {
		log.Println("info: ------------ TTSMessages Function -------------- ")
		log.Println("info: Enabled                      " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.Enabled))
		log.Println("info: LocalPlay                    " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.LocalPlay))
		log.Println("info: Play Into Stream             " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.PlayIntoStream))
		log.Println("info: TTS Speak Volume Into Stream " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.SpeakVolumeIntoStream))
		log.Println("info: TTS Play Volume Into Stream  " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.PlayVolumeIntoStream))
		log.Println("info: TTSLanguage                  " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.TTSLanguage))
		log.Println("info: TTSSoundDirectory            " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.TTSSoundDirectory))
		log.Println("info: TTSAnnouncementTone Enabled  " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.TTSTone.ToneEnabled))
		log.Println("info: TTSAnnouncementTone File     " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.TTSTone.ToneFile))
		log.Println("info: TTSMessageFromTag            " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.TTSMessageFromTag))
		log.Println("info: TTSGPIOEnabled               " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.GPIO.Enabled))
		log.Println("info: TTSGPIOName                  " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.GPIO.Name))
		log.Println("info: TTSPreDelay                  " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.PreDelay))
		log.Println("info: TTSPostDelay                 " + fmt.Sprintf("%v", Config.Global.Software.TTSMessages.PreDelay))
	}

	if !Config.Global.Software.PrintVariables.PrintIgnoreUser {
		log.Println("info: ------------ IgnoreUserRegex Function ------- SKIPPED ")
	} else {
		log.Println("info: ------------ IgnoreUserRegex Function -------------- ")
		log.Println("info: Enabled             " + fmt.Sprintf("%v", Config.Global.Software.IgnoreUser.IgnoreUserEnabled))
		log.Println("info: IgnoreUserRegex     " + fmt.Sprintf("%v", Config.Global.Software.IgnoreUser.IgnoreUserRegex))
	}

	if !Config.Global.Software.PrintVariables.PrintHardware {
		log.Println("info: ------------  Hardware Settings -------------- SKIPPED")
	} else {
		log.Println("info: ------------  Hardware Settings -------------- ")
		log.Println("info: Target Board                 " + fmt.Sprintf("%v", Config.Global.Hardware.TargetBoard))
		log.Println("info: LED Strip Enabled            " + fmt.Sprintf("%v", Config.Global.Hardware.LedStripEnabled))
		log.Println("info: VoiceActivity LED Timer (ms) " + fmt.Sprintf("%v", Config.Global.Hardware.VoiceActivityTimermsecs))
	}

	if !Config.Global.Software.PrintVariables.PrintGPIOExpander {
		log.Println("info: ------------  GPIO Expander -------------- SKIPPED")
	} else {
		log.Println("info: ------------  GPIO Expander -------------- ")
		log.Println("info: GPIO Expander Enabled        " + fmt.Sprintf("%v", Config.Global.Hardware.IO.GPIOExpander.Enabled))
		if Config.Global.Hardware.IO.GPIOExpander.Enabled {
			for _, gpioexpander := range Config.Global.Hardware.IO.GPIOExpander.Chip {
				log.Printf("info: ID=%v I2CBus=%v MCP23017Device=%v Enabled=%v\n", gpioexpander.ID, gpioexpander.I2Cbus, gpioexpander.MCP23017Device, gpioexpander.Enabled)
			}
		}
	}

	if !Config.Global.Software.PrintVariables.PrintMax7219 {
		log.Println("info: ------------  Max7219 -------------- SKIPPPED")
	} else {
		log.Println("info: ------------  Max7219 -------------- ")
		log.Println("info: Enabled    " + fmt.Sprintf("%v", Config.Global.Hardware.IO.Max7219.Enabled))
		log.Println("info: Cascaded   " + fmt.Sprintf("%v", Config.Global.Hardware.IO.Max7219.Max7219Cascaded))
		log.Println("info: SPIBus     " + fmt.Sprintf("%v", Config.Global.Hardware.IO.Max7219.SPIBus))
		log.Println("info: SPIDevice  " + fmt.Sprintf("%v", Config.Global.Hardware.IO.Max7219.SPIDevice))
		log.Println("info: Brightness " + fmt.Sprintf("%v", Config.Global.Hardware.IO.Max7219.Brightness))
	}

	if !Config.Global.Software.PrintVariables.PrintPins {
		log.Println("info: ------------  PINS -------------- SKIPPED")
	} else {
		log.Println("info: ------------  PINS -------------- ")
		for _, pins := range Config.Global.Hardware.IO.Pins.Pin {
			log.Printf("info: Direction=%v Device%v Name=%v PinNo=%v Type=%v ID=%v Inverted=%v Enabled=%v\n", pins.Direction, pins.Device, pins.Name, pins.PinNo, pins.Type, pins.ID, pins.Inverted, pins.Enabled)
		}
	}

	if !Config.Global.Software.PrintVariables.PrintRotary {
		log.Println("info: ------------  PINS -------------- SKIPPED")
	} else {
		log.Println("info: ------------  Rotary -------------- ")
		for _, control := range Config.Global.Hardware.IO.RotaryEncoder.Control {
			log.Printf("info: Enabled=%v Fuction=%v\n", control.Enabled, control.Function)
		}
	}

	if !Config.Global.Software.PrintVariables.PrintPulse {
		log.Println("info: ------------  Pulse -------------- SKIPPED")
	} else {
		log.Println("info: ------------  Pulse -------------- ")
		log.Println("info: Leading  (ms) " + fmt.Sprintf("%v", Config.Global.Hardware.IO.Pulse.Leading))
		log.Println("info: Pulse    (ms) " + fmt.Sprintf("%v", Config.Global.Hardware.IO.Pulse.Pulse))
		log.Println("info: Trailing (ms) " + fmt.Sprintf("%v", Config.Global.Hardware.IO.Pulse.Trailing))
	}

	if !Config.Global.Software.PrintVariables.PrintVolumeButtonStep {
		log.Println("info: ------------  Volume Step -------------- SKIPPED")
	} else {
		log.Println("info: ------------  Volume Step -------------- ")
		log.Println("info: Vol Up   Step " + fmt.Sprintf("%v", Config.Global.Hardware.IO.VolumeButtonStep.VolUpStep))
		log.Println("info: Vol Down Step " + fmt.Sprintf("%v", Config.Global.Hardware.IO.VolumeButtonStep.VolDownStep))
	}

	if !Config.Global.Software.PrintVariables.PrintHeartBeat {
		log.Println("info: ---------- HEARTBEAT -------------------- SKIPPED ")
	} else {
		log.Println("info: ---------- HEARTBEAT -------------------- ")
		log.Println("info: HeartBeat Enabled " + fmt.Sprintf("%v", Config.Global.Hardware.HeartBeat.Enabled))
		log.Println("info: Period  mSecs     " + fmt.Sprintf("%v", Config.Global.Hardware.HeartBeat.Periodmsecs))
		log.Println("info: Led On  mSecs     " + fmt.Sprintf("%v", Config.Global.Hardware.HeartBeat.LEDOnmsecs))
		log.Println("info: Led Off mSecs     " + fmt.Sprintf("%v", Config.Global.Hardware.HeartBeat.LEDOffmsecs))
	}

	if !Config.Global.Software.PrintVariables.PrintComment {
		log.Println("info: ------------ Comment  ------------------- SKIPPED ")
	} else {
		log.Println("info: ------------ Comment  ------------------- ")
		log.Println("info: Comment Button Pin            " + fmt.Sprintf("%v", CommentButtonPin))
		log.Println("info: Comment Message State 1 (off) " + fmt.Sprintf("%v", Config.Global.Hardware.Comment.CommentMessageOff))
		log.Println("info: Comment Message State 2 (on)  " + fmt.Sprintf("%v", Config.Global.Hardware.Comment.CommentMessageOn))
	}

	if !Config.Global.Software.PrintVariables.PrintLCD {
		log.Println("info: ------------ LCD  ----------------------- SKIPPED ")
	} else {
		log.Println("info: ------------ LCD HD44780 ----------------------- ")
		log.Println("info: LCDEnabled               " + fmt.Sprintf("%v", LCDEnabled))
		log.Println("info: LCDInterfaceType         " + fmt.Sprintf("%v", LCDInterfaceType))
		log.Println("info: Lcd I2C Address          " + fmt.Sprintf("%x", LCDI2CAddress))
		log.Println("info: Back Light Timer Enabled " + fmt.Sprintf("%t", LCDBackLightTimerEnabled))
		log.Println("info: Back Light Timer Timeout " + fmt.Sprintf("%v", LCDBackLightTimeout))
		log.Println("info: RS Pin " + fmt.Sprintf("%v", LCDRSPin))
		log.Println("info: E  Pin " + fmt.Sprintf("%v", LCDEPin))
		log.Println("info: D4 Pin " + fmt.Sprintf("%v", LCDD4Pin))
		log.Println("info: D5 Pin " + fmt.Sprintf("%v", LCDD5Pin))
		log.Println("info: D6 Pin " + fmt.Sprintf("%v", LCDD6Pin))
		log.Println("info: D7 Pin " + fmt.Sprintf("%v", LCDD7Pin))
	}

	if !Config.Global.Software.PrintVariables.PrintOLED {
		log.Println("info: ------------ OLED ----------------------- SKIPPED ")
	} else {
		log.Println("info: ------------ OLED ----------------------- ")
		log.Println("info: Enabled                 " + fmt.Sprintf("%v", OLEDEnabled))
		log.Println("info: Interfacetype           " + fmt.Sprintf("%v", OLEDInterfacetype))
		log.Println("info: DisplayRows             " + fmt.Sprintf("%v", OLEDDisplayRows))
		log.Println("info: DisplayColumns          " + fmt.Sprintf("%v", OLEDDisplayColumns))
		log.Println("info: DefaultI2cBus           " + fmt.Sprintf("%v", OLEDDefaultI2cBus))
		log.Println("info: DefaultI2cAddress       " + fmt.Sprintf("%v", OLEDDefaultI2cAddress))
		log.Println("info: ScreenWidth             " + fmt.Sprintf("%v", OLEDScreenWidth))
		log.Println("info: ScreenHeight            " + fmt.Sprintf("%v", OLEDScreenHeight))
		log.Println("info: CommandColumnAddressing " + fmt.Sprintf("%v", OLEDCommandColumnAddressing))
		log.Println("info: AddressBasePageStart    " + fmt.Sprintf("%v", OLEDAddressBasePageStart))
		log.Println("info: CharLength              " + fmt.Sprintf("%v", OLEDCharLength))
		log.Println("info: StartColumn             " + fmt.Sprintf("%v", OLEDStartColumn))
	}

	if !Config.Global.Software.PrintVariables.PrintGPS {
		log.Println("info: ------------ GPS  ------------------------ SKIPPED ")
	} else {
		log.Println("info: ------------ GPS  ------------------------ ")
		log.Println("info: Enabled                " + fmt.Sprintf("%t", Config.Global.Hardware.GPS.Enabled))
		log.Println("info: Port                   ", Config.Global.Hardware.GPS.Port)
		log.Println("info: Baud                   " + fmt.Sprintf("%v", Config.Global.Hardware.GPS.Baud))
		log.Println("info: TxData                 ", Config.Global.Hardware.GPS.TxData)
		log.Println("info: Even                   " + fmt.Sprintf("%v", Config.Global.Hardware.GPS.Even))
		log.Println("info: Odd                    " + fmt.Sprintf("%v", Config.Global.Hardware.GPS.Odd))
		log.Println("info: RS485                  " + fmt.Sprintf("%v", Config.Global.Hardware.GPS.Rs485))
		log.Println("info: RS485 High During Send " + fmt.Sprintf("%v", Config.Global.Hardware.GPS.Rs485HighDuringSend))
		log.Println("info: RS485 High After Send  " + fmt.Sprintf("%v", Config.Global.Hardware.GPS.Rs485HighAfterSend))
		log.Println("info: Stop Bits              " + fmt.Sprintf("%v", Config.Global.Hardware.GPS.StopBits))
		log.Println("info: Data Bits              " + fmt.Sprintf("%v", Config.Global.Hardware.GPS.DataBits))
		log.Println("info: Char Time Out          " + fmt.Sprintf("%v", Config.Global.Hardware.GPS.CharTimeOut))
		log.Println("info: Min Read               " + fmt.Sprintf("%v", Config.Global.Hardware.GPS.MinRead))
		log.Println("info: Rx                     " + fmt.Sprintf("%t", Config.Global.Hardware.GPS.Rx))
	}

	if !Config.Global.Software.PrintVariables.PrintTraccar {
		log.Println("info: ------------ Traccar  ------------------------ SKIPPED")

	} else {
		log.Println("info: ------------ Traccar  ------------------------ ")
		log.Println("info: Enabled               " + fmt.Sprintf("%t", Config.Global.Hardware.Traccar.Enabled))
		log.Println("info: Track                 ", Config.Global.Hardware.Traccar.Track)
		log.Println("info: ClientID              ", Config.Global.Hardware.Traccar.ClientId)
		log.Println("info: Device Screen Enabled " + fmt.Sprintf("%t", Config.Global.Hardware.Traccar.DeviceScreenEnabled))
	}

	if !Config.Global.Software.PrintVariables.PrintPanic {
		log.Println("info: ------------ PANIC Function -------------- SKIPPED ")
	} else {
		log.Println("info: ------------ PANIC Function -------------- ")
		log.Println("info: Panic Function Enable          ", fmt.Sprintf("%t", Config.Global.Hardware.PanicFunction.Enabled))
		log.Println("info: Panic Sound Filename and Path  ", Config.Global.Hardware.PanicFunction.FilenameAndPath)
		log.Println("info: Panic Message                  ", Config.Global.Hardware.PanicFunction.Message)
		log.Println("info: Panic Email Send               ", fmt.Sprintf("%t", Config.Global.Hardware.PanicFunction.PMailEnabled))
		log.Println("info: Panic Message Send Recursively ", fmt.Sprintf("%t", Config.Global.Hardware.PanicFunction.RecursiveSendMessage))
		log.Println("info: Panic Volume                   ", fmt.Sprintf("%v", Config.Global.Hardware.PanicFunction.Volume))
		log.Println("info: Panic Send Ident               ", fmt.Sprintf("%t", Config.Global.Hardware.PanicFunction.SendIdent))
		log.Println("info: Panic Send GPS Location        ", fmt.Sprintf("%t", Config.Global.Hardware.PanicFunction.SendGpsLocation))
		log.Println("info: Panic TX Lock Enabled          ", fmt.Sprintf("%t", Config.Global.Hardware.PanicFunction.TxLockEnabled))
		log.Println("info: Panic TX Lock Timeout Secs     ", fmt.Sprintf("%v", Config.Global.Hardware.PanicFunction.TxLockEnabled))
		log.Println("info: Panic Low Profile Lights Enable", fmt.Sprintf("%v", Config.Global.Hardware.PanicFunction.PLowProfile))
	}

	if !Config.Global.Software.PrintVariables.PrintUSBKeyboard {
		log.Println("info: ------------ USBKeyboard Function ------ SKIPPED ")
	} else {
		log.Println("info: ------------ USBKeyboard Function -------------- ")
		log.Println("info: USBKeyboardEnabled", Config.Global.Hardware.USBKeyboard.Enabled)
		log.Println("info: USBKeyboardPath", Config.Global.Hardware.USBKeyboard.USBKeyboardPath)
		log.Println("info: NumLockScanID", Config.Global.Hardware.USBKeyboard.NumlockScanID)
	}

	if !Config.Global.Software.PrintVariables.PrintAudioRecord {
		log.Println("info: ------------ AUDIO RECORDING Function ------- SKIPPED ")
	} else {
		log.Println("info: ------------ AUDIO RECORDING Function -------------- ")
		log.Println("info: Audio Recording Enabled " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.Enabled))
		log.Println("info: Audio Recording On Start " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordOnStart))
		log.Println("info: Audio Recording System " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordSystem))
		log.Println("info: Audio Record Mode " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordMode))
		log.Println("info: Audio Record Timeout " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordTimeout))
		log.Println("info: Audio Record From Output " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordFromOutput))
		log.Println("info: Audio Record From Input " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordFromInput))
		log.Println("info: Audio Recording Mic Timeout " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordMicTimeout))
		log.Println("info: Audio Recording Save Path " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordSavePath))
		log.Println("info: Audio Recording Archive Path " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordArchivePath))
		log.Println("info: Audio Recording Soft " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordSoft))
		log.Println("info: Audio Recording Profile " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordProfile))
		log.Println("info: Audio Recording File Format " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordFileFormat))
		log.Println("info: Audio Recording Chunk Size " + fmt.Sprintf("%v", Config.Global.Hardware.AudioRecordFunction.RecordChunkSize))
	}

	if !Config.Global.Software.PrintVariables.PrintKeyboardMap {
		log.Println("info: ------------ KeyboardMap Function ------ SKIPPED ")
	} else {
		log.Println("info: ------------ KeyboardMap Function -------------- ")
		counter := 1
		for _, value := range Config.Global.Hardware.Keyboard.Command {
			if value.Enabled {
				log.Printf("info: %v Enabled %v Command %v ParamValue %v\n", counter, value.Enabled, value.Action, value.Paramvalue)
				counter++
			}
			if value.Ttykeyboard.Enabled {
				log.Println("info: TTYKeyboard " + fmt.Sprintf("%+v", value.Ttykeyboard))
			}
			if value.Usbkeyboard.Enabled {
				log.Println("info: USBKeyboard " + fmt.Sprintf("%+v", value.Usbkeyboard))
			}
		}
	}

	if !Config.Global.Software.PrintVariables.PrintRadioModule {
		log.Println("info: ------------ KeyboardMap Function ------ SKIPPED ")
	} else {
		log.Println("info: ------------ RadioModule Function -------------- ")
		log.Println("info: Radio  Enabled     " + fmt.Sprintf("%v", Config.Global.Hardware.Radio.Enabled))
		log.Println("info: SA818  Enabled     " + fmt.Sprintf("%v", Config.Global.Hardware.Radio.Sa818.Enabled))
		log.Println("info: SA818  PD Enabled  " + fmt.Sprintf("%v", Config.Global.Hardware.Radio.Sa818.PDEnabled))
		log.Println("info: Connect Channel ID " + fmt.Sprintf("%v", Config.Global.Hardware.Radio.ConnectChannelID))
		log.Println("info: Serial Enabled  " + fmt.Sprintf("%v", Config.Global.Hardware.Radio.Sa818.Serial.Enabled))
		log.Println("info: Serial Port     " + fmt.Sprintf("%v", Config.Global.Hardware.Radio.Sa818.Serial.Port))
		log.Println("info: Serial Baud     " + fmt.Sprintf("%v", Config.Global.Hardware.Radio.Sa818.Serial.Baud))
		log.Println("info: Serial DataBits " + fmt.Sprintf("%v", Config.Global.Hardware.Radio.Sa818.Serial.Databits))
		log.Println("info: Serial StopBits " + fmt.Sprintf("%v", Config.Global.Hardware.Radio.Sa818.Serial.Stopbits))
		counter := 1
		for _, channel := range Config.Global.Hardware.Radio.Sa818.Channels.Channel {
			if channel.Enabled {
				var ChannelIsCurrent string = "x"
				if channel.ID == Config.Global.Hardware.Radio.ConnectChannelID {
					ChannelIsCurrent = "✓"
				}
				log.Printf("info: %v %v. Bandwidth %v RXFreq %vMhz, TXFreq %vMhz Squelch %v CTSS Tone %v DCS Tone %v Pre/DeEmph %v Highpass %v Lowpass %v Volume %v TXPower %v\n", ChannelIsCurrent, counter, channel.Bandwidth, channel.Rxfreq, channel.Txfreq, channel.Squelch, channel.Ctcsstone, channel.Dcstone, channel.Predeemph, channel.Highpass, channel.Lowpass, channel.Volume, channel.TXPower)
				counter++
			}
		}
	}

	if !Config.Global.Software.PrintVariables.PrintMultimedia {
		log.Println("info: ------------ Multimedia Function ------ SKIPPED ")
	} else {
		log.Println("info: ------------ Multimedia Function -------------- ")
		for _, value := range Config.Global.Multimedia.ID {
			if value.Enabled {
				log.Printf("info: Announcement Tone Enabled %v \n", value.Params.Announcementtone.Enabled)
				log.Printf("info: Announcement Tone File %v \n", value.Params.Announcementtone.File)
				log.Printf("info: GPIO Enabled %v \n", value.Params.GPIO.Enabled)
				log.Printf("info: GPIO Name    %v \n", value.Params.GPIO.Name)
				log.Printf("info: Local Play %v \n", value.Params.Localplay)
				log.Printf("info: Play Into Stream %v \n", value.Params.Playintostream)
				log.Printf("info: Pre  Delay  %v \n", value.Params.Predelay)
				log.Printf("info: Post Delay %v \n", value.Params.Postdelay)
				log.Printf("info: Voice Target %v \n", value.Params.Voicetarget)
				log.Printf("info: Enabled %v \n", value.Enabled)
				log.Printf("info: Media Souce %+v \n", value.Media.Source)
			}
		}
	}

	if !Config.Global.Software.PrintVariables.PrintMemoryChannels {
		log.Println("info: ------------ Memory Channels -------------- SKIPPED ")
	} else {
		log.Println("info: ------------ Memory Channels -------------- ")
		if Config.Global.Software.MemoryChannels.Enabled {
			for _, value := range Config.Global.Software.MemoryChannels.Channel {
				if value.Enabled {
					log.Printf("info: Memory Channel Enabled %v \n", value.Enabled)
					log.Printf("info: Memory Channel Name    %v \n", value.ChannelName)
					log.Printf("info: GPIO Name    %v \n", value.GPIOName)
				}
			}
		}
	}

	if !Config.Global.Software.PrintVariables.PrintPresetVoiceTargets {
		log.Println("info: ------------ Preset VoiceTargets --------------  SKIPPED ")
	} else {
		log.Println("info: ------------ Preset VoiceTargets -------------- ")
		if Config.Global.Software.PresetVoiceTargets.Enabled {
			for _, value := range Config.Global.Software.PresetVoiceTargets.VoiceTargetSet {
				if value.Enabled {
					log.Printf("info: VoiceTarget Enabled %v \n", value.Enabled)
					log.Printf("info: VoiceTarget ID      %v \n", value.ID)
					log.Printf("info: GPIO Name           %v \n", value.GPIOName)
				}
			}
		}
	}
}

func modifyXMLTagServerHopping(inputXMLFile string, newserverindex int) {

	if !FileExists(inputXMLFile) {
		log.Println("error: Cannot Server Hop Cannot Find XML Config File at ", inputXMLFile)
		return
	}

	if Config.Global.Software.Settings.NextServerIndex == newserverindex {
		log.Println("error: Server Index is Not Changed")
		return
	}

	PreparedSEDCommand := fmt.Sprintf("s#<nextserverindex>%d</nextserverindex>#<nextserverindex>%d</nextserverindex>#", Config.Global.Software.Settings.NextServerIndex, newserverindex)
	cmd := exec.Command("sed", "-i", PreparedSEDCommand, inputXMLFile)

	err := cmd.Run()
	if err != nil {
		log.Println("error: Failed to Set Next Server XML Tag with Error ", err)
		return
	}

	killSession()
}

func CheckConfigSanity(reloadxml bool) {

	Warnings := 0
	Alerts := 0

	log.Println("info: Starting XML Configuration Sanity and Logical Checks")

	Counter := 0
	for _, account := range Config.Accounts.Account {
		if account.Default {
			if len(account.Name) == 0 {
				log.Print("warn: Config Error [Section Accounts] Account Name Not Defined for Enabled Account")
			}
			if len(account.ServerAndPort) == 0 {
				log.Print("alert: Config Error [Section Accounts] Account Server And Port Not Defined for Enabled Account")
			}

			if len(account.Certificate) > 0 && !FileExists(account.Certificate) {
				log.Print("warn: Config Error [Section Accounts] Certificate Enabled but Not Found")
			}
			Counter++
		}
	}

	if Counter == 0 {
		log.Print("alert: Config Error [Section Accounts] No Default/Enabled Accounts Found in Config")
		Alerts++
	}

	if Config.Global.Software.Settings.NextServerIndex > Counter {
		if Counter > 0 {
			log.Print("warn: Config Error [Section Settings] Next Server Index Invalid Defaulting back to 0")
			Config.Global.Software.Settings.NextServerIndex = 0
			Warnings++
		} else {
			FatalCleanUp("alert: NextServerIndex is Not Correct Check NextServerIndex in Accounts Section of XML config file!, talkkonnect stopping now!")
		}
	}

	if Config.Global.Software.AutoProvisioning.Enabled {

		if len(Config.Global.Software.AutoProvisioning.TkID) == 0 || len(Config.Global.Software.AutoProvisioning.URL) == 0 || len(Config.Global.Software.AutoProvisioning.SaveFilePath) == 0 || len(Config.Global.Software.AutoProvisioning.SaveFilename) == 0 {
			log.Print("warn: Config Error [Section Autoprovisioning] Some Parameters Not Defined Disabling AutoProvisioning")
			Config.Global.Software.AutoProvisioning.Enabled = false
			Warnings++
		}

	}

	if Config.Global.Software.Beacon.Enabled {
		if len(Config.Global.Software.Beacon.BeaconFileAndPath) == 0 || Config.Global.Software.Beacon.BeaconTimerSecs == 0 || (!Config.Global.Software.Beacon.LocalPlay && !Config.Global.Software.Beacon.Playintostream) {
			log.Print("warn: Config Error [Section Beacon] Some Parameters Not Defined Disabling Beacon")
			Config.Global.Software.Beacon.Enabled = false
			Warnings++
		}
	}

	for index, sounds := range Config.Global.Software.Sounds.Sound {
		if sounds.Enabled {
			if len(sounds.File) > 0 {
				if !FileExists(sounds.File) {
					if !checkRegex("(http|rtsp)", sounds.File) {
						log.Printf("warn: Config Error [Section Sounds] Enabled Sound Event %v File/Link Missing in Config\n", sounds.Event)
						Config.Global.Software.Sounds.Sound[index].Enabled = false
						Warnings++
					}
				}
			}

			volume, _ := strconv.Atoi(sounds.Volume)
			if volume == 0 {
				log.Printf("warn: Config Error [Section Sounds] Enabled Sound Event %v Volume = 0 in Config\n", sounds.Event)
				Config.Global.Software.Sounds.Sound[index].Enabled = false
				Warnings++
			}
		}
	}

	if Config.Global.Software.RemoteSSHConsole.Enabled {
		if len(Config.Global.Software.RemoteSSHConsole.Username) == 0 || len(Config.Global.Software.RemoteSSHConsole.Password) == 0 || len(Config.Global.Software.RemoteSSHConsole.IDRSAFile) == 0 || len(Config.Global.Software.RemoteSSHConsole.Listen) == 0 {
			log.Print("warn: Config Error [Section RemoteConsole] Some Parameters Not Defined Disabling RemoteSSHConsole")
			Config.Global.Software.RemoteSSHConsole.Enabled = false
			Warnings++
		}
	}
	if Config.Global.Software.SMTP.Enabled {
		if len(Config.Global.Software.SMTP.Username) == 0 || len(Config.Global.Software.SMTP.Password) == 0 || len(Config.Global.Software.SMTP.Receiver) == 0 {
			log.Print("warn: Config Error [Section SMTP] Some Parameters Not Defined Disabling SMTP")
			Config.Global.Software.SMTP.Enabled = false
			Warnings++
		}
	}

	if Config.Global.Software.MemoryChannels.Enabled {
		for _, memorychannel := range Config.Global.Software.MemoryChannels.Channel {
			if !(memorychannel.GPIOName == "memorychannel1" || memorychannel.GPIOName == "memorychannel2" || memorychannel.GPIOName == "memorychannel3" || memorychannel.GPIOName == "memorychannel4") {
				log.Print("warn: Config Error [Section MEMORYCHANNELS] Some Parameters Not Defined CorrectlyDisabling Memory Channels")
				Config.Global.Software.MemoryChannels.Enabled = false
				Warnings++
			}
		}
	}

	if Config.Global.Software.PresetVoiceTargets.Enabled {
		for _, voicetarget := range Config.Global.Software.PresetVoiceTargets.VoiceTargetSet {
			if !(voicetarget.GPIOName == "presetvoicetarget1" || voicetarget.GPIOName == "presetvoicetarget2" || voicetarget.GPIOName == "presetvoicetarget3" || voicetarget.GPIOName == "presetvoicetarget4" || voicetarget.GPIOName == "presetvoicetarget5") {
				log.Print("warn: Config Error [Section MEMORYCHANNELS] Some Parameters Not Defined CorrectlyDisabling Memory Channels")
				Config.Global.Software.PresetVoiceTargets.Enabled = false
				Warnings++
			}
		}
	}

	if Config.Global.Hardware.VoiceActivityTimermsecs < 200 {
		log.Print("warn: Config Error [Section Hardware] VoiceActivityTimersecs < 200 setting to 200")
		Config.Global.Hardware.VoiceActivityTimermsecs = 200
		Warnings++
	}

	for index, gpio := range Config.Global.Hardware.IO.Pins.Pin {
		if gpio.Enabled {
			if !(gpio.Direction == "input" || gpio.Direction == "output") {
				log.Printf("warn: Config Error [Section GPIO] Enabled GPIO Name %v Pin Number %v Direction %v Misconfiguired\n", gpio.Name, gpio.PinNo, gpio.Direction)
				Config.Global.Hardware.IO.Pins.Pin[index].Enabled = false
				Warnings++
			}
			if (gpio.Direction == "input") && !(gpio.Device == "pushbutton" || gpio.Device == "toggleswitch" || gpio.Device == "rotaryencoder") {
				log.Printf("warn: Config Error [Section GPIO] Enabled Input GPIO Name %v Pin Number %v Name Mis-Configured\n", gpio.Name, gpio.PinNo)
				Config.Global.Hardware.IO.Pins.Pin[index].Enabled = false
				Warnings++
			}
			if (gpio.Direction == "output") && !(gpio.Device == "led/relay" || gpio.Device == "lcd") {
				log.Printf("warn: Config Error [Section GPIO] Enabled Output GPIO Name %v Pin Number %v Name Mis-Configured\n", gpio.Name, gpio.PinNo)
				Config.Global.Hardware.IO.Pins.Pin[index].Enabled = false
				Warnings++
			}

			if !(gpio.Name == "voiceactivity" || gpio.Name == "participants" || gpio.Name == "transmit" || gpio.Name == "online" || gpio.Name == "attention" || gpio.Name == "voicetarget" || gpio.Name == "heartbeat" || gpio.Name == "backlight" || gpio.Name == "relay0" || gpio.Name == "txptt" || gpio.Name == "txtoggle" || gpio.Name == "channelup" || gpio.Name == "channeldown" || gpio.Name == "panic" || gpio.Name == "streamtoggle" || gpio.Name == "comment" || gpio.Name == "rotarya" || gpio.Name == "rotaryb" || gpio.Name == "rotarybutton" || gpio.Name == "volup" || gpio.Name == "voldown" || gpio.Name == "nextserver" || gpio.Name == "memorychannel1" || gpio.Name == "memorychannel2" || gpio.Name == "memorychannel3" || gpio.Name == "memorychannel4" || gpio.Name == "analogrelay1" || gpio.Name == "analogrelay2" || gpio.Name == "shutdown" || gpio.Name == "presetvoicetarget1" || gpio.Name == "presetvoicetarget2" || gpio.Name == "presetvoicetarget3" || gpio.Name == "presetvoicetarget4" || gpio.Name == "presetvoicetarget5") {
				log.Printf("warn: Config Error [Section GPIO] Enabled GPIO Name %v Pin Number %v Invalid Name\n", gpio.Name, gpio.PinNo)
				Config.Global.Hardware.IO.Pins.Pin[index].Enabled = false
				Warnings++
			}

			if gpio.PinNo < 0 || gpio.PinNo > 27 {
				log.Printf("warn: Config Error [Section GPIO] Enabled GPIO Name %v Pin Number %v Invalid GPIO Number\n", gpio.Name, gpio.PinNo)
				Config.Global.Hardware.IO.Pins.Pin[index].Enabled = false
				Warnings++
			}

			if (Config.Global.Hardware.OLED.Enabled || Config.Global.Hardware.IO.GPIOExpander.Enabled || Config.Global.Hardware.LCD.InterfaceType == "i2c") && (gpio.PinNo == 2 || gpio.PinNo == 3) {
				log.Printf("warn: Config Possible Pins Clash with I2C Interface [Section GPIO] Enabled GPIO Name %v Pin Number %v Invalid GPIO Number\n", gpio.Name, gpio.PinNo)
				Warnings++
			}

			if (Config.Global.Hardware.LedStripEnabled) && (gpio.PinNo == 7 || gpio.PinNo == 8 || gpio.PinNo == 9 || gpio.PinNo == 10 || gpio.PinNo == 11) {
				log.Printf("warn: Config Possible Pins Clash with SPI Interface [Section GPIO] Enabled GPIO Name %v Pin Number %v Invalid GPIO Number\n", gpio.Name, gpio.PinNo)
				Warnings++
			}

			if gpio.ID > 8 {
				log.Print("warn: Config Error [Section GPIO] Invalid ChipID Address")
				Config.Global.Hardware.IO.Pins.Pin[index].Enabled = false
				Warnings++
			}

			if gpio.Name == "heartbeat" {
				if Config.Global.Hardware.HeartBeat.Periodmsecs < 100 || Config.Global.Hardware.HeartBeat.LEDOnmsecs < 100 || Config.Global.Hardware.HeartBeat.LEDOffmsecs < 100 {
					if gpio.PinNo == 0 {
						log.Printf("warn: Config Error [Section GPIO] Name %v Invalid GPIO Pin %v Value\n", gpio.Name, gpio.PinNo)
						Config.Global.Hardware.IO.Pins.Pin[index].Enabled = false
						Warnings++
					}
				}
			}

		}
	}

	if Config.Global.Hardware.LCD.BacklightTimerEnabled && (!Config.Global.Hardware.OLED.Enabled || !Config.Global.Hardware.LCD.Enabled) {
		log.Println("warn: Disabling Backlight Timer Since Neither LCD or OLED Displays Enabled")
		Config.Global.Hardware.LCD.BacklightTimerEnabled = false
		Warnings++
	}

	if Config.Global.Hardware.LCD.Enabled {
		if !(Config.Global.Hardware.LCD.InterfaceType == "i2c" || Config.Global.Hardware.LCD.InterfaceType == "parallel") {
			log.Printf("warn: Config Error [Section LCD] Enabled LCD Interface Type %v Invalid\n", Config.Global.Hardware.LCD.InterfaceType)
			Config.Global.Hardware.LCD.Enabled = false
			Warnings++
		}

		if Config.Global.Hardware.LCD.InterfaceType == "i2c" {
			if Config.Global.Hardware.LCD.I2CAddress == 0 {
				log.Printf("warn: Config Error [Section LCD] Enabled LCD Interface %v Invalid\n", Config.Global.Hardware.LCD.InterfaceType)
				Config.Global.Hardware.LCD.Enabled = false
				Warnings++
			}
		}

		if Config.Global.Hardware.LCD.InterfaceType == "parallel" {
			if Config.Global.Hardware.LCD.RsPin == 0 {
				log.Printf("warn: Config Error [Section LCD] Enabled LCD Interface %v RsPin %v Invalid\n", Config.Global.Hardware.LCD.InterfaceType, Config.Global.Hardware.LCD.RsPin)
				Config.Global.Hardware.LCD.Enabled = false
				Warnings++
			}
			if Config.Global.Hardware.LCD.EPin == 0 {
				log.Printf("warn: Config Error [Section LCD] Enabled LCD Interface %v EPin %v Invalid\n", Config.Global.Hardware.LCD.InterfaceType, Config.Global.Hardware.LCD.RsPin)
				Config.Global.Hardware.LCD.Enabled = false
				Warnings++
			}
			if Config.Global.Hardware.LCD.D4Pin == 0 {
				log.Printf("warn: Config Error [Section LCD] Enabled LCD Interface %v D4Pin %v Invalid\n", Config.Global.Hardware.LCD.InterfaceType, Config.Global.Hardware.LCD.RsPin)
				Config.Global.Hardware.LCD.Enabled = false
				Warnings++
			}
			if Config.Global.Hardware.LCD.D5Pin == 0 {
				log.Printf("warn: Config Error [Section LCD] Enabled LCD Interface %v D5Pin %v Invalid\n", Config.Global.Hardware.LCD.InterfaceType, Config.Global.Hardware.LCD.RsPin)
				Config.Global.Hardware.LCD.Enabled = false
				Warnings++
			}
			if Config.Global.Hardware.LCD.D6Pin == 0 {
				log.Printf("warn: Config Error [Section LCD] Enabled LCD Interface %v D6Pin %v Invalid\n", Config.Global.Hardware.LCD.InterfaceType, Config.Global.Hardware.LCD.RsPin)
				Config.Global.Hardware.LCD.Enabled = false
				Warnings++
			}
			if Config.Global.Hardware.LCD.D7Pin == 0 {
				log.Printf("warn: Config Error [Section LCD] Enabled LCD Interface %v D7Pin %v Invalid\n", Config.Global.Hardware.LCD.InterfaceType, Config.Global.Hardware.LCD.RsPin)
				Config.Global.Hardware.LCD.Enabled = false
				Warnings++
			}
		}
	}
	if Config.Global.Hardware.LCD.BacklightTimerEnabled {
		if Config.Global.Hardware.LCD.BackLightTimeoutSecs == 0 {
			log.Print("warn: Config Error [Section LCD] Disabling Invalid Backlight Timer")
			Config.Global.Hardware.LCD.BacklightTimerEnabled = false
			Warnings++
		}
	}

	if Config.Global.Hardware.GPS.Enabled {
		if !FileExists(Config.Global.Hardware.GPS.Port) {
			log.Printf("warn: Config Error [Section GPS] Enabled GPS Port %v Invalid\n", Config.Global.Hardware.GPS.Port)
			Config.Global.Hardware.GPS.Enabled = false
			Warnings++
		}
		if !(Config.Global.Hardware.GPS.Baud == 2400 || Config.Global.Hardware.GPS.Baud == 4800 || Config.Global.Hardware.GPS.Baud == 9600 || Config.Global.Hardware.GPS.Baud == 14400 || Config.Global.Hardware.GPS.Baud == 19200 || Config.Global.Hardware.GPS.Baud == 38400 || Config.Global.Hardware.GPS.Baud == 57600 || Config.Global.Hardware.GPS.Baud == 115200) {
			log.Printf("warn: Config Error [Section GPS] Enabled GPS Port %v Invalid Baud %v Setting\n", Config.Global.Hardware.GPS.Port, Config.Global.Hardware.GPS.Baud)
			Config.Global.Hardware.GPS.Enabled = false
			Warnings++
		}

		if Config.Global.Hardware.GPS.Even && Config.Global.Hardware.GPS.Odd {
			log.Printf("warn: Config Error [Section GPS] Enabled GPS Port %v Invalid Parity Both Even & Odd Set\n", Config.Global.Hardware.GPS.Port)
			Config.Global.Hardware.GPS.Enabled = false
			Warnings++
		}

		if Config.Global.Hardware.GPS.StopBits == 0 {
			log.Printf("warn: Config Error [Section GPS] Enabled GPS Port %v Invalid Stop Bits\n", Config.Global.Hardware.GPS.Port)
			Config.Global.Hardware.GPS.Enabled = false
			Warnings++
		}

		if Config.Global.Hardware.GPS.DataBits == 0 {
			log.Printf("warn: Config Error [Section GPS] Enabled GPS Port %v Invalid Data Bits\n", Config.Global.Hardware.GPS.Port)
			Config.Global.Hardware.GPS.Enabled = false
			Warnings++
		}
	}

	if Config.Global.Hardware.AudioRecordFunction.Enabled {
		if !(Config.Global.Hardware.AudioRecordFunction.RecordSystem == "alsa" || Config.Global.Hardware.AudioRecordFunction.RecordSystem == "pulseaudio") {
			Config.Global.Hardware.AudioRecordFunction.RecordSystem = "alsa"
		}
	}

	if Config.Global.Software.RemoteControl.MQTT.Enabled {

		if len(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTSubTopic) == 0 {
			log.Println("warn: Config Error [Section MQTT] Enabled MQTT With Empty Sub Topic")
			Config.Global.Software.RemoteControl.MQTT.Enabled = false
			Warnings++
		}
		if len(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTPubTopic) == 0 {
			log.Println("warn: Config Error [Section MQTT] Enabled MQTT With Empty Pub Topic")
			Config.Global.Software.RemoteControl.MQTT.Enabled = false
			Warnings++
		}
		if len(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTBroker) == 0 {
			log.Println("warn: Config Error [Section MQTT] Enabled MQTT With Empty Broker")
			Config.Global.Software.RemoteControl.MQTT.Enabled = false
			Warnings++
		}
		if len(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTPassword) == 0 {
			log.Println("warn: Config Error [Section MQTT] Enabled MQTT With Empty MQTTPassword")
			Config.Global.Software.RemoteControl.MQTT.Enabled = false
			Warnings++
		}
		if len(Config.Global.Software.RemoteControl.MQTT.Settings.MQTTId) == 0 {
			log.Println("warn: Config Error [Section MQTT] Enabled MQTT With Empty MQTTID")
			Config.Global.Software.RemoteControl.MQTT.Enabled = false
			Warnings++
		}

	}

	if Config.Global.Software.IgnoreUser.IgnoreUserEnabled {
		if len(Config.Global.Software.IgnoreUser.IgnoreUserRegex) < 4 {
			log.Printf("warn: Config Error [Section ignoreuser]  %v Invalid Regex\n", Config.Global.Software.IgnoreUser.IgnoreUserRegex)
			Config.Global.Software.IgnoreUser.IgnoreUserEnabled = false
		}
	}

	for index, keyboard := range Config.Global.Hardware.Keyboard.Command {
		if keyboard.Enabled {
			if !(keyboard.Action == "channelup" || keyboard.Action == "channeldown" || keyboard.Action == "serverup" || keyboard.Action == "serverdown" || keyboard.Action == "mute" || keyboard.Action == "unmute" || keyboard.Action == "mute-toggle" || keyboard.Action == "stream-toggle" || keyboard.Action == "volumeup" || keyboard.Action == "volumedown" || keyboard.Action == "setcomment" || keyboard.Action == "transmitstart" || keyboard.Action == "transmitstop" || keyboard.Action == "pttkey" || keyboard.Action == "soundinterfacepttkey" || keyboard.Action == "record" || keyboard.Action == "voicetargetset" || keyboard.Action == "volup" || keyboard.Action == "voldown" || keyboard.Action == "mqttpubpayloadset" || keyboard.Action == "changechannel" || keyboard.Action == "listentochannelon" || keyboard.Action == "listentochanneloff" || keyboard.Action == "gpioinput" || keyboard.Action == "gpiooutput" || keyboard.Action == "volumetxup" || keyboard.Action == "volumetxdown") {
				log.Printf("warn: Config Error [Section Keyboard] Enabled Keyboard Action %v Invalid\n", keyboard.Action)
				Config.Global.Hardware.Keyboard.Command[index].Enabled = false
				Warnings++

			}
			if keyboard.Ttykeyboard.Enabled {
				if keyboard.Ttykeyboard.Scanid == 0 || keyboard.Ttykeyboard.Scanid > 255 {
					log.Printf("warn: Config Error [Section Keyboard] Enabled TTYKeyboard ScanID %v Invalid\n", keyboard.Ttykeyboard.Scanid)
					Config.Global.Hardware.Keyboard.Command[index].Ttykeyboard.Enabled = false
					Warnings++
				}
			}
			if keyboard.Usbkeyboard.Enabled {
				if keyboard.Usbkeyboard.Scanid == 0 || keyboard.Usbkeyboard.Scanid > 255 {
					log.Printf("warn: Config Error [Section Keyboard] Enabled USBKeyboard ScanID %v Invalid\n", keyboard.Usbkeyboard.Scanid)
					Config.Global.Hardware.Keyboard.Command[index].Usbkeyboard.Enabled = false
					Warnings++
				}
			}
		}
	}

	if Warnings+Alerts > 0 {
		if Alerts > 0 {
			FatalCleanUp("alert: Fatal Errors Found In talkkonnect.xml config file please fix errors, talkkonnect stopping now!")
		}

		if Warnings > 0 {
			log.Println("warn: Non-Critical Errors Found In talkkonnect.xml config file please fix errors or talkkonnect may not behave as expected")
		}
	} else {
		log.Println("info: Finished XML Configuration Sanity and Logical Checks Without Any Alerts/Errors/Warnings")
	}
}
