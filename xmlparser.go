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
	"errors"
	"fmt"
	goled "github.com/talkkonnect/go-oled-i2c"
	"github.com/talkkonnect/go-openal/openal"
	"github.com/talkkonnect/gumble/gumbleffmpeg"
	"golang.org/x/sys/unix"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//version and release date
const (
	talkkonnectVersion  string = "1.52.02"
	talkkonnectReleased string = "December 04 2020"
)

var (
	pstream       *gumbleffmpeg.Stream
	AccountCount  int  = 0
	KillHeartBeat bool = false
	IsPlayStream  bool = false
)

// Generic Global Variables
var (
	BackLightTime              = time.NewTicker(5 * time.Second)
	BackLightTimePtr           = &BackLightTime
	ConnectAttempts            = 0
	IsConnected           bool = false
	source                     = openal.NewSource()
	StartTime                  = time.Now()
	BufferToOpenALCounter      = 0
)

//account settings
var (
	Default     []bool
	Name        []string
	Server      []string
	Username    []string
	Password    []string
	Insecure    []bool
	Certificate []string
	Channel     []string
	Ident       []string
)

//software settings
var (
	OutputDevice       string = "PCM"
	LogFilenameAndPath string
	Logging            string = "screen"
	Loglevel           string = "info"
	Daemonize          bool
	SimplexWithMute    bool = true
	TxCounter          bool
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
	TTSPlayChimes                        bool
	TTSPlayChimesFilenameAndPath         string
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
	EventSoundFilenameAndPath         string
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
	RepeaterToneFilenameAndPath       string
	RepeaterToneVolume                float32
	ChimesSoundEnabled                bool
	ChimesSoundFilenameAndPath        string
	ChimesSoundVolume                 float32
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
	APIPlayChimes         bool
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
	PrintPanic        bool
	PrintAudioRecord  bool
)

// target board settings
var (
	TargetBoard string
)

//indicator light settings
var (
	VoiceActivityLEDPin uint
	ParticipantsLEDPin  uint
	TransmitLEDPin      uint
	OnlineLEDPin        uint
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
	ChimesButtonPin uint
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
	LCDBackLightTimeoutSecs  time.Duration
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
)

//panic function settings
var (
	PEnabled           bool
	PFilenameAndPath   string
	PMessage           string
	PRecursive         bool
	PVolume            float32
	PSendIdent         bool
	PSendGpsLocation   bool
	PTxLockEnabled     bool
	PTxlockTimeOutSecs uint
)

//audio recording settings // New
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

//other global variables used for state tracking
var (
	txcounter         int
	togglecounter     int
	isTx              bool
	isPlayStream      bool
	CancellableStream bool
)

