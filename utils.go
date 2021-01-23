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
	"errors"
	"flag"
	"fmt"

	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/talkkonnect/gumble/gumble"
	term "github.com/talkkonnect/termbox-go"
	"github.com/talkkonnect/volume-go"
	"github.com/xackery/gomail"
	"github.com/glendc/go-external-ip"
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

func copyFile(source string, dest string) {
	temp, _ := ioutil.ReadFile(source)
	ioutil.WriteFile(dest, temp, 0777)

}

func deleteFile(source string) {
	err := os.Remove(source)
	if err != nil {
		log.Fatal("Alert: Cannot Remove Config File ", err)
	}
}

func localAddresses() {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print(fmt.Sprintf("error: localAddresses %v", err.Error()))
		return
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()

		if err != nil {
			log.Print(fmt.Sprintf("error: localAddresses %v", err.Error()))
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
		log.Println(fmt.Sprintf("error: Ping Error %s", err))
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

func playWavLocal(filepath string, playbackvolume int) error {
	origVolume, _ = volume.GetVolume(OutputDevice)
	var player string

	if path, err := exec.LookPath("aplay"); err == nil {
		player = path
	} else if path, err := exec.LookPath("paplay"); err == nil {
		player = path
	} else {
		return errors.New("Failed to find either aplay or paplay in PATH")
	}

	cmd := exec.Command(player, filepath)
	err := volume.SetVolume(playbackvolume, OutputDevice)

	if err != nil {
		return fmt.Errorf("error: set volume failed: %+v", err)
	}
	_, err = cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("error: cmd.Run() for %s failed with %s\n", player, err)
	}
	err = volume.SetVolume(origVolume, OutputDevice)

	if err != nil {
		return fmt.Errorf("error: set volume failed: %+v", err)
	}
	return nil
}

func sendviagmail(username string, password string, receiver string, subject string, message string) error {

	err := gomail.Send(username, password, []string{receiver}, subject, message)
	if err != nil {
		return fmt.Errorf("sending Email Via GMAIL Error")
	}

	go LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)

	return nil
}

func clearfiles() { // Testing os.Remove to delete files
	err := os.RemoveAll(`/avrec`)
	if err != nil {
		fmt.Println("error: cannot remove file error ",err)
		return
	}
}

func fileserve3() {
	port := flag.String("psox", "8085", "port to serve on")
	directory := flag.String("dsox", AudioRecordSavePath, "the directory of static file to host")
	//. "dot" or / or ./img or AudioRecordSavePath, AudioRecordArchivePath
	flag.Parse()
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(*directory)))
	//http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img/"))))
	// in case of problem with img dir
	time.Sleep(5 * time.Second)
	log.Println("debug: Serving Audio Files", *directory, "over HTTP port:", *port)
	log.Println("info: HTTP Server Waiting")
	// log.Fatal(http.ListenAndServe(":" + *port, nil))
	log.Fatal(http.ListenAndServe(":"+*port, mux))
}

func fileserve4() {
	port := flag.String("pavrec", "8086", "port to serve on")
	directory := flag.String("davrec", "/avrec", "the directory of static file to host")
	flag.Parse()
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(*directory)))
	//http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img/"))))
	// in case of problem with img dir
	time.Sleep(5 * time.Second)
	log.Println("debug: Serving Directory", *directory, "over HTTP port:", *port)
	log.Println("info: HTTP Server Waiting")
	// log.Fatal(http.ListenAndServe(":" + *port, nil))
	log.Fatal(http.ListenAndServe(":"+*port, mux))
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

func unzip(archive, target string) error {
	reader, err := zip.OpenReader(archive)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(target, 0755); err != nil {
		return err
	}

	for _, file := range reader.File {
		path := filepath.Join(target, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		fileReader, err := file.Open()
		if err != nil {
			return err
		}
		defer fileReader.Close()

		targetFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer targetFile.Close()

		if _, err := io.Copy(targetFile, fileReader); err != nil {
			return err
		}
	}

	return nil
}

func createDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			panic(err)
		}
	}
}

func clearDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*"))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = os.RemoveAll(file)
		if err != nil {
			//os.RemoveAll(dir) //  can do dir's
			return err
		}
	}
	return nil
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
		return false, err // Not Empty
		log.Println("debug: Dir is Not Empty", "%t")
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)  // empty
	if err == io.EOF {
		return true, nil
		log.Println("debug: Dir is Empty", "%t")
	}
	return false, err // Either not empty or error, suits both cases
}

func fileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		// exist
		return true
	}
	// not exist
	return false
}

func fileNotExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// not exist
		return true
	}
	// exist
	return false
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
		log.Fatal(err)
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

func between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func after(value string, a string) string {
	// Get substring after a string.
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:len(value)]
}

type dateTimeScheduleStruct struct {
	startDateTime string
	endDateTime   string
	matched       bool
	defaultLogic  bool
	stopOnMatch   bool
}

type dayScheduleStruct struct {
	dayint       int
	startTime    int
	endTime      int
	matched      bool
	defaultLogic bool
	stopOnMatch  bool
}

func dateTimeWithinRange(dateTimeSchedule dateTimeScheduleStruct) (bool, bool, bool, error) {
	var dateFormat string = "02/01/2006 15:04"
	startDateTime, err := time.Parse(dateFormat, dateTimeSchedule.startDateTime)
	if err != nil {
		return false, false, false, err
	}

	endDateTime, err := time.Parse(dateFormat, dateTimeSchedule.endDateTime)
	if err != nil {
		return false, false, false, err
	}

	checkDateTime, err := time.Parse(dateFormat, time.Now().Format("02/01/2006 15:04"))
	if err != nil {
		return false, false, false, err
	}
	log.Println("------")
	log.Println("debug: startdate is ", startDateTime, " enddate is ", endDateTime, " check date is ", checkDateTime)
	if startDateTime.Before(checkDateTime) && endDateTime.After(checkDateTime) {
		return true, dateTimeSchedule.defaultLogic, dateTimeSchedule.stopOnMatch, nil
	}
	return false, dateTimeSchedule.defaultLogic, dateTimeSchedule.stopOnMatch, nil
}

//func dayTimeWithinRange(startTime string, endTime string, dayCheck string, dateFormat string, defaultLogicDay string) (bool, error) {
func dayTimeWithinRange(dayTimeWithinRange dayScheduleStruct) (bool, bool, bool, error) {

	t1 := time.Now()
	t1Day := int(t1.Weekday())
	t1Minute := int((t1.Hour() * 60) + t1.Minute())

	log.Println("------")
	log.Println("debug: day is ", t1Day, " starttime is ", dayTimeWithinRange.startTime, " endtime is ", dayTimeWithinRange.endTime, " checkday is ", t1Day, " check time is ", t1Minute)

	if t1Day == dayTimeWithinRange.dayint && (dayTimeWithinRange.startTime <= t1Minute && dayTimeWithinRange.endTime >= t1Minute) {
		return true, dayTimeWithinRange.defaultLogic, dayTimeWithinRange.stopOnMatch, nil
	}
	return false, dayTimeWithinRange.defaultLogic, dayTimeWithinRange.stopOnMatch, nil
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
