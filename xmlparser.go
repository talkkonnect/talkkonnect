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
 * Zoran Dimitrijevic
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
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	goled "github.com/talkkonnect/go-oled-i2c"
	"github.com/talkkonnect/go-openal/openal"
	"github.com/talkkonnect/gumble/gumble"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	"golang.org/x/sys/unix"
)

//version and release date
const (
	talkkonnectVersion  string = "1.67.07"
	talkkonnectReleased string = "Aug 20 2021"
)

// Generic Global Variables
var (
	pstream               *gumbleffmpeg.Stream
	AccountCount          int  = 0
	KillHeartBeat         bool = false
	IsPlayStream          bool = false
	BackLightTime              = time.NewTicker(5 * time.Second)
	BackLightTimePtr           = &BackLightTime
	ConnectAttempts            = 0
	IsConnected           bool = false
	source                     = openal.NewSource()
	StartTime                  = time.Now()
	BufferToOpenALCounter      = 0
	AccountIndex          int  = 0
	GenericCounter        int  = 0
	IsNumlock             bool
)

//keyboard settings
var (
	USBKeyboardPath    string = "/dev/input/event0"
	USBKeyboardEnabled bool
	NumlockScanID      rune
)

//account settings
var (
	Default     []bool
	Name        []string
	Server      []string
	Username    []string
	Password    []string
	Insecure    []bool
	Register    []bool
	Certificate []string
	Channel     []string
	Ident       []string
	Tokens      []gumble.AccessTokens
	VT          []VTStruct
)

//software settings
var (
	OutputDevice       string = "Speaker"
	OutputDeviceShort  string
	LogFilenameAndPath string = "/var/log/talkkonnect.log"
	Logging            string = "screen"
	Loglevel           string = "info"
	Daemonize          bool
	SimplexWithMute    bool = true
	TxCounter          bool
	NextServerIndex    int = 0
)

//autoprovision settings
var (
	APEnabled    bool
	TkID         string
	URL          string
	SaveFilePath string
	SaveFilename string
)

//beacon settings
var (
	BeaconEnabled         bool
	BeaconTimerSecs       int = 30
	BeaconFilenameAndPath string
	BVolume               float32 = 1.0
)

//tts
var (
	TTSEnabled                           bool
	TTSVolumeLevel                       int
	TTSParticipants                      bool
	TTSChannelUp                         bool
	TTSChannelUpFilenameAndPath          string
	TTSChannelDown                       bool
	TTSChannelDownFilenameAndPath        string
	TTSMuteUnMuteSpeaker                 bool
	TTSMuteUnMuteSpeakerFilenameAndPath  string
	TTSCurrentVolumeLevel                bool
	TTSCurrentVolumeLevelFilenameAndPath string
	TTSDigitalVolumeUp                   bool
	TTSDigitalVolumeUpFilenameAndPath    string
	TTSDigitalVolumeDown                 bool
	TTSDigitalVolumeDownFilenameAndPath  string
	TTSListServerChannels                bool
	TTSListServerChannelsFilenameAndPath string
	TTSStartTransmitting                 bool
	TTSStartTransmittingFilenameAndPath  string
	TTSStopTransmitting                  bool
	TTSStopTransmittingFilenameAndPath   string
	TTSListOnlineUsers                   bool
	TTSListOnlineUsersFilenameAndPath    string
	TTSPlayStream                        bool
	TTSPlayStreamFilenameAndPath         string
	TTSRequestGpsPosition                bool
	TTSRequestGpsPositionFilenameAndPath string
	TTSNextServer                        bool
	TTSNextServerFilenameAndPath         string
	TTSPreviousServer                    bool
	TTSPreviousServerFilenameAndPath     string
	TTSPanicSimulation                   bool
	TTSPanicSimulationFilenameAndPath    string
	TTSPrintXmlConfig                    bool
	TTSPrintXmlConfigFilenameAndPath     string
	TTSSendEmail                         bool
	TTSSendEmailFilenameAndPath          string
	TTSDisplayMenu                       bool
	TTSDisplayMenuFilenameAndPath        string
	TTSQuitTalkkonnect                   bool
	TTSQuitTalkkonnectFilenameAndPath    string
	TTSTalkkonnectLoaded                 bool
	TTSTalkkonnectLoadedFilenameAndPath  string
	TTSPingServers                       bool
	TTSPingServersFilenameAndPath        string
	TTSScan                              bool
	TTSScanFilenameAndPath               string
)

//gmail smtp settings
var (
	EmailEnabled       bool
	EmailUsername      string
	EmailPassword      string
	EmailReceiver      string
	EmailSubject       string
	EmailMessage       string
	EmailGpsDateTime   bool
	EmailGpsLatLong    bool
	EmailGoogleMapsURL bool
)

//sound settings
var (
	EventSoundEnabled                 bool
	EventJoinedSoundFilenameAndPath   string
	EventLeftSoundFilenameAndPath     string
	EventMessageSoundFilenameAndPath  string
	AlertSoundEnabled                 bool
	AlertSoundFilenameAndPath         string
	AlertSoundVolume                  float32 = 1
	IncommingBeepSoundEnabled         bool
	IncommingBeepSoundFilenameAndPath string
	IncommingBeepSoundVolume          float32
	RogerBeepSoundEnabled             bool
	RogerBeepSoundFilenameAndPath     string
	RogerBeepSoundVolume              float32
	RepeaterToneEnabled               bool
	RepeaterToneFrequencyHz           int
	RepeaterToneDurationSec           int
	StreamSoundEnabled                bool
	StreamSoundFilenameAndPath        string
	StreamSoundVolume                 float32
)

//api settings
var (
	APIEnabled            bool
	APIListenPort         string
	APIDisplayMenu        bool
	APIChannelUp          bool
	APIChannelDown        bool
	APIMute               bool
	APICurrentVolumeLevel bool
	APIDigitalVolumeUp    bool
	APIDigitalVolumeDown  bool
	APIListServerChannels bool
	APIStartTransmitting  bool
	APIStopTransmitting   bool
	APIListOnlineUsers    bool
	APIPlayStream         bool
	APIRequestGpsPosition bool
	APIEmailEnabled       bool
	APINextServer         bool
	APIPreviousServer     bool
	APIPanicSimulation    bool
	APIScanChannels       bool
	APIDisplayVersion     bool
	APIClearScreen        bool
	APIPingServersEnabled bool
	APIRepeatTxLoopTest   bool
	APIPrintXmlConfig     bool
)

//print xml config sections for easy debugging, set any section to false to prevent printing to screen
var (
	PrintAccount      bool
	PrintLogging      bool
	PrintProvisioning bool
	PrintBeacon       bool
	PrintTTS          bool
	PrintSMTP         bool
	PrintSounds       bool
	PrintTxTimeout    bool
	PrintHTTPAPI      bool
	PrintTargetboard  bool
	PrintLeds         bool
	PrintHeartbeat    bool
	PrintButtons      bool
	PrintComment      bool
	PrintLcd          bool
	PrintOled         bool
	PrintGps          bool
	PrintTraccar      bool
	PrintPanic        bool
	PrintAudioRecord  bool
	PrintMQTT         bool
	PrintKeyboardMap  bool
	PrintUSBKeyboard  bool
)

// mqtt settings
var (
	MQTTEnabled     bool = false
	Iotuuid         string
	RelayAllState   bool = false
	RelayPulseMills time.Duration
	TotalRelays     uint
	RelayPins       = [9]uint{}
	MQTTTopic       string
	MQTTBroker      string
	MQTTPassword    string
	MQTTUser        string
	MQTTId          string
	MQTTCleansess   bool
	MQTTQos         int
	MQTTNum         int
	MQTTPayload     string
	MQTTAction      string
	MQTTStore       string
)

var (
	Key0Enabled  bool
	Key0Targetid uint32
	Key1Enabled  bool
	Key1Targetid uint32
	Key2Enabled  bool
	Key2Targetid uint32
	Key3Enabled  bool
	Key3Targetid uint32
	Key4Enabled  bool
	Key4Targetid uint32
	Key5Enabled  bool
	Key5Targetid uint32
	Key6Enabled  bool
	Key6Targetid uint32
	Key7Enabled  bool
	Key7Targetid uint32
	Key8Enabled  bool
	Key8Targetid uint32
	Key9Enabled  bool
	Key9Targetid uint32
)

// target board settings
var (
	TargetBoard string = "pc"
)

//indicator light settings
var (
	LedStripEnabled     bool
	VoiceActivityLEDPin uint
	ParticipantsLEDPin  uint
	TransmitLEDPin      uint
	OnlineLEDPin        uint
	AttentionLEDPin     uint
)

//heartbeat light settings
var (
	HeartBeatEnabled bool
	HeartBeatLEDPin  uint
	PeriodmSecs      int
	LEDOnmSecs       int
	LEDOffmSecs      int
)

//button settings
var (
	TxButtonPin     uint
	TxTogglePin     uint
	UpButtonPin     uint
	DownButtonPin   uint
	PanicButtonPin  uint
	StreamButtonPin uint
)

//comment settings
var (
	CommentButtonPin  uint
	CommentMessageOff string
	CommentMessageOn  string
)

//HD44780 screen lcd settings
var (
	LCDEnabled               bool
	LCDInterfaceType         string
	LCDI2CAddress            uint8
	LCDBackLightTimerEnabled bool
	LCDBackLightTimeout      time.Duration
	LCDBackLightLEDPin       int
	LCDRSPin                 int
	LCDEPin                  int
	LCDD4Pin                 int
	LCDD5Pin                 int
	LCDD6Pin                 int
	LCDD7Pin                 int
	LCDIsDark                bool
)

//OLED screen settings
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

//txtimeout settings
var (
	TxTimeOutEnabled bool
	TxTimeOutSecs    int
)

//gps settings
var (
	GpsEnabled          bool
	Port                string
	Baud                uint
	TxData              string
	Even                bool
	Odd                 bool
	Rs485               bool
	Rs485HighDuringSend bool
	Rs485HighAfterSend  bool
	StopBits            uint
	DataBits            uint
	CharTimeOut         uint
	MinRead             uint
	Rx                  bool
	GpsInfoVerbose      bool
)

var (
	TrackEnabled           bool
	TraccarSendTo          bool
	TraccarServerURL       string
	TraccarServerIP        string
	TraccarClientId        string
	TraccarReportFrequency int64
	TraccarProto           string
	TraccarServerFullURL   string
	TrackGPSShowLCD        bool
	TrackVerbose           bool
)

//panic function settings
var (
	PEnabled           bool
	PFilenameAndPath   string
	PMessage           string
	PMailEnabled       bool
	PRecursive         bool
	PVolume            float32
	PSendIdent         bool
	PSendGpsLocation   bool
	PTxLockEnabled     bool
	PTxlockTimeOutSecs uint
	PLowProfile        bool
)

var (
	AudioRecordEnabled     bool
	AudioRecordOnStart     bool
	AudioRecordSystem      string
	AudioRecordMode        string
	AudioRecordTimeout     int64
	AudioRecordFromOutput  string
	AudioRecordFromInput   string
	AudioRecordMicTimeout  int64
	AudioRecordSoft        string
	AudioRecordSavePath    string
	AudioRecordArchivePath string
	AudioRecordProfile     string
	AudioRecordFileFormat  string
	AudioRecordChunkSize   string
)