type Document struct {
	XMLName  xml.Name `xml:"document"`
	Text     string   `xml:",chardata"`
	Type     string   `xml:"type,attr"`
	Accounts struct {
		Text    string `xml:",chardata"`
		Account []struct {
			Text          string `xml:",chardata"`
			Name          string `xml:"name,attr"`
			Default       bool   `xml:"default,attr"`
			ServerAndPort string `xml:"serverandport"`
			UserName      string `xml:"username"`
			Password      string `xml:"password"`
			Insecure      bool   `xml:"insecure"`
			Certificate   string `xml:"certificate"`
			Channel       string `xml:"channel"`
			Ident         string `xml:"ident"`
		} `xml:"account"`
	} `xml:"accounts"`
	Global struct {
		Text     string `xml:",chardata"`
		Software struct {
			Text     string `xml:",chardata"`
			Settings struct {
				Text               string `xml:",chardata"`
				OutputDevice       string `xml:"outputdevice"`
				LogFilenameAndPath string `xml:"logfilenameandpath"`
				Logging            string `xml:"logging"`
				Loglevel           string `xml:"loglevel"`
				Daemonize          bool   `xml:"daemonize"`
				CancellableStream  bool   `xml:"cancellablestream"`
				SimplexWithMute    bool   `xml:"simplexwithmute"`
				TxCounter          bool   `xml:"txcounter"`
			} `xml:"settings"`
			AutoProvisioning struct {
				Text         string `xml:",chardata"`
				Enabled      bool   `xml:"enabled,attr"`
				TkID         string `xml:"tkid"`
				URL          string `xml:"url"`
				SaveFilePath string `xml:"savefilepath"`
				SaveFilename string `xml:"savefilename"`
			} `xml:"autoprovisioning"`
			Beacon struct {
				Text              string  `xml:",chardata"`
				Enabled           bool    `xml:"enabled,attr"`
				BeaconTimerSecs   int     `xml:"beacontimersecs"`
				BeaconFileAndPath string  `xml:"beaconfileandpath"`
				Volume            float32 `xml:"volume"`
			} `xml:"beacon"`
			TTS struct {
				Text                              string `xml:",chardata"`
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
				PlayChimes                        bool   `xml:"playchimes"`
				PlayChimesFilenameAndPath         string `xml:"playchimesfilenameandpath"`
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
				Text          string `xml:",chardata"`
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
				Text  string `xml:",chardata"`
				Event struct {
					Text            string `xml:",chardata"`
					Enabled         bool   `xml:"enabled,attr"`
					FilenameAndPath string `xml:"filenameandpath"`
				} `xml:"event"`
				Alert struct {
					Text            string  `xml:",chardata"`
					Enabled         bool    `xml:"enabled,attr"`
					FilenameAndPath string  `xml:"filenameandpath"`
					Volume          float32 `xml:"volume"`
				} `xml:"alert"`
				IncommingBeep struct {
					Text            string  `xml:",chardata"`
					Enabled         bool    `xml:"enabled,attr"`
					FilenameAndPath string  `xml:"filenameandpath"`
					Volume          float32 `xml:"volume"`
				} `xml:"incommingbeep"`
				RogerBeep struct {
					Text            string  `xml:",chardata"`
					Enabled         bool    `xml:"enabled,attr"`
					FilenameAndPath string  `xml:"filenameandpath"`
					Volume          float32 `xml:"volume"`
				} `xml:"rogerbeep"`
				RepeaterTone struct {
					Text            string  `xml:",chardata"`
					Enabled         bool    `xml:"enabled,attr"`
					FilenameAndPath string  `xml:"filenameandpath"`
					Volume          float32 `xml:"volume"`
				} `xml:"repeatertone"`
				Chimes struct {
					Text            string  `xml:",chardata"`
					Enabled         bool    `xml:"enabled,attr"`
					FilenameAndPath string  `xml:"filenameandpath"`
					Volume          float32 `xml:"volume"`
				} `xml:"chimes"`
			} `xml:"sounds"`
			TxTimeOut struct {
				Text          string `xml:",chardata"`
				Enabled       bool   `xml:"enabled,attr"`
				TxTimeOutSecs int    `xml:"txtimeoutsecs"`
			} `xml:"txtimeout"`
			API struct {
				Text               string `xml:",chardata"`
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
				PlayChimes         bool   `xml:"playchimes"`
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
				Text              string `xml:",chardata"`
				PrintAccount      bool   `xml:"printaccount"`
				PrintLogging      bool   `xml:"printlogging"`
				PrintProvisioning bool   `xml:"printprovisioning"`
				PrintBeacon       bool   `xml:"printbeacon"`
				PrintTTS          bool   `xml:"printtts"`
				PrintSMTP         bool   `xml:"printsmtp"`
				PrintSounds       bool   `xml:"printsounds"`
				PrintTxTimeout    bool   `xml:"printtxtimeout"`
				PrintHTTPAPI      bool   `xml:"printhttpapi"`
				PrintTargetBoard  bool   `xml:"printtargetboard"`
				PrintLeds         bool   `xml:"printleds"`
				PrintHeartbeat    bool   `xml:"printheartbeat"`
				PrintButtons      bool   `xml:"printbuttons"`
				PrintComment      bool   `xml:"printcomment"`
				PrintLcd          bool   `xml:"printlcd"`
				PrintOled         bool   `xml:"printoled"`
				PrintGps          bool   `xml:"printgps"`
				PrintPanic        bool   `xml:"printpanic"`
				PrintAudioRecord  bool   `xml:"printaudiorecord"`
			} `xml:"printvariables"`
		} `xml:"software"`
		Hardware struct {
			Text        string `xml:",chardata"`
			TargetBoard string `xml:"targetboard,attr"`
			Lights      struct {
				Text                string `xml:",chardata"`
				VoiceActivityLedPin string `xml:"voiceactivityledpin"`
				ParticipantsLedPin  string `xml:"participantsledpin"`
				TransmitLedPin      string `xml:"transmitledpin"`
				OnlineLedPin        string `xml:"onlineledpin"`
			} `xml:"lights"`
			HeartBeat struct {
				Text        string `xml:",chardata"`
				Enabled     bool   `xml:"enabled,attr"`
				LEDPin      string `xml:"heartbeatledpin"`
				Periodmsecs int    `xml:"periodmsecs"`
				LEDOnmsecs  int    `xml:"ledonmsecs"`
				LEDOffmsecs int    `xml:"ledoffmsecs"`
			} `xml:"heartbeat"`
			Buttons struct {
				Text            string `xml:",chardata"`
				TxButtonPin     string `xml:"txbuttonpin"`
				TxTogglePin     string `xml:"txtogglepin"`
				UpButtonPin     string `xml:"upbuttonpin"`
				DownButtonPin   string `xml:"downbuttonpin"`
				PanicButtonPin  string `xml:"panicbuttonpin"`
				ChimesButtonPin string `xml:"chimesbuttonpin"`
			} `xml:"buttons"`
			Comment struct {
				Text              string `xml:",chardata"`
				CommentButtonPin  string `xml:"commentbuttonpin"`
				CommentMessageOff string `xml:"commentmessageoff"`
				CommentMessageOn  string `xml:"commentmessageon"`
			} `xml:"comment"`
			LCD struct {
				Text                  string `xml:",chardata"`
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
				Text                    string `xml:",chardata"`
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
				Text                string `xml:",chardata"`
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
			} `xml:"gps"`
			PanicFunction struct {
				Text                 string  `xml:",chardata"`
				Enabled              bool    `xml:"enabled,attr"`
				FilenameAndPath      string  `xml:"filenameandpath"`
				Volume               float32 `xml:"volume"`
				SendIdent            bool    `xml:"sendident"`
				Message              string  `xml:"panicmessage"`
				RecursiveSendMessage string  `xml:"recursivesendmessage"`
				SendGpsLocation      bool    `xml:"sendgpslocation"`
				TxLockEnabled        bool    `xml:"txlockenabled"`
				TxLockTimeOutSecs    uint    `xml:"txlocktimeoutsecs"`
			} `xml:"panicfunction"`
			AudioRecordFunction struct {
				Text              string `xml:",chardata"`
				Enabled           bool   `xml:"enabled,attr"`
				RecordOnStart     bool   `xml:"recordonstart"`
				RecordSystem      string `xml:"recordsystem"` // New
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
		} `xml:"hardware"`
	} `xml:"global"`
}

func readxmlconfig(file string) error {
	xmlFile, err := os.Open(file)
	if err != nil {
		return errors.New(fmt.Sprintf("error: cannot open configuration file "+filepath.Base(file), err))
	}
	log.Println("info: Successfully Opened file " + filepath.Base(file))
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var document Document

	err = xml.Unmarshal(byteValue, &document)
	if err != nil {
		errors.New(fmt.Sprintf("error: File "+filepath.Base(file)+" formatting error Please fix! ", err))
	}
	log.Println("info: Document               : " + document.Type)

	for i := 0; i < len(document.Accounts.Account); i++ {
		if document.Accounts.Account[i].Default == true {
			Name = append(Name, document.Accounts.Account[i].Name)
			Server = append(Server, document.Accounts.Account[i].ServerAndPort)
			Username = append(Username, document.Accounts.Account[i].UserName)
			Password = append(Password, document.Accounts.Account[i].Password)
			Insecure = append(Insecure, document.Accounts.Account[i].Insecure)
			Certificate = append(Certificate, document.Accounts.Account[i].Certificate)
			Channel = append(Channel, document.Accounts.Account[i].Channel)
			Ident = append(Ident, document.Accounts.Account[i].Ident)
			AccountCount++
		}
	}

	if AccountCount == 0 {
		log.Fatal("No Default Accounts Found! Please Add at least 1 Default Account in XML File")
	}

	exec, err := os.Executable()
	if err != nil {
		exec = "./talkkonnect" //Hardcode our default name
	}

	// Set our default config file path (for autoprovision)
	defaultConfPath, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		log.Fatal("error: Unable to get path for config file: " + err.Error())
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

	OutputDevice = document.Global.Software.Settings.OutputDevice
	LogFilenameAndPath = document.Global.Software.Settings.LogFilenameAndPath
	Logging = document.Global.Software.Settings.Logging

	if document.Global.Software.Settings.Loglevel == "trace" || document.Global.Software.Settings.Loglevel == "debug" || document.Global.Software.Settings.Loglevel == "info" || document.Global.Software.Settings.Loglevel == "warning" || document.Global.Software.Settings.Loglevel == "error" || document.Global.Software.Settings.Loglevel == "alert" {
		Loglevel = document.Global.Software.Settings.Loglevel
	}

	if strings.ToLower(Logging) != "screen" && LogFilenameAndPath == "" {
		LogFilenameAndPath = defaultLogPath
	}

	Daemonize = document.Global.Software.Settings.Daemonize
	CancellableStream = document.Global.Software.Settings.CancellableStream
	SimplexWithMute = document.Global.Software.Settings.SimplexWithMute
	TxCounter = document.Global.Software.Settings.TxCounter

	APEnabled = document.Global.Software.AutoProvisioning.Enabled
	TkID = document.Global.Software.AutoProvisioning.TkID
	URL = document.Global.Software.AutoProvisioning.URL
	SaveFilePath = document.Global.Software.AutoProvisioning.SaveFilePath
	SaveFilename = document.Global.Software.AutoProvisioning.SaveFilename

	if APEnabled && SaveFilePath == "" {
		SaveFilePath = defaultConfPath
	}

	if APEnabled && SaveFilename == "" {
		SaveFilename = filepath.Base(exec) + ".xml" //Should default to talkkonnect.xml
	}

	BeaconEnabled = document.Global.Software.Beacon.Enabled
	BeaconTimerSecs = document.Global.Software.Beacon.BeaconTimerSecs
	BeaconFilenameAndPath = document.Global.Software.Beacon.BeaconFileAndPath
	if BeaconEnabled && BeaconFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/Beacon.wav"
		if _, err := os.Stat(path); err == nil {
			BeaconFilenameAndPath = path
		}
	}

	BVolume = document.Global.Software.Beacon.Volume

	TTSEnabled = document.Global.Software.TTS.Enabled
	TTSVolumeLevel = document.Global.Software.TTS.VolumeLevel
	TTSParticipants = document.Global.Software.TTS.Participants
	TTSChannelUp = document.Global.Software.TTS.ChannelUp
	TTSChannelUpFilenameAndPath = document.Global.Software.TTS.ChannelUpFilenameAndPath

	if TTSChannelUp && TTSChannelUpFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/ChannelUp.wav"
		if _, err := os.Stat(path); err == nil {
			TTSChannelUpFilenameAndPath = path
		}
	}

	TTSChannelUpFilenameAndPath = document.Global.Software.TTS.ChannelUpFilenameAndPath
	TTSChannelDown = document.Global.Software.TTS.ChannelDown
	TTSChannelDownFilenameAndPath = document.Global.Software.TTS.ChannelDownFilenameAndPath

	if TTSChannelDown && TTSChannelDownFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/ChannelDown.wav"
		if _, err := os.Stat(path); err == nil {
			TTSChannelDownFilenameAndPath = path
		}
	}

	TTSMuteUnMuteSpeaker = document.Global.Software.TTS.MuteUnmuteSpeaker
	TTSMuteUnMuteSpeakerFilenameAndPath = document.Global.Software.TTS.MuteUnmuteSpeakerFilenameAndPath

	if TTSMuteUnMuteSpeaker && TTSMuteUnMuteSpeakerFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/MuteUnMuteSpeaker.wav"
		if _, err := os.Stat(path); err == nil {
			TTSMuteUnMuteSpeakerFilenameAndPath = path
		}
	}

	TTSCurrentVolumeLevel = document.Global.Software.TTS.CurrentVolumeLevel
	TTSCurrentVolumeLevelFilenameAndPath = document.Global.Software.TTS.CurrentVolumeLevelFilenameAndPath

	if TTSCurrentVolumeLevel && TTSCurrentVolumeLevelFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/CurrentVolumeLevel.wav"
		if _, err := os.Stat(path); err == nil {
			TTSCurrentVolumeLevelFilenameAndPath = path
		}
	}

	TTSDigitalVolumeUp = document.Global.Software.TTS.DigitalVolumeUp
	TTSDigitalVolumeUpFilenameAndPath = document.Global.Software.TTS.DigitalVolumeUpFilenameAndPath

	if TTSDigitalVolumeUp && TTSDigitalVolumeUpFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/DigitalVolumeUp.wav"
		if _, err := os.Stat(path); err == nil {
			TTSDigitalVolumeUpFilenameAndPath = path
		}
	}

	TTSDigitalVolumeDown = document.Global.Software.TTS.DigitalVolumeDown
	TTSDigitalVolumeDownFilenameAndPath = document.Global.Software.TTS.DigitalVolumeDownFilenameAndPath

	if TTSDigitalVolumeDown && TTSDigitalVolumeDownFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/DigitalVolumeDown.wav"
		if _, err := os.Stat(path); err == nil {
			TTSDigitalVolumeDownFilenameAndPath = path
		}
	}

	TTSListServerChannels = document.Global.Software.TTS.ListServerChannels
	TTSListServerChannelsFilenameAndPath = document.Global.Software.TTS.ListServerChannelsFilenameAndPath

	if TTSListServerChannels && TTSListServerChannelsFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/ListServerChannels.wav"
		if _, err := os.Stat(path); err == nil {
			TTSListServerChannelsFilenameAndPath = path
		}
	}

	TTSStartTransmitting = document.Global.Software.TTS.StartTransmitting
	TTSStartTransmittingFilenameAndPath = document.Global.Software.TTS.StartTransmittingFilenameAndPath

	if TTSStartTransmitting && TTSStartTransmittingFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/StartTransmitting.wav"
		if _, err := os.Stat(path); err == nil {
			TTSStartTransmittingFilenameAndPath = path
		}
	}

	TTSStopTransmitting = document.Global.Software.TTS.StopTransmitting
	TTSStopTransmittingFilenameAndPath = document.Global.Software.TTS.StopTransmittingFilenameAndPath

	if TTSStopTransmitting && TTSStopTransmittingFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/StopTransmitting.wav"
		if _, err := os.Stat(path); err == nil {
			TTSStopTransmittingFilenameAndPath = path
		}
	}

	TTSListOnlineUsers = document.Global.Software.TTS.ListOnlineUsers
	TTSListOnlineUsersFilenameAndPath = document.Global.Software.TTS.ListOnlineUsersFilenameAndPath

	if TTSListOnlineUsers && TTSListOnlineUsersFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/ListOnlineUsers.wav"
		if _, err := os.Stat(path); err == nil {
			TTSListOnlineUsersFilenameAndPath = path
		}
	}

	TTSPlayChimes = document.Global.Software.TTS.PlayChimes
	TTSPlayChimesFilenameAndPath = document.Global.Software.TTS.PlayChimesFilenameAndPath

	if TTSPlayChimes && TTSPlayChimesFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/PlayChimes.wav"
		if _, err := os.Stat(path); err == nil {
			TTSPlayChimesFilenameAndPath = path
		}
	}

	TTSRequestGpsPosition = document.Global.Software.TTS.RequestGpsPosition
	TTSRequestGpsPositionFilenameAndPath = document.Global.Software.TTS.RequestGpsPositionFilenameAndPath

	if TTSRequestGpsPosition && TTSRequestGpsPositionFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/RequestGpsPosition.wav"
		if _, err := os.Stat(path); err == nil {
			TTSRequestGpsPositionFilenameAndPath = path
		}
	}

	TTSNextServer = document.Global.Software.TTS.NextServer
	TTSNextServerFilenameAndPath = document.Global.Software.TTS.NextServerFilenameAndPath
	/*
		//TODO: No default sound available. Placeholder for now
		if TTSNextServer && TTSNextServerFilenameAndPath == "" {
			path := defaultSharePath + "/soundfiles/voiceprompts/TODO"
			if _, err := os.Stat(path); err == nil {
				TTSNextServerFilenameAndPath = path
			}
		}
	*/

	TTSPreviousServer = document.Global.Software.TTS.PreviousServer
	TTSPreviousServerFilenameAndPath = document.Global.Software.TTS.PreviousServerFilenameAndPath

	/*
		//TODO: No default sound available. Placeholder for now
		if TTSPreviousServer && TTSPreviousServerFilenameAndPath == "" {
			path := defaultSharePath + "/soundfiles/voiceprompts/TODO"
			if _, err := os.Stat(path); err == nil {
				TTSPreviousServerFilenameAndPath = path
			}
		}
	*/

	TTSPanicSimulation = document.Global.Software.TTS.PanicSimulation
	TTSPanicSimulationFilenameAndPath = document.Global.Software.TTS.PanicSimulationFilenameAndPath
	if TTSPanicSimulation && TTSPanicSimulationFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/PanicSimulation.wav"
		if _, err := os.Stat(path); err == nil {
			TTSPanicSimulationFilenameAndPath = path
		}
	}

	TTSPrintXmlConfig = document.Global.Software.TTS.PrintXmlConfig
	TTSPrintXmlConfigFilenameAndPath = document.Global.Software.TTS.PrintXmlConfigFilenameAndPath

	if TTSPrintXmlConfig && TTSPrintXmlConfigFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/PrintXmlConfig.wav"
		if _, err := os.Stat(path); err == nil {
			TTSPrintXmlConfigFilenameAndPath = path
		}
	}

	TTSSendEmail = document.Global.Software.TTS.SendEmail
	TTSSendEmailFilenameAndPath = document.Global.Software.TTS.SendEmailFilenameAndPath

	if TTSSendEmail && TTSSendEmailFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/SendEmail.wav"
		if _, err := os.Stat(path); err == nil {
			TTSSendEmailFilenameAndPath = path
		}
	}

	TTSDisplayMenu = document.Global.Software.TTS.DisplayMenu
	TTSDisplayMenuFilenameAndPath = document.Global.Software.TTS.DisplayMenuFilenameAndPath

	if TTSDisplayMenu && TTSDisplayMenuFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/DisplayMenu.wav"
		if _, err := os.Stat(path); err == nil {
			TTSDisplayMenuFilenameAndPath = path
		}
	}

	TTSQuitTalkkonnect = document.Global.Software.TTS.QuitTalkkonnect
	TTSQuitTalkkonnectFilenameAndPath = document.Global.Software.TTS.QuitTalkkonnectFilenameAndPath

	if TTSQuitTalkkonnect && TTSQuitTalkkonnectFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/QuitTalkkonnect.wav"
		if _, err := os.Stat(path); err == nil {
			TTSQuitTalkkonnectFilenameAndPath = path
		}
	}

	TTSTalkkonnectLoaded = document.Global.Software.TTS.TalkkonnectLoaded
	TTSTalkkonnectLoadedFilenameAndPath = document.Global.Software.TTS.TalkkonnectLoadedFilenameAndPath

	if TTSTalkkonnectLoaded && TTSTalkkonnectLoadedFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/voiceprompts/Loaded.wav"
		if _, err := os.Stat(path); err == nil {
			TTSTalkkonnectLoadedFilenameAndPath = path
		}
	}

	TTSPingServers = document.Global.Software.TTS.PingServers
	TTSPingServersFilenameAndPath = document.Global.Software.TTS.PingServersFilenameAndPath

	/*
		//TODO: No default sound available. Placeholder for now
		if TTSPingServers && TTSPingServersFilenameAndPath == "" {
			path := defaultSharePath + "/soundfiles/voiceprompts/TODO"
			if _, err := os.Stat(path); err == nil {
				TTSPingServersFilenameAndPath = path
			}
		}
	*/

	EmailEnabled = document.Global.Software.SMTP.Enabled
	EmailUsername = document.Global.Software.SMTP.Username
	EmailPassword = document.Global.Software.SMTP.Password
	EmailReceiver = document.Global.Software.SMTP.Receiver
	EmailSubject = document.Global.Software.SMTP.Subject
	EmailMessage = document.Global.Software.SMTP.Message
	EmailGpsDateTime = document.Global.Software.SMTP.GpsDateTime
	EmailGpsLatLong = document.Global.Software.SMTP.GpsLatLong
	EmailGoogleMapsURL = document.Global.Software.SMTP.GoogleMapsURL

	EventSoundEnabled = document.Global.Software.Sounds.Event.Enabled
	EventSoundFilenameAndPath = document.Global.Software.Sounds.Event.FilenameAndPath

	if EventSoundEnabled && EventSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/events/event.wav"
		if _, err := os.Stat(path); err == nil {
			EventSoundFilenameAndPath = path
		}
	}
	AlertSoundEnabled = document.Global.Software.Sounds.Alert.Enabled
	AlertSoundFilenameAndPath = document.Global.Software.Sounds.Alert.FilenameAndPath

	if AlertSoundEnabled && AlertSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/alerts/alert.wav"
		if _, err := os.Stat(path); err == nil {
			AlertSoundFilenameAndPath = path
		}
	}

	AlertSoundVolume = document.Global.Software.Sounds.Alert.Volume

	IncommingBeepSoundEnabled = document.Global.Software.Sounds.IncommingBeep.Enabled
	IncommingBeepSoundFilenameAndPath = document.Global.Software.Sounds.IncommingBeep.FilenameAndPath

	if IncommingBeepSoundEnabled && IncommingBeepSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/rogerbeeps/Chirsp.wav"
		if _, err := os.Stat(path); err == nil {
			IncommingBeepSoundFilenameAndPath = path
		}
	}

	IncommingBeepSoundVolume = document.Global.Software.Sounds.IncommingBeep.Volume

	RogerBeepSoundEnabled = document.Global.Software.Sounds.RogerBeep.Enabled
	RogerBeepSoundFilenameAndPath = document.Global.Software.Sounds.RogerBeep.FilenameAndPath

	if RogerBeepSoundEnabled && RogerBeepSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/rogerbeeps/Chirsp.wav"
		if _, err := os.Stat(path); err == nil {
			RogerBeepSoundFilenameAndPath = path
		}
	}

	RogerBeepSoundVolume = document.Global.Software.Sounds.RogerBeep.Volume

	RepeaterToneEnabled = document.Global.Software.Sounds.RepeaterTone.Enabled
	RepeaterToneFilenameAndPath = document.Global.Software.Sounds.RepeaterTone.FilenameAndPath

	if RepeaterToneEnabled && RepeaterToneFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/repeatertones/Chirsp.wav"
		if _, err := os.Stat(path); err == nil {
			RepeaterToneFilenameAndPath = path
		}
	}

	RepeaterToneVolume = document.Global.Software.Sounds.RepeaterTone.Volume

	ChimesSoundEnabled = document.Global.Software.Sounds.Chimes.Enabled
	ChimesSoundFilenameAndPath = document.Global.Software.Sounds.Chimes.FilenameAndPath

	if ChimesSoundEnabled && ChimesSoundFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/alerts/chimes.wav"
		if _, err := os.Stat(path); err == nil {
			ChimesSoundFilenameAndPath = path
		}
	}

	ChimesSoundVolume = document.Global.Software.Sounds.Chimes.Volume

	TxTimeOutEnabled = document.Global.Software.TxTimeOut.Enabled
	TxTimeOutSecs = document.Global.Software.TxTimeOut.TxTimeOutSecs

	APIEnabled = document.Global.Software.API.Enabled
	APIListenPort = document.Global.Software.API.ListenPort
	APIDisplayMenu = document.Global.Software.API.DisplayMenu
	APIChannelUp = document.Global.Software.API.ChannelUp
	APIChannelDown = document.Global.Software.API.ChannelDown
	APIMute = document.Global.Software.API.Mute
	APICurrentVolumeLevel = document.Global.Software.API.CurrentVolumeLevel
	APIDigitalVolumeUp = document.Global.Software.API.DigitalVolumeUp
	APIDigitalVolumeDown = document.Global.Software.API.DigitalVolumeDown
	APIListServerChannels = document.Global.Software.API.ListServerChannels
	APIStartTransmitting = document.Global.Software.API.StartTransmitting
	APIStopTransmitting = document.Global.Software.API.StopTransmitting
	APIListOnlineUsers = document.Global.Software.API.ListOnlineUsers
	APIPlayChimes = document.Global.Software.API.PlayChimes
	APIRequestGpsPosition = document.Global.Software.API.RequestGpsPosition
	APIEmailEnabled = document.Global.Software.API.Enabled
	APINextServer = document.Global.Software.API.NextServer
	APIPreviousServer = document.Global.Software.API.PreviousServer
	APIPanicSimulation = document.Global.Software.API.PanicSimulation
	APIDisplayVersion = document.Global.Software.API.DisplayVersion
	APIClearScreen = document.Global.Software.API.ClearScreen
	APIPingServersEnabled = document.Global.Software.API.Enabled
	APIRepeatTxLoopTest = document.Global.Software.API.RepeatTxLoopTest
	APIPrintXmlConfig = document.Global.Software.API.PrintXmlConfig

	PrintAccount = document.Global.Software.PrintVariables.PrintAccount
	PrintLogging = document.Global.Software.PrintVariables.PrintLogging
	PrintProvisioning = document.Global.Software.PrintVariables.PrintProvisioning
	PrintBeacon = document.Global.Software.PrintVariables.PrintBeacon
	PrintTTS = document.Global.Software.PrintVariables.PrintTTS
	PrintSMTP = document.Global.Software.PrintVariables.PrintSMTP
	PrintSounds = document.Global.Software.PrintVariables.PrintSounds
	PrintTxTimeout = document.Global.Software.PrintVariables.PrintTxTimeout

	PrintHTTPAPI = document.Global.Software.PrintVariables.PrintHTTPAPI

	PrintTargetboard = document.Global.Software.PrintVariables.PrintTargetBoard
	PrintLeds = document.Global.Software.PrintVariables.PrintLeds
	PrintHeartbeat = document.Global.Software.PrintVariables.PrintHeartbeat
	PrintButtons = document.Global.Software.PrintVariables.PrintButtons
	PrintComment = document.Global.Software.PrintVariables.PrintComment
	PrintLcd = document.Global.Software.PrintVariables.PrintLcd
	PrintOled = document.Global.Software.PrintVariables.PrintOled
	PrintGps = document.Global.Software.PrintVariables.PrintGps
	PrintPanic = document.Global.Software.PrintVariables.PrintPanic
	PrintAudioRecord = document.Global.Software.PrintVariables.PrintAudioRecord

	TargetBoard = document.Global.Hardware.TargetBoard

	// my stupid work arround for null uint xml unmarshelling problem with numbers so use strings and convert it 2 times
	temp0, _ := strconv.ParseUint(document.Global.Hardware.Lights.VoiceActivityLedPin, 10, 64)
	VoiceActivityLEDPin = uint(temp0)
	temp1, _ := strconv.ParseUint(document.Global.Hardware.Lights.VoiceActivityLedPin, 10, 64)
	VoiceActivityLEDPin = uint(temp1)
	temp2, _ := strconv.ParseUint(document.Global.Hardware.Lights.ParticipantsLedPin, 10, 64)
	ParticipantsLEDPin = uint(temp2)
	temp3, _ := strconv.ParseUint(document.Global.Hardware.Lights.TransmitLedPin, 10, 64)
	TransmitLEDPin = uint(temp3)
	temp4, _ := strconv.ParseUint(document.Global.Hardware.Lights.OnlineLedPin, 10, 64)
	OnlineLEDPin = uint(temp4)
	temp5, _ := strconv.ParseUint(document.Global.Hardware.HeartBeat.LEDPin, 10, 64)

	HeartBeatLEDPin = uint(temp5)
	HeartBeatEnabled = document.Global.Hardware.HeartBeat.Enabled
	PeriodmSecs = document.Global.Hardware.HeartBeat.Periodmsecs
	LEDOnmSecs = document.Global.Hardware.HeartBeat.LEDOnmsecs
	LEDOffmSecs = document.Global.Hardware.HeartBeat.LEDOffmsecs

	// my stupid work arround for null uint xml unmarshelling problem with numbers so use strings and convert it 2 times
	temp6, _ := strconv.ParseUint(document.Global.Hardware.Buttons.TxButtonPin, 10, 64)
	TxButtonPin = uint(temp6)
	temp7, _ := strconv.ParseUint(document.Global.Hardware.Buttons.TxTogglePin, 10, 64)
	TxTogglePin = uint(temp7)
	temp8, _ := strconv.ParseUint(document.Global.Hardware.Buttons.UpButtonPin, 10, 64)
	UpButtonPin = uint(temp8)
	temp9, _ := strconv.ParseUint(document.Global.Hardware.Buttons.DownButtonPin, 10, 64)
	DownButtonPin = uint(temp9)
	temp10, _ := strconv.ParseUint(document.Global.Hardware.Buttons.PanicButtonPin, 10, 64)
	PanicButtonPin = uint(temp10)
	temp11, _ := strconv.ParseUint(document.Global.Hardware.Comment.CommentButtonPin, 10, 64)
	CommentButtonPin = uint(temp11)
	CommentMessageOff = document.Global.Hardware.Comment.CommentMessageOff
	CommentMessageOn = document.Global.Hardware.Comment.CommentMessageOn
	temp12, _ := strconv.ParseUint(document.Global.Hardware.Buttons.ChimesButtonPin, 10, 64)
	ChimesButtonPin = uint(temp12)

	LCDEnabled = document.Global.Hardware.LCD.Enabled
	LCDInterfaceType = document.Global.Hardware.LCD.InterfaceType
	LCDI2CAddress = document.Global.Hardware.LCD.I2CAddress
	LCDBackLightTimerEnabled = document.Global.Hardware.LCD.Enabled
	LCDBackLightTimeoutSecs = time.Duration(document.Global.Hardware.LCD.BackLightTimeoutSecs)

	// my stupid work arround for null uint xml unmarshelling problem with numbers so use strings and convert it 2 times
	temp13, _ := strconv.ParseUint(document.Global.Hardware.LCD.BackLightLEDPin, 10, 64)
	LCDBackLightLEDPin = int(temp13)

	LCDRSPin = document.Global.Hardware.LCD.RsPin
	LCDEPin = document.Global.Hardware.LCD.EPin
	LCDD4Pin = document.Global.Hardware.LCD.D4Pin
	LCDD5Pin = document.Global.Hardware.LCD.D5Pin
	LCDD6Pin = document.Global.Hardware.LCD.D6Pin
	LCDD7Pin = document.Global.Hardware.LCD.D7Pin

	OLEDEnabled = document.Global.Hardware.OLED.Enabled
	OLEDInterfacetype = document.Global.Hardware.OLED.InterfaceType
	OLEDDisplayRows = document.Global.Hardware.OLED.DisplayRows
	OLEDDisplayColumns = document.Global.Hardware.OLED.DisplayColumns
	OLEDDefaultI2cBus = document.Global.Hardware.OLED.DefaultI2CBus
	OLEDDefaultI2cAddress = document.Global.Hardware.OLED.DefaultI2CAddress
	OLEDScreenWidth = document.Global.Hardware.OLED.ScreenWidth
	OLEDScreenHeight = document.Global.Hardware.OLED.ScreenHeight
	OLEDCommandColumnAddressing = document.Global.Hardware.OLED.CommandColumnAddressing
	OLEDAddressBasePageStart = document.Global.Hardware.OLED.AddressBasePageStart
	OLEDCharLength = document.Global.Hardware.OLED.CharLength
	OLEDStartColumn = document.Global.Hardware.OLED.StartColumn

	GpsEnabled = document.Global.Hardware.GPS.Enabled
	Port = document.Global.Hardware.GPS.Port
	Baud = document.Global.Hardware.GPS.Baud
	TxData = document.Global.Hardware.GPS.TxData
	Even = document.Global.Hardware.GPS.Even
	Odd = document.Global.Hardware.GPS.Odd
	Rs485 = document.Global.Hardware.GPS.Rs485
	Rs485HighDuringSend = document.Global.Hardware.GPS.Rs485HighDuringSend
	Rs485HighAfterSend = document.Global.Hardware.GPS.Rs485HighAfterSend
	StopBits = document.Global.Hardware.GPS.StopBits
	DataBits = document.Global.Hardware.GPS.DataBits
	CharTimeOut = document.Global.Hardware.GPS.CharTimeOut
	MinRead = document.Global.Hardware.GPS.MinRead
	Rx = document.Global.Hardware.GPS.Rx

	PEnabled = document.Global.Hardware.PanicFunction.Enabled
	PFilenameAndPath = document.Global.Hardware.PanicFunction.FilenameAndPath

	if PEnabled && PFilenameAndPath == "" {
		path := defaultSharePath + "/soundfiles/alerts/alert.wav"
		if _, err := os.Stat(path); err == nil {
			PFilenameAndPath = path
		}
	}

	PMessage = document.Global.Hardware.PanicFunction.Message
	PVolume = document.Global.Hardware.PanicFunction.Volume
	PSendIdent = document.Global.Hardware.PanicFunction.SendIdent
	PSendGpsLocation = document.Global.Hardware.PanicFunction.SendGpsLocation
	PTxLockEnabled = document.Global.Hardware.PanicFunction.TxLockEnabled
	PTxlockTimeOutSecs = document.Global.Hardware.PanicFunction.TxLockTimeOutSecs

	AudioRecordEnabled = document.Global.Hardware.AudioRecordFunction.Enabled
	AudioRecordOnStart = document.Global.Hardware.AudioRecordFunction.RecordOnStart
	AudioRecordSystem = document.Global.Hardware.AudioRecordFunction.RecordSystem
	AudioRecordMode = document.Global.Hardware.AudioRecordFunction.RecordMode
	AudioRecordTimeout = document.Global.Hardware.AudioRecordFunction.RecordTimeout
	AudioRecordFromOutput = document.Global.Hardware.AudioRecordFunction.RecordFromOutput
	AudioRecordFromInput = document.Global.Hardware.AudioRecordFunction.RecordFromInput
	AudioRecordMicTimeout = document.Global.Hardware.AudioRecordFunction.RecordMicTimeout
	AudioRecordSavePath = document.Global.Hardware.AudioRecordFunction.RecordSavePath
	AudioRecordArchivePath = document.Global.Hardware.AudioRecordFunction.RecordArchivePath
	AudioRecordSoft = document.Global.Hardware.AudioRecordFunction.RecordSoft
	AudioRecordProfile = document.Global.Hardware.AudioRecordFunction.RecordProfile
	AudioRecordFileFormat = document.Global.Hardware.AudioRecordFunction.RecordFileFormat
	AudioRecordChunkSize = document.Global.Hardware.AudioRecordFunction.RecordChunkSize

	if TargetBoard != "rpi" {
		LCDBackLightTimerEnabled = false
	}

	if LCDBackLightTimerEnabled == true && (OLEDEnabled == false && LCDEnabled == false) {
		log.Println("Alert: Logical Error in LCDBacklight Timer Check XML config file")
		log.Fatal("Backlight Timer Enabled but both LCD and OLED disabled!\n")

	}

	if OLEDEnabled == true {
		Oled, err = goled.BeginOled(OLEDDefaultI2cAddress, OLEDDefaultI2cBus, OLEDScreenWidth, OLEDScreenHeight, OLEDDisplayRows, OLEDDisplayColumns, OLEDStartColumn, OLEDCharLength, OLEDCommandColumnAddressing, OLEDAddressBasePageStart)
	}

	log.Println("Successfully loaded XML configuration file into memory")

	for i := 0; i < len(document.Accounts.Account); i++ {
		if document.Accounts.Account[i].Default == true {
			log.Printf("info: Successfully Added Account %s to Index [%d]\n", document.Accounts.Account[i].Name, i)
		}
	}

	return nil
}

