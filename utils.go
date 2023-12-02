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
 * Code Copied from https://www.socketloop.com/tutorials/golang-convert-seconds-to-human-readable-time-format-example
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
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	externalip "github.com/glendc/go-external-ip"
	"github.com/kennygrant/sanitize"
	"github.com/talkkonnect/gumble/gumble"
	term "github.com/talkkonnect/termbox-go"
	"github.com/xackery/gomail"
)

func reset() {
	term.Sync()
}

func esc(str string) string {
	return sanitize.HTML(str)
}

func cleanstring(str string) string {
	return sanitize.Name(str)
}

func plural(count int, singular string) (result string) {
	if (count == 1) || (count == 0) {
		result = strconv.Itoa(count) + " " + singular + " "
	} else {
		result = strconv.Itoa(count) + " " + singular + "s "
	}
	return
}

func secondsToHuman(input int) (result string) {
	years := math.Floor(float64(input) / 60 / 60 / 24 / 7 / 30 / 12)
	seconds := input % (60 * 60 * 24 * 7 * 30 * 12)
	months := math.Floor(float64(seconds) / 60 / 60 / 24 / 7 / 30)
	seconds = input % (60 * 60 * 24 * 7 * 30)
	weeks := math.Floor(float64(seconds) / 60 / 60 / 24 / 7)
	seconds = input % (60 * 60 * 24 * 7)
	days := math.Floor(float64(seconds) / 60 / 60 / 24)
	seconds = input % (60 * 60 * 24)
	hours := math.Floor(float64(seconds) / 60 / 60)
	seconds = input % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	seconds = input % 60

	if years > 0 {
		result = plural(int(years), "year") + plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if months > 0 {
		result = plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if weeks > 0 {
		result = plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if days > 0 {
		result = plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if hours > 0 {
		result = plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else if minutes > 0 {
		result = plural(int(minutes), "minute") + plural(int(seconds), "second")
	} else {
		result = plural(int(seconds), "second")
	}

	return
}

func localAddresses() {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("error: localAddresses %v\n", err.Error())
		return
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()

		if err != nil {
			log.Printf("error: localAddresses %v\n", err.Error())
			continue
		}

		for _, a := range addrs {
			if i.Name != "lo" {
				log.Printf("info: %v %v\n", i.Name, a)
			}
		}
	}
}

func (b *Talkkonnect) pingconnectedserver() {

	resp, err := gumble.Ping(b.Address, time.Second*1, time.Second*5)

	if err != nil {
		log.Printf("error: Ping Error %s", err)
		return
	}

	major, minor, patch := resp.Version.SemanticVersion()

	log.Println("info: Server Address:         ", resp.Address)
	log.Println("info: Current Channel:        ", b.Client.Self.Channel.Name)
	log.Println("info: Server Ping:            ", resp.Ping)
	log.Println("info: Server Version:         ", major, ".", minor, ".", patch)
	log.Println("info: Server Users:           ", resp.ConnectedUsers, "/", resp.MaximumUsers)
	log.Println("info: Server Maximum Bitrate: ", resp.MaximumBitrate)
}

func sendviagmail(username string, password string, receiver string, subject string, message string) error {

	err := gomail.Send(username, password, []string{receiver}, subject, message)
	if err != nil {
		return fmt.Errorf("sending Email Via GMAIL Error")
	}

	if Config.Global.Hardware.TargetBoard == "rpi" {
		if LCDEnabled {
			LcdText = [4]string{"nil", "nil", "nil", "Sending Email"}
			go LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		}
		if OLEDEnabled {
			oledDisplay(false, 6, OLEDStartColumn, "Sending Email")
		}
	}

	return nil
}

func zipit(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	info, err := os.Stat(source)
	if err != nil {
		return nil
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			panic(err)
		}
	}
}

func cleardir(dir string) {
	// The target directory.
	//directory := CamImageSavePath	// path must end on "/"... fix for no "/"?
	directory := dir + "/" // path with "/"
	// Open the directory and read all its files.
	dirRead, _ := os.Open(directory)
	dirFiles, _ := dirRead.Readdir(0)
	// Loop over the directory's files.
	for index := range dirFiles {
		fileHere := dirFiles[index]
		// Get name of file and its full path.
		nameHere := fileHere.Name()
		fullPath := directory + nameHere
		// Remove the files.
		os.Remove(fullPath)
		log.Println("info: Removed file", fullPath)

	}
}

func dirIsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		log.Println("debug: Dir is Not Empty")
		return false, err // Not Empty
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)  // empty
	if err == io.EOF {
		log.Println("debug: Dir is Empty")
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func check(err error) {
	if err != nil {
		FatalCleanUp(err.Error())
	}
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	//d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	//s := m / time.Second
	return fmt.Sprintf("%02d:%02d", h, m) // show sec’s also?
}

func before(value string, a string) string { // used for sox time
	// Get substring before a string.
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	return value[0:pos]
}

func getMacAddr() ([]string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var as []string
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			as = append(as, a)
		}
	}
	return as, nil
}

func getOutboundIP() string {
	consensus := externalip.DefaultConsensus(nil, nil)
	ip, err := consensus.ExternalIP()
	if err == nil {
		return ip.String()
	}
	return "Could Not Get Public WAN IP"
}

func FileExists(filepath string) bool {

	fileinfo, err := os.Stat(filepath)

	if os.IsNotExist(err) {
		return false
	}

	return !fileinfo.IsDir()
}

func killSession() {
	time.Sleep(2 * time.Second)
	c := exec.Command("reset")
	c.Stdout = os.Stdout
	c.Run()
	os.Exit(0)
}

func checkRegex(regex string, compareto string) bool {
	match, err := regexp.MatchString(regex, compareto)
	if err != nil {
		log.Println("error: Cannot Match Regular Expression Error", err)
		return false
	}
	return match
}

func createFolderIfNotExists(folder string) {
	dir, err := os.Open(folder)
	if os.IsNotExist(err) {
		os.MkdirAll(folder, 0700)
		return
	}

	dir.Close()
}

func downloadIfNotExists(fileName string, text string, language string) {
	f, err := os.Open(fileName)
	if err != nil {
		url := fmt.Sprintf("http://translate.google.com/translate_tts?ie=UTF-8&total=1&idx=0&textlen=32&client=tw-ob&q=%s&tl=%s", url.QueryEscape(text), language)
		response, err := http.Get(url)
		if err != nil {
			log.Println("error: TTS Module URL Error ", err)
			return
		}
		defer response.Body.Close()

		output, err := os.Create(fileName)
		if err != nil {
			log.Println("error: TTS Module Create File Error ", err)
			return
		}
		defer output.Close()

		_, err = io.Copy(output, response.Body)
		if err != nil {
			log.Println("error: TTS Module IO Copy Error ", err)
			return
		}
		log.Printf("debug: TTS Module Created File %v From TTS Message=%v\n", fileName, text)
	} else {
		log.Printf("debug: TTS Module Used Existing File %v From TTS Message=%v\n", fileName, text)
	}
	defer f.Close()
}

func generateHashName(name string) string {
	hash := md5.Sum([]byte(name))
	return hex.EncodeToString(hash[:])
}

func checkGitHubVersion() string {

	tmpfileName := "githubversion.txt"

	if FileExists(tmpfileName) {
		err := os.Remove(tmpfileName)
		if err != nil {
			log.Println("error: Cannot Remove Version File so Cannot Check Current GitHub Version")
			return talkkonnectVersion
		}
	}

	file, err := os.Open(tmpfileName)

	if err != nil {
		defer file.Close()
		url := "https://raw.githubusercontent.com/talkkonnect/talkkonnect/main/version.go"
		response, err := http.Get(url)
		if err != nil {
			log.Println("error: Cannot Get Version from GitHub")
			return talkkonnectVersion
		}
		defer response.Body.Close()

		output, err := os.Create(tmpfileName)
		if err != nil {
			log.Println("error: Cannot Create Temporary File for Version Checking")
			return talkkonnectVersion
		}

		_, _ = io.Copy(output, response.Body)
	}

	fileContent, err := os.ReadFile(tmpfileName)
	if err != nil {
		log.Println("error: Cannot Read Temporary File for Version Checking")
		return talkkonnectVersion
	}

	temp := strings.Split(string(fileContent), "\n")

	for _, item := range temp {
		if checkRegex("talkkonnectVersion", item) {
			regex := regexp.MustCompile(`"(.*)"`)
			match := regex.FindStringSubmatch(item)[1]
			return match // this will return the version found on github
		}
	}

	return talkkonnectVersion
}

func checkSBCVersion() string {
	if Config.Global.Hardware.TargetBoard != "rpi" {
		return "unknown"
	}

	fileContent, err := os.ReadFile("/proc/device-tree/model")
	if err != nil {
		log.Println("error: Cannot Check Raspberry Pi Board Version")
		return "unknown"
	}

	return string(fileContent[:])
}

func findMQTTButton(findMQTTButton string) mqttPubButtonStruct {
	for _, button := range Config.Global.Software.RemoteControl.MQTT.Settings.Pubpayload.Mqtt {
		if button.Item == findMQTTButton && button.Enabled {
			return mqttPubButtonStruct{button.Item, button.Payload, button.Enabled}
		}
	}
	return mqttPubButtonStruct{"", "", false}
}

func txScreen() {
	if LCDEnabled {
		LcdText[0] = "Online/TX"
		LcdText[3] = "TX at " + time.Now().Format("15:04:05")
		LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
	}
	if OLEDEnabled {
		Oled.DisplayOn()
		LCDIsDark = false
		oledDisplay(false, 0, OLEDStartColumn, "Online/TX")
		oledDisplay(false, 3, OLEDStartColumn, "TX at "+time.Now().Format("15:04:05"))
		oledDisplay(false, 4, OLEDStartColumn, "")
		oledDisplay(false, 5, OLEDStartColumn, "")
		oledDisplay(false, 6, OLEDStartColumn, "Please Visit")
		oledDisplay(false, 7, OLEDStartColumn, "www.talkkonnect.com")
	}
}

func rxScreen(LastSpeaker string) {
	if LCDEnabled && Config.Global.Hardware.TargetBoard == "rpi" {
		GPIOOutPin("backlight", "on")
		lcdtext = [4]string{"nil", "", "", LastSpeaker + " " + time.Now().Format("15:04:05")}
		LcdDisplay(lcdtext, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
		BackLightTime.Reset(time.Duration(LCDBackLightTimeout) * time.Second)
	}
	if OLEDEnabled && Config.Global.Hardware.TargetBoard == "rpi" {
		Oled.DisplayOn()
		oledDisplay(false, 0, OLEDStartColumn, "Online/RX")
		oledDisplay(false, 3, OLEDStartColumn, LastSpeaker+" "+time.Now().Format("15:04:05"))
		oledDisplay(false, 4, OLEDStartColumn, "")
		oledDisplay(false, 5, OLEDStartColumn, "")
		oledDisplay(false, 6, OLEDStartColumn, "Please Visit")
		oledDisplay(false, 7, OLEDStartColumn, "www.talkkonnect.com")
		BackLightTime.Reset(time.Duration(LCDBackLightTimeout) * time.Second)
	}
	if Config.Global.Software.Beacon.Enabled {
		BeaconTime.Reset(time.Duration(time.Duration(Config.Global.Software.Beacon.BeaconTimerSecs) * time.Second))
	}
}

func joinedLeftScreen(user string, info string) {
	if LCDEnabled {
		LcdText[0] = "Online/RX"
		LcdText[2] = user
		LcdText[3] = info
		LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
	}
	if OLEDEnabled {
		Oled.DisplayOn()
		LCDIsDark = false
		oledDisplay(false, 0, OLEDStartColumn, "Online/RX")
		oledDisplay(false, 3, OLEDStartColumn, user)
		oledDisplay(false, 4, OLEDStartColumn, info)
		oledDisplay(false, 5, OLEDStartColumn, "")
		oledDisplay(false, 6, OLEDStartColumn, "Please Visit       ")
		oledDisplay(false, 7, OLEDStartColumn, "www.talkkonnect.com")
	}
}

func (b *Talkkonnect) VTMove(command string) {
	var TargetID uint32
	for Index := range Config.Accounts.Account[AccountIndex].Voicetargets.ID {
		targetCount := len(Config.Accounts.Account[AccountIndex].Voicetargets.ID) - 1
		if command == "up" {
			if Index <= targetCount {
				if Config.Accounts.Account[AccountIndex].Voicetargets.ID[Index].IsCurrent {
					Config.Accounts.Account[AccountIndex].Voicetargets.ID[CurrentIndex].IsCurrent = false
					if Index < targetCount {
						CurrentIndex = Index + 1
					}
					if Index == targetCount {
						CurrentIndex = 0
					}
					Config.Accounts.Account[AccountIndex].Voicetargets.ID[CurrentIndex].IsCurrent = true
					TargetID = Config.Accounts.Account[AccountIndex].Voicetargets.ID[CurrentIndex].Value
					break
				}
			}
		}
		if command == "down" {
			if Index >= 0 {
				if Config.Accounts.Account[AccountIndex].Voicetargets.ID[Index].IsCurrent {
					Config.Accounts.Account[AccountIndex].Voicetargets.ID[CurrentIndex].IsCurrent = false
					if Index > 0 {
						CurrentIndex = Index - 1
					}
					if Index == 0 {
						CurrentIndex = targetCount
					}
					Config.Accounts.Account[AccountIndex].Voicetargets.ID[CurrentIndex].IsCurrent = true
					TargetID = Config.Accounts.Account[AccountIndex].Voicetargets.ID[CurrentIndex].Value
					break
				}
			}
		}
	}
	b.cmdSendVoiceTargets(TargetID)
}

func stripRegex(in string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9-.:/_ ()]+")
	return reg.ReplaceAllString(in, "")
}

func UniqueSliceElements[T comparable](inputSlice []T) []T {
	uniqueSlice := make([]T, 0, len(inputSlice))
	seen := make(map[T]bool, len(inputSlice))
	for _, element := range inputSlice {
		if !seen[element] {
			uniqueSlice = append(uniqueSlice, element)
			seen[element] = true
		}
	}
	return uniqueSlice
}