var (
	txcounter           int
	isTx                bool
	CancellableStream   bool = true
	StreamOnStart       bool
	StreamStartAfter    uint
	Accounts            int
	MaxTokensInAccounts int
)

var Document DocumentStruct
var TTYKeyMap = make(map[rune]TTYKBStruct)
var USBKeyMap = make(map[rune]USBKBStruct)

type DocumentStruct struct {
	XMLName  xml.Name `xml:"document"`
	Accounts struct {
		Account []struct {
			Name          string `xml:"name,attr"`
			Default       bool   `xml:"default,attr"`
			ServerAndPort string `xml:"serverandport"`
			UserName      string `xml:"username"`
			Password      string `xml:"password"`
			Insecure      bool   `xml:"insecure"`
			Register      bool   `xml:"register"`
			Certificate   string `xml:"certificate"`
			Channel       string `xml:"channel"`
			Ident         string `xml:"ident"`
			TokensEnabled bool   `xml:"enabled,attr"`
			Tokens        struct {
				Token []string `xml:"token"`
			} `xml:"tokens"`
			Voicetargets struct {
				ID []struct {
					Value uint32 `xml:"value,attr"`
					Users struct {
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
				OutputDevice       string `xml:"outputdevice"`
				OutputDeviceShort  string `xml:"outputdeviceshort"`
				LogFilenameAndPath string `xml:"logfilenameandpath"`
				Logging            string `xml:"logging"`
				Loglevel           string `xml:"loglevel"`
				Daemonize          bool   `xml:"daemonize"`
				CancellableStream  bool   `xml:"cancellablestream"`
				StreamOnStart      bool   `xml:"streamonstart"`
				SimplexWithMute    bool   `xml:"simplexwithmute"`
				TxCounter          bool   `xml:"txcounter"`
				NextServerIndex    int    `xml:"nextserverindex"`
			} `xml:"settings"`
			AutoProvisioning struct {
				Enabled      bool   `xml:"enabled,attr"`
				TkID         string `xml:"tkid"`
				URL          string `xml:"url"`
				SaveFilePath string `xml:"savefilepath"`
				SaveFilename string `xml:"savefilename"`
			} `xml:"autoprovisioning"`
			Beacon struct {
				Enabled           bool    `xml:"enabled,attr"`
				BeaconTimerSecs   int     `xml:"beacontimersecs"`
				BeaconFileAndPath string  `xml:"beaconfileandpath"`
				Volume            float32 `xml:"volume"`
			} `xml:"beacon"`
			TTS struct {
				Enabled                           bool   `xml:"enabled,attr"`
				VolumeLevel                       int    `xml:"volumelevel"`
				Participants                      bool   `xml:"participants"`
				ChannelUp                         bool   `xml:"channelup"`
				ChannelUpFilenameAndPath          string `xml:"channelupfilenameandpath"`
				ChannelDown                       bool   `xml:"channeldown"`
				ChannelDownFilenameAndPath        string `xml:"channeldownfilenameandpath"`
				MuteUnmuteSpeaker                 bool   `xml:"muteunmutespeaker"`
				MuteUnmuteSpeakerFilenameAndPath  string `xml:"muteunmutespeakerfilenameandpath"`
				CurrentVolumeLevel                bool   `xml:"currentvolumelevel"`
				CurrentVolumeLevelFilenameAndPath string `xml:"currentvolumelevelfilenameandpath"`
				DigitalVolumeUp                   bool   `xml:"digitalvolumeup"`
				DigitalVolumeUpFilenameAndPath    string `xml:"digitalvolumeupfilenameandpath"`
				DigitalVolumeDown                 bool   `xml:"digitalvolumedown"`
				DigitalVolumeDownFilenameAndPath  string `xml:"digitalvolumedownfilenameandpath"`
				ListServerChannels                bool   `xml:"listserverchannels"`
				ListServerChannelsFilenameAndPath string `xml:"listserverchannelsfilenameandpath"`
				StartTransmitting                 bool   `xml:"starttransmitting"`
				StartTransmittingFilenameAndPath  string `xml:"starttransmittingfilenameandpath"`
				StopTransmitting                  bool   `xml:"stoptransmitting"`
				StopTransmittingFilenameAndPath   string `xml:"stoptransmittingfilenameandpath"`
				ListOnlineUsers                   bool   `xml:"listonlineusers"`
				ListOnlineUsersFilenameAndPath    string `xml:"listonlineusersfilenameandpath"`
				PlayStream                        bool   `xml:"playstream"`
				PlayStreamFilenameAndPath         string `xml:"playstreamfilenameandpath"`
				RequestGpsPosition                bool   `xml:"requestgpsposition"`
				RequestGpsPositionFilenameAndPath string `xml:"requestgpspositionfilenameandpath"`
				NextServer                        bool   `xml:"nextserver"`
				NextServerFilenameAndPath         string `xml:"nextserverfilenameandpath"`
				PreviousServer                    bool   `xml:"previousserver"`
				PreviousServerFilenameAndPath     string `xml:"previousserverfilenameandpath"`
				PanicSimulation                   bool   `xml:"panicsimulation"`
				PanicSimulationFilenameAndPath    string `xml:"panicsimulationfilenameandpath"`
				PrintXmlConfig                    bool   `xml:"printxmlconfig"`
				PrintXmlConfigFilenameAndPath     string `xml:"printxmlconfigfilenameandpath"`
				SendEmail                         bool   `xml:"sendemail"`
				SendEmailFilenameAndPath          string `xml:"sendemailfilenameandpath"`
				DisplayMenu                       bool   `xml:"displaymenu"`
				DisplayMenuFilenameAndPath        string `xml:"displaymenufilenameandpath"`
				QuitTalkkonnect                   bool   `xml:"quittalkkonnect"`
				QuitTalkkonnectFilenameAndPath    string `xml:"quittalkkonnectfilenameandpath"`
				TalkkonnectLoaded                 bool   `xml:"talkkonnectloaded"`
				TalkkonnectLoadedFilenameAndPath  string `xml:"talkkonnectloadedfilenameandpath"`
				PingServers                       bool   `xml:"pingservers"`
				PingServersFilenameAndPath        string `xml:"pingserversfilenameandpath"`
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
				Event struct {
					Enabled                bool   `xml:"enabled,attr"`
					JoinedFilenameAndPath  string `xml:"joinedfilenameandpath"`
					LeftFilenameAndPath    string `xml:"leftfilenameandpath"`
					MessageFilenameAndPath string `xml:"messagefilenameandpath"`
				} `xml:"event"`
				Alert struct {
					Enabled         bool    `xml:"enabled,attr"`
					FilenameAndPath string  `xml:"filenameandpath"`
					Volume          float32 `xml:"volume"`
				} `xml:"alert"`
				IncommingBeep struct {
					Enabled         bool    `xml:"enabled,attr"`
					FilenameAndPath string  `xml:"filenameandpath"`
					Volume          float32 `xml:"volume"`
				} `xml:"incommingbeep"`
				RogerBeep struct {
					Enabled         bool    `xml:"enabled,attr"`
					FilenameAndPath string  `xml:"filenameandpath"`
					Volume          float32 `xml:"volume"`
				} `xml:"rogerbeep"`
				RepeaterTone struct {
					Enabled         bool `xml:"enabled,attr"`
					ToneFrequencyHz int  `xml:"tonefrequencyhz"`
					ToneDurationSec int  `xml:"tonedurationsec"`
				} `xml:"repeatertone"`
				Stream struct {
					Enabled         bool    `xml:"enabled,attr"`
					FilenameAndPath string  `xml:"filenameandpath"`
					Volume          float32 `xml:"volume"`
				} `xml:"stream"`
			} `xml:"sounds"`
			TxTimeOut struct {
				Enabled       bool `xml:"enabled,attr"`
				TxTimeOutSecs int  `xml:"txtimeoutsecs"`
			} `xml:"txtimeout"`
			API struct {
				Enabled            bool   `xml:"enabled,attr"`
				ListenPort         string `xml:"apilistenport"`
				DisplayMenu        bool   `xml:"displaymenu"`
				ChannelUp          bool   `xml:"channelup"`
				ChannelDown        bool   `xml:"channeldown"`
				Mute               bool   `xml:"mute"`
				CurrentVolumeLevel bool   `xml:"currentvolumelevel"`
				DigitalVolumeUp    bool   `xml:"digitalvolumeup"`
				DigitalVolumeDown  bool   `xml:"digitalvolumedown"`
				ListServerChannels bool   `xml:"listserverchannels"`
				StartTransmitting  bool   `xml:"starttransmitting"`
				StopTransmitting   bool   `xml:"stoptransmitting"`
				ListOnlineUsers    bool   `xml:"listonlineusers"`
				PlayStream         bool   `xml:"playstream"`
				RequestGpsPosition bool   `xml:"requestgpsposition"`
				PreviousServer     bool   `xml:"previousserver"`
				NextServer         bool   `xml:"nextserver"`
				PanicSimulation    bool   `xml:"panicsimulation"`
				ScanChannels       bool   `xml:"scanchannels"`
				DisplayVersion     bool   `xml:"displayversion"`
				ClearScreen        bool   `xml:"clearscreen"`
				RepeatTxLoopTest   bool   `xml:"repeattxlooptest"`
				PrintXmlConfig     bool   `xml:"printxmlconfig"`
				SendEmail          bool   `xml:"sendemail"`
				PingServers        bool   `xml:"pingservers"`
			} `xml:"api"`
			PrintVariables struct {
				PrintAccount      bool `xml:"printaccount"`
				PrintLogging      bool `xml:"printlogging"`
				PrintProvisioning bool `xml:"printprovisioning"`
				PrintBeacon       bool `xml:"printbeacon"`
				PrintTTS          bool `xml:"printtts"`
				PrintSMTP         bool `xml:"printsmtp"`
				PrintSounds       bool `xml:"printsounds"`
				PrintTxTimeout    bool `xml:"printtxtimeout"`
				PrintHTTPAPI      bool `xml:"printhttpapi"`
				PrintTargetBoard  bool `xml:"printtargetboard"`
				PrintLeds         bool `xml:"printleds"`
				PrintHeartbeat    bool `xml:"printheartbeat"`
				PrintButtons      bool `xml:"printbuttons"`
				PrintComment      bool `xml:"printcomment"`
				PrintLcd          bool `xml:"printlcd"`
				PrintOled         bool `xml:"printoled"`
				PrintGps          bool `xml:"printgps"`
				PrintTraccar      bool `xml:"printtraccar"`
				PrintPanic        bool `xml:"printpanic"`
				PrintAudioRecord  bool `xml:"printaudiorecord"`
				PrintMQTT         bool `xml:"printmqtt"`
				PrintKeyboardMap  bool `xml:"printkeyboardmap"`
				PrintUSBKeyboard  bool `xml:"printusbkeyboard"`
			} `xml:"printvariables"`
			MQTT struct {
				MQTTEnabled   bool   `xml:"enabled,attr"`
				MQTTTopic     string `xml:"mqtttopic"`
				MQTTBroker    string `xml:"mqttbroker"`
				MQTTPassword  string `xml:"mqttpassword"`
				MQTTUser      string `xml:"mqttuser"`
				MQTTId        string `xml:"mqttid"`
				MQTTCleansess bool   `xml:"cleansess"`
				MQTTQos       int    `xml:"qos"`
				MQTTNum       int    `xml:"num"`
				MQTTPayload   string `xml:"payload"`
				MQTTAction    string `xml:"action"`
				MQTTStore     string `xml:"store"`
			} `xml:"mqtt"`
		} `xml:"software"`
		Hardware struct {
			TargetBoard string `xml:"targetboard,attr"`
			Lights      struct {
				LedStripEnabled     bool   `xml:"ledstripenabled"`
				VoiceActivityLedPin string `xml:"voiceactivityledpin"`
				ParticipantsLedPin  string `xml:"participantsledpin"`
				TransmitLedPin      string `xml:"transmitledpin"`
				OnlineLedPin        string `xml:"onlineledpin"`
				AttentionLedPin     string `xml:"attentionledpin"`
			} `xml:"lights"`
			HeartBeat struct {
				Enabled     bool   `xml:"enabled,attr"`
				LEDPin      string `xml:"heartbeatledpin"`
				Periodmsecs int    `xml:"periodmsecs"`
				LEDOnmsecs  int    `xml:"ledonmsecs"`
				LEDOffmsecs int    `xml:"ledoffmsecs"`
			} `xml:"heartbeat"`
			Buttons struct {
				TxButtonPin     string `xml:"txbuttonpin"`
				TxTogglePin     string `xml:"txtogglepin"`
				UpButtonPin     string `xml:"upbuttonpin"`
				DownButtonPin   string `xml:"downbuttonpin"`
				PanicButtonPin  string `xml:"panicbuttonpin"`
				StreamButtonPin string `xml:"streambuttonpin"`
			} `xml:"buttons"`
			Comment struct {
				CommentButtonPin  string `xml:"commentbuttonpin"`
				CommentMessageOff string `xml:"commentmessageoff"`
				CommentMessageOn  string `xml:"commentmessageon"`
			} `xml:"comment"`
			LCD struct {
				Enabled               bool   `xml:"enabled,attr"`
				InterfaceType         string `xml:"lcdinterfacetype"`
				I2CAddress            uint8  `xml:"lcdi2caddress"`
				BacklightTimerEnabled bool   `xml:"lcdbacklighttimerenabled"`
				BackLightTimeoutSecs  int    `xml:"LCDBackLightTimeout"`
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
			} `xml:"gps"`
			GPSTrackingFunction struct {
				TrackEnabled           bool   `xml:"enabled,attr"`
				TraccarSendTo          bool   `xml:"traccarsendto"`
				TraccarServerURL       string `xml:"traccarserverurl"`
				TraccarServerIP        string `xml:"traccarserverip"`
				TraccarClientId        string `xml:"traccarclientid"`
				TraccarReportFrequency int64  `xml:"traccarreportfrequency"`
				TraccarProto           string `xml:"traccarproto"`
				TraccarServerFullURL   string `xml:"traccarserverfullurl"`
				TrackGPSShowLCD        bool   `xml:"trackgpsshowlcd"`
				TrackVerbose           bool   `xml:"trackverbose"`
			} `xml:"gpstrackingfunction"`
			PanicFunction struct {
				Enabled              bool    `xml:"enabled,attr"`
				FilenameAndPath      string  `xml:"filenameandpath"`
				Volume               float32 `xml:"volume"`
				SendIdent            bool    `xml:"sendident"`
				Message              string  `xml:"panicmessage"`
				PMailEnabled         bool    `xml:"panicemail"`
				PEavesdropEnabled    bool    `xml:"eavesdrop"`
				RecursiveSendMessage string  `xml:"recursivesendmessage"`
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
			KeyboardCommands struct {
				Command []struct {
					Name    string `xml:"name,attr"`
					Enabled bool   `xml:"enabled,attr"`
					Params  struct {
						Param []struct {
							Name  string `xml:"name,attr"`
							Value uint32 `xml:"value,attr"`
						} `xml:"param"`
					} `xml:"params"`
					Ttykeyboard struct {
						Scanid   rune   `xml:"scanid,attr"`
						Enabled  bool   `xml:"enabled,attr"`
						Keylabel uint32 `xml:"keylabel"`
					} `xml:"ttykeyboard"`
					Usbkeyboard struct {
						Scanid   rune   `xml:"scanid,attr"`
						Enabled  bool   `xml:"enabled,attr"`
						Keylabel uint32 `xml:"keylabel"`
					} `xml:"usbkeyboard"`
				} `xml:"command"`
			} `xml:"keyboardcommands"`
		} `xml:"hardware"`
	} `xml:"global"`
}

type VTStruct struct {
	ID []struct {
		Value uint32
		Users struct {
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

type TTYKBStruct struct {
	Enabled    bool
	KeyLabel   uint32
	Command    string
	ParamName  string
	ParamValue uint32
}

type USBKBStruct struct {
	Enabled    bool
	KeyLabel   uint32
	Command    string
	ParamName  string
	ParamValue uint32
}

func readxmlconfig(file string) error {
	xmlFile, err := os.Open(file)
	if err != nil {
		return fmt.Errorf(err.Error())
	}
	log.Println("info: Successfully Read file " + filepath.Base(file))
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	err = xml.Unmarshal(byteValue, &Document)
	if err != nil {
		return fmt.Errorf(filepath.Base(file) + " " + err.Error())
	}
	for _, account := range Document.Accounts.Account {
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
			AccountCount++
		}
	}

	if AccountCount == 0 {
		FatalCleanUp("No Default Accounts Found in talkkonnect.xml File! Please Add At Least 1 Account in XML")
	}

	for _, KMainCommands := range Document.Global.Hardware.KeyboardCommands.Command {
		if KMainCommands.Enabled {
			for _, KSubCommands := range KMainCommands.Params.Param {
				if KMainCommands.Ttykeyboard.Enabled {
					TTYKeyMap[KMainCommands.Ttykeyboard.Scanid] = TTYKBStruct{KMainCommands.Ttykeyboard.Enabled, KMainCommands.Ttykeyboard.Keylabel, KMainCommands.Name, KSubCommands.Name, KSubCommands.Value}
				}
				if KMainCommands.Usbkeyboard.Enabled {
					USBKeyMap[KMainCommands.Usbkeyboard.Scanid] = USBKBStruct{KMainCommands.Usbkeyboard.Enabled, KMainCommands.Usbkeyboard.Keylabel, KMainCommands.Name, KSubCommands.Name, KSubCommands.Value}
				}
			}
		}
	}

	// insert the voice target back here

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

	// Set our default sharefile path
	defaultSharePath := "/tmp"
	dir := filepath.Dir(exec)
	//Check for soundfiles directory in various locations
	// First, check env for $GOPATH and check in the hardcoded talkkonnect/talkkonnect dir
	if os.Getenv("GOPATH") != "" {
		defaultRepo := os.Getenv("GOPATH") + "/src/github.com/talkkonnect/talkkonnect"
		if stat, err := os.Stat(defaultRepo); err == nil && stat.IsDir() {
			defaultSharePath = defaultRepo
		}
	}
	// Next, check the same dir as executable for 'soundfiles'
	if stat, err := os.Stat(dir + "/soundfiles"); err == nil && stat.IsDir() {
		defaultSharePath = dir
	}
	// Last, if its in a bin directory, we check for ../share/talkkonnect/ and prioritize it if it exists
	if strings.HasSuffix(dir, "bin") {
		shareDir := filepath.Dir(dir) + "/share/" + filepath.Base(exec)
		if stat, err := os.Stat(shareDir); err == nil && stat.IsDir() {
			defaultSharePath = shareDir
		}
	}

	OutputDevice = Document.Global.Software.Settings.OutputDevice
	OutputDeviceShort = Document.Global.Software.Settings.OutputDeviceShort

	if len(OutputDeviceShort) == 0 {
		OutputDeviceShort = Document.Global.Software.Settings.OutputDevice
	}

	LogFilenameAndPath = Document.Global.Software.Settings.LogFilenameAndPath
	Logging = Document.Global.Software.Settings.Logging

	if Document.Global.Software.Settings.Loglevel == "trace" || Document.Global.Software.Settings.Loglevel == "debug" || Document.Global.Software.Settings.Loglevel == "info" || Document.Global.Software.Settings.Loglevel == "warning" || Document.Global.Software.Settings.Loglevel == "error" || Document.Global.Software.Settings.Loglevel == "alert" {
		Loglevel = Document.Global.Software.Settings.Loglevel
	}

	if strings.ToLower(Logging) != "screen" && LogFilenameAndPath == "" {
		LogFilenameAndPath = defaultLogPath
	}

	Daemonize = Document.Global.Software.Settings.Daemonize

	CancellableStream = Document.Global.Software.Settings.CancellableStream
	StreamOnStart = Document.Global.Software.Settings.StreamOnStart

	SimplexWithMute = Document.Global.Software.Settings.SimplexWithMute
	TxCounter = Document.Global.Software.Settings.TxCounter
	NextServerIndex = Document.Global.Software.Settings.NextServerIndex

	APEnabled = Document.Global.Software.AutoProvisioning.Enabled
	TkID = Document.Global.Software.AutoProvisioning.TkID
	URL = Document.Global.Software.AutoProvisioning.URL
	SaveFilePath = Document.Global.Software.AutoProvisioning.SaveFilePath
	SaveFilename = Document.Global.Software.AutoProvisioning.SaveFilename

	if APEnabled && SaveFilePath == "" {
		SaveFilePath = defaultConfPath
	}

	if APEnabled && SaveFilename == "" {
		SaveFilename = filepath.Base(exec) + "talkkonnect.xml"
	}

	BeaconEnabled = Document.Global.Software.Beacon.Enabled
	BeaconTimerSecs = Document.Global.Software.Beacon.BeaconTimerSecs
	BeaconFilenameAndPath = Document.Global.Software.Beacon.BeaconFileAndPath
	if BeaconEnabled && BeaconFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/Beacon.wav"
		if _, err := os.Stat(path); err == nil {
			BeaconFilenameAndPath = path
		}
	}

	BVolume = Document.Global.Software.Beacon.Volume

	TTSEnabled = Document.Global.Software.TTS.Enabled
	TTSVolumeLevel = Document.Global.Software.TTS.VolumeLevel
	TTSParticipants = Document.Global.Software.TTS.Participants
	TTSChannelUp = Document.Global.Software.TTS.ChannelUp
	TTSChannelUpFilenameAndPath = Document.Global.Software.TTS.ChannelUpFilenameAndPath

	if TTSChannelUp && TTSChannelUpFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/ChannelUp.wav"
		if _, err := os.Stat(path); err == nil {
			TTSChannelUpFilenameAndPath = path
		}
	}

	TTSChannelUpFilenameAndPath = Document.Global.Software.TTS.ChannelUpFilenameAndPath
	TTSChannelDown = Document.Global.Software.TTS.ChannelDown
	TTSChannelDownFilenameAndPath = Document.Global.Software.TTS.ChannelDownFilenameAndPath

	if TTSChannelDown && TTSChannelDownFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/ChannelDown.wav"
		if _, err := os.Stat(path); err == nil {
			TTSChannelDownFilenameAndPath = path
		}
	}

	TTSMuteUnMuteSpeaker = Document.Global.Software.TTS.MuteUnmuteSpeaker
	TTSMuteUnMuteSpeakerFilenameAndPath = Document.Global.Software.TTS.MuteUnmuteSpeakerFilenameAndPath

	if TTSMuteUnMuteSpeaker && TTSMuteUnMuteSpeakerFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/MuteUnMuteSpeaker.wav"
		if _, err := os.Stat(path); err == nil {
			TTSMuteUnMuteSpeakerFilenameAndPath = path
		}
	}

	TTSCurrentVolumeLevel = Document.Global.Software.TTS.CurrentVolumeLevel
	TTSCurrentVolumeLevelFilenameAndPath = Document.Global.Software.TTS.CurrentVolumeLevelFilenameAndPath

	if TTSCurrentVolumeLevel && TTSCurrentVolumeLevelFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/CurrentVolumeLevel.wav"
		if _, err := os.Stat(path); err == nil {
			TTSCurrentVolumeLevelFilenameAndPath = path
		}
	}

	TTSDigitalVolumeUp = Document.Global.Software.TTS.DigitalVolumeUp
	TTSDigitalVolumeUpFilenameAndPath = Document.Global.Software.TTS.DigitalVolumeUpFilenameAndPath

	if TTSDigitalVolumeUp && TTSDigitalVolumeUpFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/DigitalVolumeUp.wav"
		if _, err := os.Stat(path); err == nil {
			TTSDigitalVolumeUpFilenameAndPath = path
		}
	}

	TTSDigitalVolumeDown = Document.Global.Software.TTS.DigitalVolumeDown
	TTSDigitalVolumeDownFilenameAndPath = Document.Global.Software.TTS.DigitalVolumeDownFilenameAndPath

	if TTSDigitalVolumeDown && TTSDigitalVolumeDownFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/DigitalVolumeDown.wav"
		if _, err := os.Stat(path); err == nil {
			TTSDigitalVolumeDownFilenameAndPath = path
		}
	}

	TTSListServerChannels = Document.Global.Software.TTS.ListServerChannels
	TTSListServerChannelsFilenameAndPath = Document.Global.Software.TTS.ListServerChannelsFilenameAndPath

	if TTSListServerChannels && TTSListServerChannelsFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/ListServerChannels.wav"
		if _, err := os.Stat(path); err == nil {
			TTSListServerChannelsFilenameAndPath = path
		}
	}

	TTSStartTransmitting = Document.Global.Software.TTS.StartTransmitting
	TTSStartTransmittingFilenameAndPath = Document.Global.Software.TTS.StartTransmittingFilenameAndPath

	if TTSStartTransmitting && TTSStartTransmittingFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/StartTransmitting.wav"
		if _, err := os.Stat(path); err == nil {
			TTSStartTransmittingFilenameAndPath = path
		}
	}

	TTSStopTransmitting = Document.Global.Software.TTS.StopTransmitting
	TTSStopTransmittingFilenameAndPath = Document.Global.Software.TTS.StopTransmittingFilenameAndPath

	if TTSStopTransmitting && TTSStopTransmittingFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/StopTransmitting.wav"
		if _, err := os.Stat(path); err == nil {
			TTSStopTransmittingFilenameAndPath = path
		}
	}

	TTSListOnlineUsers = Document.Global.Software.TTS.ListOnlineUsers
	TTSListOnlineUsersFilenameAndPath = Document.Global.Software.TTS.ListOnlineUsersFilenameAndPath

	if TTSListOnlineUsers && TTSListOnlineUsersFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/ListOnlineUsers.wav"
		if _, err := os.Stat(path); err == nil {
			TTSListOnlineUsersFilenameAndPath = path
		}
	}

	TTSPlayStream = Document.Global.Software.TTS.PlayStream
	TTSPlayStreamFilenameAndPath = Document.Global.Software.TTS.PlayStreamFilenameAndPath

	if TTSPlayStream && TTSPlayStreamFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/PlayStream.wav"
		if _, err := os.Stat(path); err == nil {
			TTSPlayStreamFilenameAndPath = path
		}
	}

	TTSRequestGpsPosition = Document.Global.Software.TTS.RequestGpsPosition
	TTSRequestGpsPositionFilenameAndPath = Document.Global.Software.TTS.RequestGpsPositionFilenameAndPath

	if TTSRequestGpsPosition && TTSRequestGpsPositionFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/RequestGpsPosition.wav"
		if _, err := os.Stat(path); err == nil {
			TTSRequestGpsPositionFilenameAndPath = path
		}
	}

	TTSNextServer = Document.Global.Software.TTS.NextServer
	TTSNextServerFilenameAndPath = Document.Global.Software.TTS.NextServerFilenameAndPath
	/*
		//TODO: No default sound available. Placeholder for now
		if TTSNextServer && TTSNextServerFilenameAndPath == "" {
			path := defaultSharePath + "/soundfiles/voiceprompts/TODO"
			if _, err := os.Stat(path); err == nil {
				TTSNextServerFilenameAndPath = path
			}
		}
	*/

	TTSPreviousServer = Document.Global.Software.TTS.PreviousServer
	TTSPreviousServerFilenameAndPath = Document.Global.Software.TTS.PreviousServerFilenameAndPath

	/*
		//TODO: No default sound available. Placeholder for now
		if TTSPreviousServer && TTSPreviousServerFilenameAndPath == "" {
			path := defaultSharePath + "/soundfiles/voiceprompts/TODO"
			if _, err := os.Stat(path); err == nil {
				TTSPreviousServerFilenameAndPath = path
			}
		}
	*/

	TTSPanicSimulation = Document.Global.Software.TTS.PanicSimulation
	TTSPanicSimulationFilenameAndPath = Document.Global.Software.TTS.PanicSimulationFilenameAndPath
	if TTSPanicSimulation && TTSPanicSimulationFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/PanicSimulation.wav"
		if _, err := os.Stat(path); err == nil {
			TTSPanicSimulationFilenameAndPath = path
		}
	}

	TTSPrintXmlConfig = Document.Global.Software.TTS.PrintXmlConfig
	TTSPrintXmlConfigFilenameAndPath = Document.Global.Software.TTS.PrintXmlConfigFilenameAndPath

	if TTSPrintXmlConfig && TTSPrintXmlConfigFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/PrintXmlConfig.wav"
		if _, err := os.Stat(path); err == nil {
			TTSPrintXmlConfigFilenameAndPath = path
		}
	}

	TTSSendEmail = Document.Global.Software.TTS.SendEmail
	TTSSendEmailFilenameAndPath = Document.Global.Software.TTS.SendEmailFilenameAndPath

	if TTSSendEmail && TTSSendEmailFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/SendEmail.wav"
		if _, err := os.Stat(path); err == nil {
			TTSSendEmailFilenameAndPath = path
		}
	}

	TTSDisplayMenu = Document.Global.Software.TTS.DisplayMenu
	TTSDisplayMenuFilenameAndPath = Document.Global.Software.TTS.DisplayMenuFilenameAndPath

	if TTSDisplayMenu && TTSDisplayMenuFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/DisplayMenu.wav"
		if _, err := os.Stat(path); err == nil {
			TTSDisplayMenuFilenameAndPath = path
		}
	}

	TTSQuitTalkkonnect = Document.Global.Software.TTS.QuitTalkkonnect
	TTSQuitTalkkonnectFilenameAndPath = Document.Global.Software.TTS.QuitTalkkonnectFilenameAndPath

	if TTSQuitTalkkonnect && TTSQuitTalkkonnectFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/QuitTalkkonnect.wav"
		if _, err := os.Stat(path); err == nil {
			TTSQuitTalkkonnectFilenameAndPath = path
		}
	}

	TTSTalkkonnectLoaded = Document.Global.Software.TTS.TalkkonnectLoaded
	TTSTalkkonnectLoadedFilenameAndPath = Document.Global.Software.TTS.TalkkonnectLoadedFilenameAndPath

	if TTSTalkkonnectLoaded && TTSTalkkonnectLoadedFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/Loaded.wav"
		if _, err := os.Stat(path); err == nil {
			TTSTalkkonnectLoadedFilenameAndPath = path
		}
	}

	TTSPingServers = Document.Global.Software.TTS.PingServers
	TTSPingServersFilenameAndPath = Document.Global.Software.TTS.PingServersFilenameAndPath

	/*
		//TODO: No default sound available. Placeholder for now
		if TTSPingServers && TTSPingServersFilenameAndPath == "" {
			path := defaultSharePath + "/soundfiles/voiceprompts/TODO"
			if _, err := os.Stat(path); err == nil {
				TTSPingServersFilenameAndPath = path
			}
		}
	*/

	EmailEnabled = Document.Global.Software.SMTP.Enabled
	EmailUsername = Document.Global.Software.SMTP.Username
	EmailPassword = Document.Global.Software.SMTP.Password
	EmailReceiver = Document.Global.Software.SMTP.Receiver
	EmailSubject = Document.Global.Software.SMTP.Subject
	EmailMessage = Document.Global.Software.SMTP.Message
	EmailGpsDateTime = Document.Global.Software.SMTP.GpsDateTime
	EmailGpsLatLong = Document.Global.Software.SMTP.GpsLatLong
	EmailGoogleMapsURL = Document.Global.Software.SMTP.GoogleMapsURL

	EventSoundEnabled = Document.Global.Software.Sounds.Event.Enabled
	EventJoinedSoundFilenameAndPath = Document.Global.Software.Sounds.Event.JoinedFilenameAndPath
	EventLeftSoundFilenameAndPath = Document.Global.Software.Sounds.Event.LeftFilenameAndPath
	EventMessageSoundFilenameAndPath = Document.Global.Software.Sounds.Event.MessageFilenameAndPath

	if EventSoundEnabled && EventJoinedSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/events/event.wav"
		if _, err := os.Stat(path); err == nil {
			EventJoinedSoundFilenameAndPath = path
		}
	}
	if EventSoundEnabled && EventLeftSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/events/event.wav"
		if _, err := os.Stat(path); err == nil {
			EventLeftSoundFilenameAndPath = path
		}
	}
	if EventSoundEnabled && EventMessageSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/events/event.wav"
		if _, err := os.Stat(path); err == nil {
			EventMessageSoundFilenameAndPath = path
		}
	}
	AlertSoundEnabled = Document.Global.Software.Sounds.Alert.Enabled
	AlertSoundFilenameAndPath = Document.Global.Software.Sounds.Alert.FilenameAndPath

	if AlertSoundEnabled && AlertSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/alerts/alert.wav"
		if _, err := os.Stat(path); err == nil {
			AlertSoundFilenameAndPath = path
		}
	}

	AlertSoundVolume = Document.Global.Software.Sounds.Alert.Volume

	IncommingBeepSoundEnabled = Document.Global.Software.Sounds.IncommingBeep.Enabled
	IncommingBeepSoundFilenameAndPath = Document.Global.Software.Sounds.IncommingBeep.FilenameAndPath

	if IncommingBeepSoundEnabled && IncommingBeepSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/rogerbeeps/Chirsp.wav"
		if _, err := os.Stat(path); err == nil {
			IncommingBeepSoundFilenameAndPath = path
		}
	}

	IncommingBeepSoundVolume = Document.Global.Software.Sounds.IncommingBeep.Volume

	RogerBeepSoundEnabled = Document.Global.Software.Sounds.RogerBeep.Enabled
	RogerBeepSoundFilenameAndPath = Document.Global.Software.Sounds.RogerBeep.FilenameAndPath

	if RogerBeepSoundEnabled && RogerBeepSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/rogerbeeps/Chirsp.wav"
		if _, err := os.Stat(path); err == nil {
			RogerBeepSoundFilenameAndPath = path
		}
	}

	RogerBeepSoundVolume = Document.Global.Software.Sounds.RogerBeep.Volume

	RepeaterToneEnabled = Document.Global.Software.Sounds.RepeaterTone.Enabled
	RepeaterToneFrequencyHz = Document.Global.Software.Sounds.RepeaterTone.ToneFrequencyHz
	RepeaterToneDurationSec = Document.Global.Software.Sounds.RepeaterTone.ToneDurationSec

	StreamSoundEnabled = Document.Global.Software.Sounds.Stream.Enabled
	StreamSoundFilenameAndPath = Document.Global.Software.Sounds.Stream.FilenameAndPath

	if StreamSoundEnabled && StreamSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/alerts/stream.wav"
		if _, err := os.Stat(path); err == nil {
			StreamSoundFilenameAndPath = path
		}
	}

	StreamSoundVolume = Document.Global.Software.Sounds.Stream.Volume

	TxTimeOutEnabled = Document.Global.Software.TxTimeOut.Enabled
	TxTimeOutSecs = Document.Global.Software.TxTimeOut.TxTimeOutSecs

	APIEnabled = Document.Global.Software.API.Enabled
	APIListenPort = Document.Global.Software.API.ListenPort
	APIDisplayMenu = Document.Global.Software.API.DisplayMenu
	APIChannelUp = Document.Global.Software.API.ChannelUp
	APIChannelDown = Document.Global.Software.API.ChannelDown
	APIMute = Document.Global.Software.API.Mute
	APICurrentVolumeLevel = Document.Global.Software.API.CurrentVolumeLevel
	APIDigitalVolumeUp = Document.Global.Software.API.DigitalVolumeUp
	APIDigitalVolumeDown = Document.Global.Software.API.DigitalVolumeDown
	APIListServerChannels = Document.Global.Software.API.ListServerChannels
	APIStartTransmitting = Document.Global.Software.API.StartTransmitting
	APIStopTransmitting = Document.Global.Software.API.StopTransmitting
	APIListOnlineUsers = Document.Global.Software.API.ListOnlineUsers
	APIPlayStream = Document.Global.Software.API.PlayStream
	APIRequestGpsPosition = Document.Global.Software.API.RequestGpsPosition
	APIEmailEnabled = Document.Global.Software.API.Enabled
	APINextServer = Document.Global.Software.API.NextServer
	APIPreviousServer = Document.Global.Software.API.PreviousServer
	APIPanicSimulation = Document.Global.Software.API.PanicSimulation
	APIDisplayVersion = Document.Global.Software.API.DisplayVersion
	APIClearScreen = Document.Global.Software.API.ClearScreen
	APIPingServersEnabled = Document.Global.Software.API.Enabled
	APIRepeatTxLoopTest = Document.Global.Software.API.RepeatTxLoopTest
	APIPrintXmlConfig = Document.Global.Software.API.PrintXmlConfig

	PrintAccount = Document.Global.Software.PrintVariables.PrintAccount
	PrintLogging = Document.Global.Software.PrintVariables.PrintLogging
	PrintProvisioning = Document.Global.Software.PrintVariables.PrintProvisioning
	PrintBeacon = Document.Global.Software.PrintVariables.PrintBeacon
	PrintTTS = Document.Global.Software.PrintVariables.PrintTTS
	PrintSMTP = Document.Global.Software.PrintVariables.PrintSMTP
	PrintSounds = Document.Global.Software.PrintVariables.PrintSounds
	PrintTxTimeout = Document.Global.Software.PrintVariables.PrintTxTimeout

	MQTTEnabled = Document.Global.Software.MQTT.MQTTEnabled
	MQTTTopic = Document.Global.Software.MQTT.MQTTTopic
	MQTTBroker = Document.Global.Software.MQTT.MQTTBroker
	MQTTPassword = Document.Global.Software.MQTT.MQTTPassword
	MQTTUser = Document.Global.Software.MQTT.MQTTUser
	MQTTId = Document.Global.Software.MQTT.MQTTId
	MQTTCleansess = Document.Global.Software.MQTT.MQTTCleansess
	MQTTQos = Document.Global.Software.MQTT.MQTTQos
	MQTTNum = Document.Global.Software.MQTT.MQTTNum
	MQTTPayload = Document.Global.Software.MQTT.MQTTPayload
	MQTTAction = Document.Global.Software.MQTT.MQTTAction
	MQTTStore = Document.Global.Software.MQTT.MQTTStore

	PrintHTTPAPI = Document.Global.Software.PrintVariables.PrintHTTPAPI
	PrintTargetboard = Document.Global.Software.PrintVariables.PrintTargetBoard
	PrintLeds = Document.Global.Software.PrintVariables.PrintLeds
	PrintHeartbeat = Document.Global.Software.PrintVariables.PrintHeartbeat
	PrintButtons = Document.Global.Software.PrintVariables.PrintButtons
	PrintComment = Document.Global.Software.PrintVariables.PrintComment
	PrintLcd = Document.Global.Software.PrintVariables.PrintLcd
	PrintOled = Document.Global.Software.PrintVariables.PrintOled
	PrintGps = Document.Global.Software.PrintVariables.PrintGps
	PrintTraccar = Document.Global.Software.PrintVariables.PrintTraccar
	PrintPanic = Document.Global.Software.PrintVariables.PrintPanic
	PrintAudioRecord = Document.Global.Software.PrintVariables.PrintAudioRecord
	PrintMQTT = Document.Global.Software.PrintVariables.PrintMQTT
	PrintKeyboardMap = Document.Global.Software.PrintVariables.PrintKeyboardMap
	PrintUSBKeyboard = Document.Global.Software.PrintVariables.PrintUSBKeyboard
	TargetBoard = Document.Global.Hardware.TargetBoard
	LedStripEnabled = Document.Global.Hardware.Lights.LedStripEnabled
	// my stupid work around for null uint xml unmarshelling problem with numbers so use strings and convert it 2 times
	temp0, _ := strconv.ParseUint(Document.Global.Hardware.Lights.VoiceActivityLedPin, 10, 64)
	VoiceActivityLEDPin = uint(temp0)
	temp1, _ := strconv.ParseUint(Document.Global.Hardware.Lights.VoiceActivityLedPin, 10, 64)
	VoiceActivityLEDPin = uint(temp1)
	temp2, _ := strconv.ParseUint(Document.Global.Hardware.Lights.ParticipantsLedPin, 10, 64)
	ParticipantsLEDPin = uint(temp2)
	temp3, _ := strconv.ParseUint(Document.Global.Hardware.Lights.TransmitLedPin, 10, 64)
	TransmitLEDPin = uint(temp3)
	temp4, _ := strconv.ParseUint(Document.Global.Hardware.Lights.OnlineLedPin, 10, 64)
	OnlineLEDPin = uint(temp4)
	temp14, _ := strconv.ParseUint(Document.Global.Hardware.Lights.AttentionLedPin, 10, 64)
	AttentionLEDPin = uint(temp14)

	temp5, _ := strconv.ParseUint(Document.Global.Hardware.HeartBeat.LEDPin, 10, 64)
	HeartBeatLEDPin = uint(temp5)
	HeartBeatEnabled = Document.Global.Hardware.HeartBeat.Enabled
	PeriodmSecs = Document.Global.Hardware.HeartBeat.Periodmsecs
	LEDOnmSecs = Document.Global.Hardware.HeartBeat.LEDOnmsecs
	LEDOffmSecs = Document.Global.Hardware.HeartBeat.LEDOffmsecs

	// my stupid work around for null uint xml unmarshelling problem with numbers so use strings and convert it 2 times
	temp6, _ := strconv.ParseUint(Document.Global.Hardware.Buttons.TxButtonPin, 10, 64)
	TxButtonPin = uint(temp6)
	temp7, _ := strconv.ParseUint(Document.Global.Hardware.Buttons.TxTogglePin, 10, 64)
	TxTogglePin = uint(temp7)
	temp8, _ := strconv.ParseUint(Document.Global.Hardware.Buttons.UpButtonPin, 10, 64)
	UpButtonPin = uint(temp8)
	temp9, _ := strconv.ParseUint(Document.Global.Hardware.Buttons.DownButtonPin, 10, 64)
	DownButtonPin = uint(temp9)
	temp10, _ := strconv.ParseUint(Document.Global.Hardware.Buttons.PanicButtonPin, 10, 64)
	PanicButtonPin = uint(temp10)
	temp11, _ := strconv.ParseUint(Document.Global.Hardware.Comment.CommentButtonPin, 10, 64)
	CommentButtonPin = uint(temp11)
	CommentMessageOff = Document.Global.Hardware.Comment.CommentMessageOff
	CommentMessageOn = Document.Global.Hardware.Comment.CommentMessageOn
	temp12, _ := strconv.ParseUint(Document.Global.Hardware.Buttons.StreamButtonPin, 10, 64)
	StreamButtonPin = uint(temp12)

	LCDEnabled = Document.Global.Hardware.LCD.Enabled
	LCDInterfaceType = Document.Global.Hardware.LCD.InterfaceType
	LCDI2CAddress = Document.Global.Hardware.LCD.I2CAddress
	LCDBackLightTimerEnabled = Document.Global.Hardware.LCD.Enabled
	LCDBackLightTimeout = time.Duration(Document.Global.Hardware.LCD.BackLightTimeoutSecs)

	// my stupid work around for null uint xml unmarshelling problem with numbers so use strings and convert it 2 times
	temp13, _ := strconv.ParseUint(Document.Global.Hardware.LCD.BackLightLEDPin, 10, 64)
	LCDBackLightLEDPin = int(temp13)

	LCDRSPin = Document.Global.Hardware.LCD.RsPin
	LCDEPin = Document.Global.Hardware.LCD.EPin
	LCDD4Pin = Document.Global.Hardware.LCD.D4Pin
	LCDD5Pin = Document.Global.Hardware.LCD.D5Pin
	LCDD6Pin = Document.Global.Hardware.LCD.D6Pin
	LCDD7Pin = Document.Global.Hardware.LCD.D7Pin

	OLEDEnabled = Document.Global.Hardware.OLED.Enabled
	OLEDInterfacetype = Document.Global.Hardware.OLED.InterfaceType
	OLEDDisplayRows = Document.Global.Hardware.OLED.DisplayRows
	OLEDDisplayColumns = Document.Global.Hardware.OLED.DisplayColumns
	OLEDDefaultI2cBus = Document.Global.Hardware.OLED.DefaultI2CBus
	OLEDDefaultI2cAddress = Document.Global.Hardware.OLED.DefaultI2CAddress
	OLEDScreenWidth = Document.Global.Hardware.OLED.ScreenWidth
	OLEDScreenHeight = Document.Global.Hardware.OLED.ScreenHeight
	OLEDCommandColumnAddressing = Document.Global.Hardware.OLED.CommandColumnAddressing
	OLEDAddressBasePageStart = Document.Global.Hardware.OLED.AddressBasePageStart
	OLEDCharLength = Document.Global.Hardware.OLED.CharLength
	OLEDStartColumn = Document.Global.Hardware.OLED.StartColumn

	GpsEnabled = Document.Global.Hardware.GPS.Enabled
	Port = Document.Global.Hardware.GPS.Port
	Baud = Document.Global.Hardware.GPS.Baud
	TxData = Document.Global.Hardware.GPS.TxData
	Even = Document.Global.Hardware.GPS.Even
	Odd = Document.Global.Hardware.GPS.Odd
	Rs485 = Document.Global.Hardware.GPS.Rs485
	Rs485HighDuringSend = Document.Global.Hardware.GPS.Rs485HighDuringSend
	Rs485HighAfterSend = Document.Global.Hardware.GPS.Rs485HighAfterSend
	StopBits = Document.Global.Hardware.GPS.StopBits
	DataBits = Document.Global.Hardware.GPS.DataBits
	CharTimeOut = Document.Global.Hardware.GPS.CharTimeOut
	MinRead = Document.Global.Hardware.GPS.MinRead
	Rx = Document.Global.Hardware.GPS.Rx
	GpsInfoVerbose = Document.Global.Hardware.GPS.GpsInfoVerbose
	TrackEnabled = Document.Global.Hardware.GPSTrackingFunction.TrackEnabled
	TraccarSendTo = Document.Global.Hardware.GPSTrackingFunction.TraccarSendTo
	TraccarServerURL = Document.Global.Hardware.GPSTrackingFunction.TraccarServerURL
	TraccarServerIP = Document.Global.Hardware.GPSTrackingFunction.TraccarServerIP
	TraccarClientId = Document.Global.Hardware.GPSTrackingFunction.TraccarClientId
	TraccarReportFrequency = Document.Global.Hardware.GPSTrackingFunction.TraccarReportFrequency
	TraccarProto = Document.Global.Hardware.GPSTrackingFunction.TraccarProto
	TraccarServerFullURL = Document.Global.Hardware.GPSTrackingFunction.TraccarServerFullURL
	TrackGPSShowLCD = Document.Global.Hardware.GPSTrackingFunction.TrackGPSShowLCD
	TrackVerbose = Document.Global.Hardware.GPSTrackingFunction.TrackVerbose
	PEnabled = Document.Global.Hardware.PanicFunction.Enabled
	PFilenameAndPath = Document.Global.Hardware.PanicFunction.FilenameAndPath

	if PEnabled && PFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/alerts/alert.wav"
		if _, err := os.Stat(path); err == nil {
			PFilenameAndPath = path
		}
	}

	PMessage = Document.Global.Hardware.PanicFunction.Message
	PMailEnabled = Document.Global.Hardware.PanicFunction.PMailEnabled
	PVolume = Document.Global.Hardware.PanicFunction.Volume
	PSendIdent = Document.Global.Hardware.PanicFunction.SendIdent
	PSendGpsLocation = Document.Global.Hardware.PanicFunction.SendGpsLocation
	PTxLockEnabled = Document.Global.Hardware.PanicFunction.TxLockEnabled
	PTxlockTimeOutSecs = Document.Global.Hardware.PanicFunction.TxLockTimeOutSecs
	PLowProfile = Document.Global.Hardware.PanicFunction.PLowProfile
	USBKeyboardEnabled = Document.Global.Hardware.USBKeyboard.Enabled
	USBKeyboardPath = Document.Global.Hardware.USBKeyboard.USBKeyboardPath
	NumlockScanID = Document.Global.Hardware.USBKeyboard.NumlockScanID
	AudioRecordEnabled = Document.Global.Hardware.AudioRecordFunction.Enabled
	AudioRecordOnStart = Document.Global.Hardware.AudioRecordFunction.RecordOnStart
	AudioRecordSystem = Document.Global.Hardware.AudioRecordFunction.RecordSystem
	AudioRecordMode = Document.Global.Hardware.AudioRecordFunction.RecordMode
	AudioRecordTimeout = Document.Global.Hardware.AudioRecordFunction.RecordTimeout
	AudioRecordFromOutput = Document.Global.Hardware.AudioRecordFunction.RecordFromOutput
	AudioRecordFromInput = Document.Global.Hardware.AudioRecordFunction.RecordFromInput
	AudioRecordMicTimeout = Document.Global.Hardware.AudioRecordFunction.RecordMicTimeout
	AudioRecordSavePath = Document.Global.Hardware.AudioRecordFunction.RecordSavePath
	AudioRecordArchivePath = Document.Global.Hardware.AudioRecordFunction.RecordArchivePath
	AudioRecordSoft = Document.Global.Hardware.AudioRecordFunction.RecordSoft
	AudioRecordProfile = Document.Global.Hardware.AudioRecordFunction.RecordProfile
	AudioRecordFileFormat = Document.Global.Hardware.AudioRecordFunction.RecordFileFormat
	AudioRecordChunkSize = Document.Global.Hardware.AudioRecordFunction.RecordChunkSize

	if TargetBoard != "rpi" {
		LCDBackLightTimerEnabled = false
	}

	if LCDBackLightTimerEnabled && (!OLEDEnabled && !LCDEnabled) {
		FatalCleanUp("Alert: Logical Error in LCDBacklight Timer Check XML config file. Backlight Timer Enabled but both LCD and OLED disabled!")
	}

	if OLEDEnabled {
		Oled, err = goled.BeginOled(OLEDDefaultI2cAddress, OLEDDefaultI2cBus, OLEDScreenWidth, OLEDScreenHeight, OLEDDisplayRows, OLEDDisplayColumns, OLEDStartColumn, OLEDCharLength, OLEDCommandColumnAddressing, OLEDAddressBasePageStart)
		if err != nil {
			log.Println("error: Cannot Communicate with OLED")
		}
	}

	log.Println("Successfully loaded XML configuration file into memory")

	for i := 0; i < len(Document.Accounts.Account); i++ {
		if Document.Accounts.Account[i].Default {
			log.Printf("info: Successfully Added Account %s to Index [%d]\n", Document.Accounts.Account[i].Name, i)
		}
	}

	return nil
}