func printxmlconfig() {

	if PrintAccount {
		log.Println("info: ---------- Account Information -------- ")
		log.Println("info: Default     " + fmt.Sprintf("%t", Default))
		log.Println("info: Server      " + Server[0])
		log.Println("info: Username    " + Username[0])
		log.Println("info: Password    " + Password[0])
		log.Println("info: Insecure    " + fmt.Sprintf("%t", Insecure[0]))
		log.Println("info: Certificate " + Certificate[0])
		log.Println("info: Channel     " + Channel[0])
		log.Println("info: Ident       " + Ident[0])
	} else {
		log.Println("info: ---------- Account Information -------- SKIPPED ")
	}

	if PrintLogging {
		log.Println("info: -------- Logging & Daemonizing -------- ")
		log.Println("info: Output Device     " + OutputDevice)
		log.Println("info: Log File          " + LogFilenameAndPath)
		log.Println("info: Logging           " + Logging)
		log.Println("info: Loglevel          " + Loglevel)
		log.Println("info: Daemonize         " + fmt.Sprintf("%t", Daemonize))
		log.Println("info: CancellableStream " + fmt.Sprintf("%t", CancellableStream))
		log.Println("info: SimplexWithMute   " + fmt.Sprintf("%t", SimplexWithMute))
		log.Println("info: TxCounter         " + fmt.Sprintf("%t", TxCounter))
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
		log.Println("info: TTS PlayChimes         ", fmt.Sprintf("%t", TTSPlayChimes))
		log.Println("info: TTS PlayChimesFilenameAndPath ", TTSPlayChimesFilenameAndPath)
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
		log.Println("info: Event Sound Enabled    " + fmt.Sprintf("%t", EventSoundEnabled))
		log.Println("info: Event Sound Filename   " + EventSoundFilenameAndPath)
		log.Println("info: Alert Sound Enabled    " + fmt.Sprintf("%t", AlertSoundEnabled))
		log.Println("info: Alert Sound Filename   " + AlertSoundFilenameAndPath)
		log.Println("info: Alert Sound Volume     " + fmt.Sprintf("%v", AlertSoundVolume))
		log.Println("info: Incomming Beep Enabled " + fmt.Sprintf("%t", IncommingBeepSoundEnabled))
		log.Println("info: Incomming Beep File    " + IncommingBeepSoundFilenameAndPath)
		log.Println("info: Incomming Beep Volume  " + fmt.Sprintf("%v", IncommingBeepSoundVolume))
		log.Println("info: Roger Beep Enabled     " + fmt.Sprintf("%t", RogerBeepSoundEnabled))
		log.Println("info: Roger Beep File        " + RogerBeepSoundFilenameAndPath)
		log.Println("info: Roger Beep Volume      " + fmt.Sprintf("%v", RogerBeepSoundVolume))
		log.Println("info: Repeater Tone Enabled  " + fmt.Sprintf("%t", RepeaterToneEnabled))
		log.Println("info: Repeater Tone File     " + RepeaterToneFilenameAndPath)
		log.Println("info: Repeater Tone Volume   " + fmt.Sprintf("%v", RepeaterToneVolume))
		log.Println("info: Chimes Enabled         " + fmt.Sprintf("%t", ChimesSoundEnabled))
		log.Println("info: Chimes File            " + ChimesSoundFilenameAndPath)
		log.Println("info: Chimes Volume          " + fmt.Sprintf("%v", ChimesSoundVolume))
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
		log.Println("info: PlayChimes         " + fmt.Sprintf("%t", APIPlayChimes))
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
		log.Println("info: Voice Activity Led Pin " + fmt.Sprintf("%v", VoiceActivityLEDPin))
		log.Println("info: Participants Led Pin   " + fmt.Sprintf("%v", ParticipantsLEDPin))
		log.Println("info: Transmit Led Pin       " + fmt.Sprintf("%v", TransmitLEDPin))
		log.Println("info: Online Led Pin         " + fmt.Sprintf("%v", OnlineLEDPin))
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
		log.Println("info: Chimes Button Pin       " + fmt.Sprintf("%v", ChimesButtonPin))
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
		log.Println("info: Back Light Timer Timeout " + fmt.Sprintf("%v", LCDBackLightTimeoutSecs))
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
		log.Println("info: Port                   " + fmt.Sprintf("%s", Port))
		log.Println("info: Baud                   " + fmt.Sprintf("%v", Baud))
		log.Println("info: TxData                 " + fmt.Sprintf("%s", TxData))
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

	if PrintPanic {
		log.Println("info: ------------ PANIC Function -------------- ")
		log.Println("info: Panic Function Enable          " + fmt.Sprintf("%t", PEnabled))
		log.Println("info: Panic Sound Filename and Path  " + fmt.Sprintf("%s", PFilenameAndPath))
		log.Println("info: Panic Message                  " + fmt.Sprintf("%s", PMessage))
		log.Println("info: Panic Message Send Recursively " + fmt.Sprintf("%t", PRecursive))
		log.Println("info: Panic Volume                   " + fmt.Sprintf("%v", PVolume))
		log.Println("info: Panic Send Ident               " + fmt.Sprintf("%t", PSendIdent))
		log.Println("info: Panic Send GPS Location        " + fmt.Sprintf("%t", PSendGpsLocation))
		log.Println("info: Panic TX Lock Enabled          " + fmt.Sprintf("%t", PTxLockEnabled))
		log.Println("info: Panic TX Lock Timeout Secs     " + fmt.Sprintf("%v", PTxlockTimeOutSecs))
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
}