func printxmlconfig() {

	if PrintAccount {
		log.Println("info: ---------- Account Information -------- ")
		log.Println("info: Default              ", fmt.Sprintf("%t", Default))
		log.Println("info: Server               ", Server[AccountIndex])
		log.Println("info: Username             ", Username[AccountIndex])
		log.Println("info: Password             ", Password[AccountIndex])
		log.Println("info: Insecure             ", fmt.Sprintf("%t", Insecure[AccountIndex]))
		log.Println("info: Register             ", fmt.Sprintf("%t", Register[AccountIndex]))
		log.Println("info: Certificate          ", Certificate[AccountIndex])
		log.Println("info: Channel              ", Channel[AccountIndex])
		log.Println("info: Ident                ", Ident[AccountIndex])
		log.Println("info: Tokens               ", Tokens[AccountIndex])
		log.Println("info: VoiceTargets         ", VT[AccountIndex])

	} else {
		log.Println("info: ---------- Account Information -------- SKIPPED ")
	}

	if PrintLogging {
		log.Println("info: -------- Logging & Daemonizing -------- ")
		log.Println("info: Output Device        ", OutputDevice)
		log.Println("info: Output Device(Short) ", OutputDeviceShort)
		log.Println("info: Log File             ", LogFilenameAndPath)
		log.Println("info: Logging              ", Logging)
		log.Println("info: Loglevel             ", Loglevel)
		log.Println("info: Daemonize            ", fmt.Sprintf("%t", Daemonize))
		log.Println("info: CancellableStream    ", fmt.Sprintf("%t", CancellableStream))
		log.Println("info: StreamOnStart        ", fmt.Sprintf("%t", StreamOnStart))
		log.Println("info: SimplexWithMute      ", fmt.Sprintf("%t", SimplexWithMute))
		log.Println("info: TxCounter            ", fmt.Sprintf("%t", TxCounter))
		log.Println("info: NextServerIndex      ", fmt.Sprintf("%v", NextServerIndex))
	} else {
		log.Println("info: --------   Logging & Daemonizing -------- SKIPPED ")
	}

	if PrintProvisioning {
		log.Println("info: --------   AutoProvisioning   --------- ")
		log.Println("info: AutoProvisioning Enabled    " + fmt.Sprintf("%t", APEnabled))
		log.Println("info: Talkkonned ID (tkid)        " + TkID)
		log.Println("info: AutoProvisioning Server URL " + URL)
		log.Println("info: Config Local Path           " + SaveFilePath)
		log.Println("info: Config Local Filename       " + SaveFilename)
	} else {
		log.Println("info: --------   AutoProvisioning   --------- SKIPPED ")
	}

	if PrintBeacon {
		log.Println("info: --------  Beacon   --------- ")
		log.Println("info: Beacon Enabled         " + fmt.Sprintf("%t", BeaconEnabled))
		log.Println("info: Beacon Time (Secs)     " + fmt.Sprintf("%v", BeaconTimerSecs))
		log.Println("info: Beacon Filename & Path " + BeaconFilenameAndPath)
		log.Println("info: Beacon Playback Volume " + fmt.Sprintf("%v", BVolume))
	} else {
		log.Println("info: --------   Beacon   --------- SKIPPED ")
	}

	if PrintTTS {
		log.Println("info: -------- TTS  -------- ")
		log.Println("info: TTS Global Enabled     ", fmt.Sprintf("%t", TTSEnabled))
		log.Println("info: TTS Volume Level (%)   ", fmt.Sprintf("%d", TTSVolumeLevel))
		log.Println("info: TTS Participants       ", fmt.Sprintf("%t", TTSParticipants))
		log.Println("info: TTS ChannelUp          ", fmt.Sprintf("%t", TTSChannelUp))
		log.Println("info: TTS ChannelUpFilenameAndPath ", TTSChannelUpFilenameAndPath)
		log.Println("info: TTS ChannelDown        ", fmt.Sprintf("%t", TTSChannelDown))
		log.Println("info: TTS ChannelDownFilenameAndPath  ", TTSChannelDownFilenameAndPath)
		log.Println("info: TTS MuteUnMuteSpeaker  ", fmt.Sprintf("%t", TTSMuteUnMuteSpeaker))
		log.Println("info: TTS MuteUnMuteSpeakerFilenameAndPath ", TTSMuteUnMuteSpeakerFilenameAndPath)
		log.Println("info: TTS CurrentVolumeLevel ", fmt.Sprintf("%t", TTSCurrentVolumeLevel))
		log.Println("info: TTS CurrentVolumeLevelFilenameAndPath ", TTSCurrentVolumeLevelFilenameAndPath)
		log.Println("info: TTS DigitalVolumeUp    ", fmt.Sprintf("%t", TTSDigitalVolumeUp))
		log.Println("info: TTS DigitalVolumeUpFilenameAndPath ", TTSDigitalVolumeUpFilenameAndPath)
		log.Println("info: TTS DigitalVolumeDown  ", fmt.Sprintf("%t", TTSDigitalVolumeDown))
		log.Println("info: TTS DigitalVolumeDownFilenameAndPath ", TTSDigitalVolumeDownFilenameAndPath)
		log.Println("info: TTS ListServerChannels ", fmt.Sprintf("%t", TTSListServerChannels))
		log.Println("info: TTS ListServerChannelsFilenameAndPath  ", TTSListServerChannelsFilenameAndPath)
		log.Println("info: TTS StartTransmitting  ", fmt.Sprintf("%t", TTSStartTransmitting))
		log.Println("info: TTS StartTransmittingFilenameAndPath ", TTSStartTransmittingFilenameAndPath)
		log.Println("info: TTS StopTransmitting   ", fmt.Sprintf("%t", TTSStopTransmitting))
		log.Println("info: TTS StopTransmittingFilenameAndPath ", TTSStopTransmittingFilenameAndPath)
		log.Println("info: TTS ListOnlineUsers    ", fmt.Sprintf("%t", TTSListOnlineUsers))
		log.Println("info: TTS ListOnlineUsersFilenameAndPath ", TTSListOnlineUsersFilenameAndPath)
		log.Println("info: TTS PlayStream         ", fmt.Sprintf("%t", TTSPlayStream))
		log.Println("info: TTS PlayStreamFilenameAndPath ", TTSPlayStreamFilenameAndPath)
		log.Println("info: TTS RequestGpsPosition ", fmt.Sprintf("%t", TTSRequestGpsPosition))
		log.Println("info: TTS RequestGpsPositionFilenameAndPath ", TTSRequestGpsPositionFilenameAndPath)
		log.Println("info: TTS NextServer         ", fmt.Sprintf("%t", TTSNextServer))
		log.Println("info: TTS NextServerFilenameAndPath         ", TTSNextServerFilenameAndPath)
		log.Println("info: TTS PreviousServer     ", fmt.Sprintf("%t", TTSPreviousServer))
		log.Println("info: TTS PreviousServerFilenameAndPath  ", TTSPreviousServerFilenameAndPath)
		log.Println("info: TTS PanicSimulation    ", fmt.Sprintf("%t", TTSPanicSimulation))
		log.Println("info: TTS PanicSimulationFilenameAndPath ", TTSPanicSimulationFilenameAndPath)
		log.Println("info: TTS PrintXmlConfig     ", fmt.Sprintf("%t", TTSPrintXmlConfig))
		log.Println("info: TTS PrintXmlConfigFilenameAndPath ", TTSPrintXmlConfigFilenameAndPath)
		log.Println("info: TTS SendEmail          ", fmt.Sprintf("%t", TTSSendEmail))
		log.Println("info: TTS SendEmailFilenameAndPath ", TTSSendEmailFilenameAndPath)
		log.Println("info: TTS DisplayMenu        ", fmt.Sprintf("%t", TTSDisplayMenu))
		log.Println("info: TTS DisplayMenuFilenameAndPath ", TTSDisplayMenuFilenameAndPath)
		log.Println("info: TTS QuitTalkkonnect    ", fmt.Sprintf("%t", TTSQuitTalkkonnect))
		log.Println("info: TTS QuitTalkkonnectFilenameAndPath ", TTSQuitTalkkonnectFilenameAndPath)
		log.Println("info: TTS TalkkonnectLoaded  ", fmt.Sprintf("%t", TTSTalkkonnectLoaded))
		log.Println("info: TTS TalkkonnectLoadedFilenameAndPath ", TTSTalkkonnectLoadedFilenameAndPath)
		log.Println("info: TTS TalkkonnectLoaded  " + fmt.Sprintf("%t", TTSTalkkonnectLoaded))
		log.Println("info: TTS PingServersFilenameAndPath ", TTSPingServersFilenameAndPath)
		log.Println("info: TTS PingServers " + fmt.Sprintf("%t", TTSPingServers))
	} else {
		log.Println("info: --------   TTS  -------- SKIPPED ")
	}

	if PrintSMTP {
		log.Println("info: --------  Gmail SMTP Settings  -------- ")
		log.Println("info: Email Enabled   " + fmt.Sprintf("%t", EmailEnabled))
		log.Println("info: Username        " + EmailUsername)
		log.Println("info: Password        " + EmailPassword)
		log.Println("info: Receiver        " + EmailReceiver)
		log.Println("info: Subject         " + EmailSubject)
		log.Println("info: Message         " + EmailMessage)
		log.Println("info: GPS Date/Time   " + fmt.Sprintf("%t", EmailGpsDateTime))
		log.Println("info: GPS Lat/Long    " + fmt.Sprintf("%t", EmailGpsLatLong))
		log.Println("info: Google Maps URL " + fmt.Sprintf("%t", EmailGoogleMapsURL))
	} else {
		log.Println("info: --------   Gmail SMTP Settings  -------- SKIPPED ")
	}

	if PrintSounds {
		log.Println("info: ------------- Sounds  ------------------ ")
		log.Println("info: Event Sound Enabled         " + fmt.Sprintf("%t", EventSoundEnabled))
		log.Println("info: Event Joined Sound Filename " + EventJoinedSoundFilenameAndPath)
		log.Println("info: Event Left Sound Filename   " + EventJoinedSoundFilenameAndPath)
		log.Println("info: Event Msg Sound Filename    " + EventMessageSoundFilenameAndPath)
		log.Println("info: Alert Sound Enabled         " + fmt.Sprintf("%t", AlertSoundEnabled))
		log.Println("info: Alert Sound Filename        " + AlertSoundFilenameAndPath)
		log.Println("info: Alert Sound Volume          " + fmt.Sprintf("%v", AlertSoundVolume))
		log.Println("info: Incoming Beep Enabled       " + fmt.Sprintf("%t", IncommingBeepSoundEnabled))
		log.Println("info: Incoming Beep File          " + IncommingBeepSoundFilenameAndPath)
		log.Println("info: Incoming Beep Volume        " + fmt.Sprintf("%v", IncommingBeepSoundVolume))
		log.Println("info: Roger Beep Enabled         " + fmt.Sprintf("%t", RogerBeepSoundEnabled))
		log.Println("info: Roger Beep File            " + RogerBeepSoundFilenameAndPath)
		log.Println("info: Roger Beep Volume          " + fmt.Sprintf("%v", RogerBeepSoundVolume))
		log.Println("info: Repeater Tone Enabled      " + fmt.Sprintf("%t", RepeaterToneEnabled))
		log.Println("info: Repeater Tone Freq (Hz)    " + fmt.Sprintf("%v", RepeaterToneFrequencyHz))
		log.Println("info: Repeater Tone Length (Sec) " + fmt.Sprintf("%v", RepeaterToneDurationSec))
		log.Println("info: Stream Enabled             " + fmt.Sprintf("%t", StreamSoundEnabled))
		log.Println("info: Stream File                " + StreamSoundFilenameAndPath)
		log.Println("info: Stream Volume              " + fmt.Sprintf("%v", StreamSoundVolume))
	} else {
		log.Println("info: ------------ Sounds  ------------------ SKIPPED ")
	}

	if PrintTxTimeout {
		log.Println("info: ------------ TX Timeout ------------------ ")
		log.Println("info: Tx Timeout Enabled  " + fmt.Sprintf("%t", TxTimeOutEnabled))
		log.Println("info: Tx Timeout Secs     " + fmt.Sprintf("%v", TxTimeOutSecs))
	} else {
		log.Println("info: ------------ TX Timeout ------------------ SKIPPED ")
	}

	if PrintHTTPAPI {
		log.Println("info: ------------ HTTP API  ----------------- ")
		log.Println("info: API Enabled        " + fmt.Sprintf("%t", APIEnabled))
		log.Println("info: API Listen Port    " + APIListenPort)
		log.Println("info: DisplayMenu        " + fmt.Sprintf("%t", APIDisplayMenu))
		log.Println("info: ChannelUp          " + fmt.Sprintf("%t", APIChannelUp))
		log.Println("info: ChannelDown        " + fmt.Sprintf("%t", APIChannelDown))
		log.Println("info: Mute               " + fmt.Sprintf("%t", APIMute))
		log.Println("info: CurentVolumeLevel  " + fmt.Sprintf("%t", APICurrentVolumeLevel))
		log.Println("info: DigitalVolumeUp    " + fmt.Sprintf("%t", APIDigitalVolumeUp))
		log.Println("info: DigitalVolumeDown  " + fmt.Sprintf("%t", APIDigitalVolumeDown))
		log.Println("info: ListServerChannels " + fmt.Sprintf("%t", APIListServerChannels))
		log.Println("info: StartTransmitting  " + fmt.Sprintf("%t", APIStartTransmitting))
		log.Println("info: StopTransmitting   " + fmt.Sprintf("%t", APIStopTransmitting))
		log.Println("info: ListOnlineUsers    " + fmt.Sprintf("%t", APIListOnlineUsers))
		log.Println("info: PlayStream         " + fmt.Sprintf("%t", APIPlayStream))
		log.Println("info: RequestGpsPosition " + fmt.Sprintf("%t", APIRequestGpsPosition))
		log.Println("info: EmailEnabled       " + fmt.Sprintf("%t", APIEmailEnabled))
		log.Println("info: NextServer         " + fmt.Sprintf("%t", APINextServer))
		log.Println("info: PreviousServer     " + fmt.Sprintf("%t", APIPreviousServer))
		log.Println("info: PanicSimulation    " + fmt.Sprintf("%t", APIPanicSimulation))
		log.Println("info: ScanChannels       " + fmt.Sprintf("%t", APIScanChannels))
		log.Println("info: DisplayVersion     " + fmt.Sprintf("%t", APIDisplayVersion))
		log.Println("info: ClearScreen        " + fmt.Sprintf("%t", APIClearScreen))
		log.Println("info: PingServersEnabled " + fmt.Sprintf("%t", APIPingServersEnabled))
		log.Println("info: TxLoopTest         " + fmt.Sprintf("%t", APIRepeatTxLoopTest))
		log.Println("info: PrintXmlConfig     " + fmt.Sprintf("%t", APIPrintXmlConfig))
	} else {
		log.Println("info: ------------ HTTP API  ----------------- SKIPPED ")
	}

	if PrintTargetboard {
		log.Println("info: ------------ Target Board --------------- ")
		log.Println("info: Target Board " + fmt.Sprintf("%v", TargetBoard))
	} else {
		log.Println("info: ------------ Target Board --------------- SKIPPED ")
	}

	if PrintLeds {
		log.Println("info: ------------ LEDS  ---------------------- ")
		log.Println("info: Led Strip Enabled      " + fmt.Sprintf("%v", LedStripEnabled))
		log.Println("info: Voice Activity Led Pin " + fmt.Sprintf("%v", VoiceActivityLEDPin))
		log.Println("info: Participants Led Pin   " + fmt.Sprintf("%v", ParticipantsLEDPin))
		log.Println("info: Transmit Led Pin       " + fmt.Sprintf("%v", TransmitLEDPin))
		log.Println("info: Online Led Pin         " + fmt.Sprintf("%v", OnlineLEDPin))
		log.Println("info: Attention Led Pin      " + fmt.Sprintf("%v", AttentionLEDPin))
	} else {
		log.Println("info: ------------ LEDS  ---------------------- SKIPPED ")
	}

	if PrintHeartbeat {
		log.Println("info: ---------- HEARTBEAT -------------------- ")
		log.Println("info: HeartBeat Enabled " + fmt.Sprintf("%v", HeartBeatEnabled))
		log.Println("info: HeartBeat LED Pin " + fmt.Sprintf("%v", HeartBeatLEDPin))
		log.Println("info: Period  mSecs     " + fmt.Sprintf("%v", PeriodmSecs))
		log.Println("info: Led On  mSecs     " + fmt.Sprintf("%v", LEDOnmSecs))
		log.Println("info: Led Off mSecs     " + fmt.Sprintf("%v", LEDOffmSecs))
	}

	if PrintButtons {
		log.Println("info: ------------ Buttons  ------------------- ")
		log.Println("info: Tx Button Pin           " + fmt.Sprintf("%v", TxButtonPin))
		log.Println("info: Tx Toggle Pin           " + fmt.Sprintf("%v", TxTogglePin))
		log.Println("info: Channel Up Button Pin   " + fmt.Sprintf("%v", UpButtonPin))
		log.Println("info: Channel Down Button Pin " + fmt.Sprintf("%v", DownButtonPin))
		log.Println("info: Panic Button Pin        " + fmt.Sprintf("%v", PanicButtonPin))
		log.Println("info: Stream Button Pin       " + fmt.Sprintf("%v", StreamButtonPin))
	} else {
		log.Println("info: ------------ Buttons  ------------------- SKIPPED ")
	}

	if PrintComment {
		log.Println("info: ------------ Comment  ------------------- ")
		log.Println("info: Comment Button Pin         " + fmt.Sprintf("%v", CommentButtonPin))
		log.Println("info: Comment Message State 1    " + fmt.Sprintf("%v", CommentMessageOff))
		log.Println("info: Comment Message State 2    " + fmt.Sprintf("%v", CommentMessageOn))
	} else {
		log.Println("info: ------------ Comment  ------------------- SKIPPED ")
	}

	if PrintLcd {
		log.Println("info: ------------ LCD HD44780 ----------------------- ")
		log.Println("info: LCDEnabled               " + fmt.Sprintf("%v", LCDEnabled))
		log.Println("info: LCDInterfaceType         " + fmt.Sprintf("%v", LCDInterfaceType))
		log.Println("info: Lcd I2C Address          " + fmt.Sprintf("%x", LCDI2CAddress))
		log.Println("info: Back Light Timer Enabled " + fmt.Sprintf("%t", LCDBackLightTimerEnabled))
		log.Println("info: Back Light Timer Timeout " + fmt.Sprintf("%v", LCDBackLightTimeout))
		log.Println("info: Back Light Pin " + fmt.Sprintf("%v", LCDBackLightLEDPin))
		log.Println("info: RS Pin " + fmt.Sprintf("%v", LCDRSPin))
		log.Println("info: E  Pin " + fmt.Sprintf("%v", LCDEPin))
		log.Println("info: D4 Pin " + fmt.Sprintf("%v", LCDD4Pin))
		log.Println("info: D5 Pin " + fmt.Sprintf("%v", LCDD5Pin))
		log.Println("info: D6 Pin " + fmt.Sprintf("%v", LCDD6Pin))
		log.Println("info: D7 Pin " + fmt.Sprintf("%v", LCDD7Pin))
	} else {
		log.Println("info: ------------ LCD  ----------------------- SKIPPED ")
	}

	if PrintOled {
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
	} else {
		log.Println("info: ------------ OLED ----------------------- SKIPPED ")
	}

	if PrintGps {
		log.Println("info: ------------ GPS  ------------------------ ")
		log.Println("info: GPS Enabled            " + fmt.Sprintf("%t", GpsEnabled))
		log.Println("info: Port                   ", Port)
		log.Println("info: Baud                   " + fmt.Sprintf("%v", Baud))
		log.Println("info: TxData                 ", TxData)
		log.Println("info: Even                   " + fmt.Sprintf("%v", Even))
		log.Println("info: Odd                    " + fmt.Sprintf("%v", Odd))
		log.Println("info: RS485                  " + fmt.Sprintf("%v", Rs485))
		log.Println("info: RS485 High During Send " + fmt.Sprintf("%v", Rs485HighDuringSend))
		log.Println("info: RS485 High After Send  " + fmt.Sprintf("%v", Rs485HighAfterSend))
		log.Println("info: Stop Bits              " + fmt.Sprintf("%v", StopBits))
		log.Println("info: Data Bits              " + fmt.Sprintf("%v", DataBits))
		log.Println("info: Char Time Out          " + fmt.Sprintf("%v", CharTimeOut))
		log.Println("info: Min Read               " + fmt.Sprintf("%v", MinRead))
		log.Println("info: Rx                     " + fmt.Sprintf("%t", Rx))
	} else {
		log.Println("info: ------------ GPS  ------------------------ SKIPPED ")
	}

	if PrintTraccar {
		log.Println("info: ------------ TRACCAR Info  ----------------------- ")
		log.Println("info: Track Enabled            " + fmt.Sprintf("%t", TrackEnabled))
		log.Println("info: Traccar Send To          " + fmt.Sprintf("%t", TraccarSendTo))
		log.Println("info: Traccar Server URL       ", TraccarServerURL)
		log.Println("info: Traccar Server IP        ", TraccarServerIP)
		log.Println("info: Traccar Client ID        ", TraccarClientId)
		log.Println("info: Traccar Report Frequency " + fmt.Sprintf("%v", TraccarReportFrequency))
		log.Println("info: Traccar Proto            ", TraccarProto)
		log.Println("info: Traccar Server Full URL  ", TraccarServerFullURL)
		log.Println("info: Track GPS Show Lcd       " + fmt.Sprintf("%t", TrackGPSShowLCD))
		log.Println("info: Track Verbose            " + fmt.Sprintf("%t", TrackVerbose))

	} else {
		log.Println("info: ------------ TRACCAR Info ------------------------ SKIPPED ")
	}

	if PrintPanic {
		log.Println("info: ------------ PANIC Function -------------- ")
		log.Println("info: Panic Function Enable          ", fmt.Sprintf("%t", PEnabled))
		log.Println("info: Panic Sound Filename and Path  ", PFilenameAndPath)
		log.Println("info: Panic Message                  ", PMessage)
		log.Println("info: Panic Email Send               ", fmt.Sprintf("%t", PMailEnabled))
		log.Println("info: Panic Message Send Recursively ", fmt.Sprintf("%t", PRecursive))
		log.Println("info: Panic Volume                   ", fmt.Sprintf("%v", PVolume))
		log.Println("info: Panic Send Ident               ", fmt.Sprintf("%t", PSendIdent))
		log.Println("info: Panic Send GPS Location        ", fmt.Sprintf("%t", PSendGpsLocation))
		log.Println("info: Panic TX Lock Enabled          ", fmt.Sprintf("%t", PTxLockEnabled))
		log.Println("info: Panic TX Lock Timeout Secs     ", fmt.Sprintf("%v", PTxlockTimeOutSecs))
		log.Println("info: Panic Low Profile Lights Enable", fmt.Sprintf("%v", PLowProfile))
	} else {
		log.Println("info: ------------ PANIC Function -------------- SKIPPED ")
	}

	if PrintAudioRecord {
		log.Println("info: ------------ AUDIO RECORDING Function -------------- ")
		log.Println("info: Audio Recording Enabled " + fmt.Sprintf("%v", AudioRecordEnabled))
		log.Println("info: Audio Recording On Start " + fmt.Sprintf("%v", AudioRecordOnStart))
		log.Println("info: Audio Recording System " + fmt.Sprintf("%v", AudioRecordSystem))
		log.Println("info: Audio Record Mode " + fmt.Sprintf("%v", AudioRecordMode))
		log.Println("info: Audio Record Timeout " + fmt.Sprintf("%v", AudioRecordTimeout))
		log.Println("info: Audio Record From Output " + fmt.Sprintf("%v", AudioRecordFromOutput))
		log.Println("info: Audio Record From Input " + fmt.Sprintf("%v", AudioRecordFromInput))
		log.Println("info: Audio Recording Mic Timeout " + fmt.Sprintf("%v", AudioRecordMicTimeout))
		log.Println("info: Audio Recording Save Path " + fmt.Sprintf("%v", AudioRecordSavePath))
		log.Println("info: Audio Recording Archive Path " + fmt.Sprintf("%v", AudioRecordArchivePath))
		log.Println("info: Audio Recording Soft " + fmt.Sprintf("%v", AudioRecordSoft))
		log.Println("info: Audio Recording Profile " + fmt.Sprintf("%v", AudioRecordProfile))
		log.Println("info: Audio Recording File Format " + fmt.Sprintf("%v", AudioRecordFileFormat))
		log.Println("info: Audio Recording Chunk Size " + fmt.Sprintf("%v", AudioRecordChunkSize))
	} else {
		log.Println("info: ------------ AUDIO RECORDING Function ------- SKIPPED ")
	}
	if PrintMQTT {
		log.Println("info: ------------ MQTT Function -------------- ")
		log.Println("info: Enabled   " + fmt.Sprintf("%v", MQTTEnabled))
		log.Println("info: Topic     " + fmt.Sprintf("%v", MQTTTopic))
		log.Println("info: Broker    " + fmt.Sprintf("%v", MQTTBroker))
		log.Println("info: Password  " + fmt.Sprintf("%v", MQTTPassword))
		log.Println("info: Id        " + fmt.Sprintf("%v", MQTTId))
		log.Println("info: Cleansess " + fmt.Sprintf("%v", MQTTCleansess))
		log.Println("info: Qos       " + fmt.Sprintf("%v", MQTTQos))
		log.Println("info: Num       " + fmt.Sprintf("%v", MQTTNum))
		log.Println("info: Payload   " + fmt.Sprintf("%v", MQTTPayload))
		log.Println("info: Action    " + fmt.Sprintf("%v", MQTTAction))
		log.Println("info: Store     " + fmt.Sprintf("%v", MQTTStore))
	} else {
		log.Println("info: ------------ MQTT Function ------- SKIPPED ")
	}

	if PrintKeyboardMap {
		log.Println("info: ------------ KeyboardMap Function -------------- ")
		log.Printf("TTYKeymap %+v\n", TTYKeyMap)
		log.Printf("USBKeymap %+v\n", USBKeyMap)
	} else {
		log.Println("info: ------------ KeyboardMap Function ------ SKIPPED ")
	}

	if PrintUSBKeyboard {
		log.Println("info: ------------ USBKeyboard Function -------------- ")
		log.Println("USBKeyboardEnabled", USBKeyboardEnabled)
		log.Println("USBKeyboardPath", USBKeyboardPath)
		log.Println("NumLockScanID", NumlockScanID)
	} else {
		log.Println("info: ------------ USBKeyboard Function ------ SKIPPED ")
	}

}

func modifyXMLTagServerHopping(inputXMLFile string, outputXMLFile string, nextserverindex int) {
	xmlfilein, err := os.Open(inputXMLFile)

	if err != nil {
		FatalCleanUp(err.Error())
	}

	xmlfileout, err := os.Create(outputXMLFile)

	if err != nil {
		FatalCleanUp(err.Error())
	}

	defer xmlfilein.Close()
	defer xmlfileout.Close()
	decoder := xml.NewDecoder(xmlfilein)
	encoder := xml.NewEncoder(xmlfileout)
	encoder.Indent("", "	")

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("error: Getting token: %v\n", err)
			break
		}

		switch v := token.(type) {
		case xml.StartElement:
			if v.Name.Local == "Document" {
				var document DocumentStruct
				if v.Name.Local != "talkkonnect/xml" {
					err = decoder.DecodeElement(&document, &v)
					if err != nil {
						FatalCleanUp("Cannot Find XML Tag Document" + err.Error())
					}
				}
				// XML Tag to Replace
				document.Global.Software.Settings.NextServerIndex = nextserverindex

				err = encoder.EncodeElement(document, v)
				if err != nil {
					FatalCleanUp(err.Error())
				}
				continue
			}

		}

		if err := encoder.EncodeToken(xml.CopyToken(token)); err != nil {
			FatalCleanUp(err.Error())
		}
	}

	if err := encoder.Flush(); err != nil {
		FatalCleanUp(err.Error())
	} else {
		time.Sleep(2 * time.Second)
		copyFile(inputXMLFile, inputXMLFile+".bak")
		deleteFile(inputXMLFile)
		copyFile(outputXMLFile, inputXMLFile)
		c := exec.Command("reset")
		c.Stdout = os.Stdout
		c.Run()
		os.Exit(0)
	}

}
