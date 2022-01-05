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
 * The Initial Developer of the Original Code is Zoran Dimitrijevic
 * Portions created by the Initial Developer are Copyright (C) Suvir Kumar. All Rights Reserved.
 *
 * Contributor(s):
 *
 * Zoran Dimitrijevic
 * Suvir Kumar
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * avrecord.go -> talkkonnect function to record audio and video with low cost USB web cameras.
 * Record incoming Mumble traffic with sox package. Record video and images with external
 * packages fswebcam, motion, ffmpeg or other.
 *
 */

package talkkonnect

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	jobIsRunning   bool // used for mux for motion, fswebcam, ffmpeg, sox
	JobIsrunningMu sync.Mutex
)

// Record incoming Mumble traffic with sox

func AudioRecordTraffic() {

	// Need a way to prevent multiple sox instances running, or kill old one.
	_, err := exec.Command("sh", "-c", "killall -SIGINT sox").Output()
	if err != nil {
		log.Println("debug: No Old sox Instance is Running. It is OK to Start sox")
	} else {
		time.Sleep(1 * time.Second)
		log.Println("debug: Old sox instance was Killed Before Running New")
	}

	createDirIfNotExist(Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
	createDirIfNotExist(Config.Global.Hardware.AudioRecordFunction.RecordArchivePath)
	emptydirchk, err := dirIsEmpty(Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
	if err == nil && !emptydirchk {

		filezip := time.Now().Format("20060102150405") + ".zip"
		go zipit(Config.Global.Hardware.AudioRecordFunction.RecordSavePath+"/", Config.Global.Hardware.AudioRecordFunction.RecordArchivePath+"/"+filezip)
		log.Println("debug: Archiving Old Audio Files to", Config.Global.Hardware.AudioRecordFunction.RecordArchivePath+"/"+filezip)
		time.Sleep(1 * time.Second)
		cleardir(Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
	} else {
		log.Println("debug: Audio Recording Folder Is Empty. No Old Files to Archive")
	}
	time.Sleep(1 * time.Second)
	go audiorecordtraffic()
	log.Println("debug: sox is Recording Traffic to", Config.Global.Hardware.AudioRecordFunction.RecordSavePath)

}

// Record ambient audio from microphone with sox

func AudioRecordAmbient() {

	createDirIfNotExist(Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
	createDirIfNotExist(Config.Global.Hardware.AudioRecordFunction.RecordArchivePath)
	emptydirchk, err := dirIsEmpty(Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
	if err == nil && !emptydirchk {
		filezip := time.Now().Format("20060102150405") + ".zip"
		go zipit(Config.Global.Hardware.AudioRecordFunction.RecordSavePath+"/", Config.Global.Hardware.AudioRecordFunction.RecordArchivePath+"/"+filezip) // path to end with "/" or not?
		log.Println("info: Archiving Old Audio Files to", Config.Global.Hardware.AudioRecordFunction.RecordArchivePath+"/"+filezip)
		time.Sleep(1 * time.Second)
		cleardir(Config.Global.Hardware.AudioRecordFunction.RecordSavePath) // Remove old files
	} else {
		log.Println("debug: Audio Recording Folder Is Empty. No Old Files to Archive")
	}
	time.Sleep(1 * time.Second)
	go audiorecordambientmux()

}

// Record both incoming Mumble traffic and ambient audio with sox

func AudioRecordCombo() {

	createDirIfNotExist(Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
	createDirIfNotExist(Config.Global.Hardware.AudioRecordFunction.RecordArchivePath)
	emptydirchk, err := dirIsEmpty(Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
	if err == nil && !emptydirchk {
		filezip := time.Now().Format("20060102150405") + ".zip"
		go zipit(Config.Global.Hardware.AudioRecordFunction.RecordSavePath+"/", Config.Global.Hardware.AudioRecordFunction.RecordArchivePath+"/"+filezip)
		log.Println("info: Archiving Old Audio Files to", Config.Global.Hardware.AudioRecordFunction.RecordArchivePath+"/"+filezip)
		time.Sleep(1 * time.Second)
		cleardir(Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
	} else {
		log.Println("debug: Audio Recording Folder Is Empty. No Old Files to Archive")
	}
	time.Sleep(1 * time.Second)
	go audiorecordcombomux()

}

//  sox function for traffic recording

func audiorecordtraffic() {

	// check if external program is installed?
	checkfile := isCommandAvailable("/usr/bin/sox")
	if !checkfile {
		log.Println("error: sox is Missing. Is the Package Installed?")
	}

	audrecfile := time.Now().Format("20060102150405") + "." + Config.Global.Hardware.AudioRecordFunction.RecordFileFormat
	log.Println("info: sox is Recording Traffic to", Config.Global.Hardware.AudioRecordFunction.RecordSavePath+"/"+audrecfile)
	log.Println("info: Audio Recording Mode:", Config.Global.Hardware.AudioRecordFunction.RecordMode)

	if Config.Global.Hardware.AudioRecordFunction.RecordTimeout != 0 { // Record traffic, but stop it after timeout, if specified. "0" for no timeout.

		args := []string{"-t", Config.Global.Hardware.AudioRecordFunction.RecordSystem, Config.Global.Hardware.AudioRecordFunction.RecordFromOutput, "-t", Config.Global.Hardware.AudioRecordFunction.RecordFileFormat, audrecfile, "trim", "0", Config.Global.Hardware.AudioRecordFunction.RecordChunkSize, ":", "newfile", ":", "restart"}

		log.Println("debug: sox Arguments: " + fmt.Sprint(strings.Trim(fmt.Sprint(args), "[]")))
		log.Println("debug: Traffic Recording will Timeout After:", Config.Global.Hardware.AudioRecordFunction.RecordTimeout, "seconds")

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = Config.Global.Hardware.AudioRecordFunction.RecordSavePath
		err := cmd.Start()
		check(err)
		done := make(chan struct{})

		time.Sleep(time.Duration(Config.Global.Hardware.AudioRecordFunction.RecordTimeout) * time.Second) // let sox record for a time, then send kill signal
		go func() {
			err := cmd.Wait()
			status := cmd.ProcessState.Sys().(syscall.WaitStatus)
			exitStatus := status.ExitStatus()
			signaled := status.Signaled()
			signal := status.Signal()
			log.Println("error: sox Error:", err)
			if signaled {
				log.Println("debug: sox Signal:", signal)
			} else {
				log.Println("debug: sox Status:", exitStatus)
			}
			close(done)
			// Did sox close ?
			log.Println("info: sox Stopped Recording Traffic to", Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
		}()
		cmd.Process.Kill()
		<-done

	} else { // if AudioRecordTimeout is zero? Just keep recording until there is disk space on media.

		audrecfile := time.Now().Format("20060102150405") + "." + Config.Global.Hardware.AudioRecordFunction.RecordFileFormat // mp3, wav

		args := []string{"-t", Config.Global.Hardware.AudioRecordFunction.RecordSystem, Config.Global.Hardware.AudioRecordFunction.RecordFromOutput, "-t", "mp3", audrecfile, "silence", "-l", "1", "1", "2%", "-1", "0.5", "2%", "trim", "0", Config.Global.Hardware.AudioRecordFunction.RecordChunkSize, ":", "newfile", ":", "restart"}

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = Config.Global.Hardware.AudioRecordFunction.RecordSavePath
		err := cmd.Start()
		check(err)
		time.Sleep(2 * time.Second)

		emptydirchk, err := dirIsEmpty(Config.Global.Hardware.AudioRecordFunction.RecordSavePath) // If sox didn't start recording for wrong parameters or any reason...  No  file.

		if err == nil && !emptydirchk {
			log.Println("info: sox is Recording Traffic to", Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
			log.Println("info: sox will Go On Recording, Until it Runs out of Space or is Interrupted")

			starttime := time.Now()
			ticker := time.NewTicker(300 * time.Second) // Reminder if sox recording program is still recording after ... 5 minutes (no timeout)

			go func() {
				for range ticker.C {
					checked := time.Since(starttime)
					checkedshort := fmt.Sprintf("%v", before(fmt.Sprint(checked), ".")) // trim  milliseconds after.  Format 00h00m00s.
					elapsed := fmtDuration(checked)                                     // hh:mm format
					log.Println("debug: sox is Still Running After", checkedshort+"s", "|", elapsed)
				}
			}()

		} else {
			log.Println("error: Something Went Wrong... sox Traffic Recording was Launched but Encountered Some Problems")
			log.Println("warn: Check ALSA Sound Settings and sox Arguments")
		}
	}
}

// If talkkonnect stops or hangs. Must close sox manually. No signaling to sox for closing in this case.
//Record traffic and Mic mux exclusion.  Allow new start only if currently not running.

func audiorecordambientmux() {

	JobIsrunningMu.Lock()
	start := !jobIsRunning
	jobIsRunning = true
	JobIsrunningMu.Unlock()

	if start {
		go func() {
			audiorecordambient()
			JobIsrunningMu.Lock()
			jobIsRunning = false
			JobIsrunningMu.Unlock()
		}()
	} else {
		log.Println("info: Ambient Audio Recording is Already Running. Please Wait.")
	}
}

// sox function for ambient recording

func audiorecordambient() {

	checkfile := isCommandAvailable("/usr/bin/sox")
	if !checkfile {
		log.Println("error: sox is Missing. Is the Package Installed?")
	}

	//Need apt-get install sox libsox-fmt-mp3 (lame)

	audrecfile := time.Now().Format("20060102150405") + "." + Config.Global.Hardware.AudioRecordFunction.RecordFileFormat // mp3, wav

	log.Println("info: sox is Recording Ambient Audio to", Config.Global.Hardware.AudioRecordFunction.RecordSavePath+"/"+audrecfile)
	log.Println("info: sox Audio Recording will Stop After", fmt.Sprint(Config.Global.Hardware.AudioRecordFunction.RecordMicTimeout), "seconds")

	if Config.Global.Hardware.AudioRecordFunction.RecordMicTimeout != 0 { // Record ambient audio, but stop it after timeout, if specified. "0" no timeout.

		args := []string{"-t", Config.Global.Hardware.AudioRecordFunction.RecordSystem, Config.Global.Hardware.AudioRecordFunction.RecordFromInput, "-t", "mp3", audrecfile, "trim", "0", Config.Global.Hardware.AudioRecordFunction.RecordChunkSize, ":", "newfile", ":", "restart"}

		cmd := exec.Command("/usr/bin/sox", args...)

		cmd.Dir = Config.Global.Hardware.AudioRecordFunction.RecordSavePath // save audio recording
		err := cmd.Start()
		check(err)
		done := make(chan struct{})
		time.Sleep(time.Duration(Config.Global.Hardware.AudioRecordFunction.RecordMicTimeout) * time.Second) // let sox record for a time, then signal kill

		go func() {
			err := cmd.Wait()
			status := cmd.ProcessState.Sys().(syscall.WaitStatus)
			exitStatus := status.ExitStatus()
			signaled := status.Signaled()
			signal := status.Signal()
			log.Println("error: sox Error:", err)
			if signaled {
				log.Println("debug: sox Signal:", signal)
			} else {
				log.Println("debug: sox Status:", exitStatus)
			}
			close(done)
			// Did sox close ?
			log.Println("info: sox Stopped Recording Traffic to", Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
		}()
		cmd.Process.Kill()
	} else {
		audrecfile := time.Now().Format("20060102150405") + "." + Config.Global.Hardware.AudioRecordFunction.RecordFileFormat // mp3, wav

		args := []string{"-t", Config.Global.Hardware.AudioRecordFunction.RecordSystem, Config.Global.Hardware.AudioRecordFunction.RecordFromInput, "-t", "mp3", audrecfile, "silence", "-l", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0", Config.Global.Hardware.AudioRecordFunction.RecordChunkSize, ":", "newfile", ":", "restart"} // voice detect, trim silence with 5 min audio chunks

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = Config.Global.Hardware.AudioRecordFunction.RecordSavePath // save audio recording to dir
		err := cmd.Start()
		check(err)

		emptydirchk, err := dirIsEmpty(Config.Global.Hardware.AudioRecordFunction.RecordSavePath) // If sox didn't start recording for wrong parameters or any reason...  No file.

		if err == nil && !emptydirchk {
			log.Println("info: sox is Recording Ambient Audio to", Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
			log.Println("warn: sox will Go On Recording, Until it Runs out of Space or is Interrupted")

			starttime := time.Now()

			ticker := time.NewTicker(300 * time.Second) // reminder if program is still recording after ... 5 minutes

			go func() {
				for range ticker.C {
					checked := time.Since(starttime)
					checkedshort := fmt.Sprintf("%v", before(fmt.Sprint(checked), ".")) // trim  milliseconds after .  Format 00h00m00s.
					elapsed := fmtDuration(checked)                                     // hh:mm format
					log.Println("info: sox is Still Running After", checkedshort+"s", "|", elapsed)
				}
			}()

		} else {
			log.Println("error: Something Went Wrong... sox Traffic Recording was Launched but Encountered Some Problems")
			log.Println("warn: Check ALSA Sound Settings and sox Arguments")
		}
	}
}

//Record traffic and Mic mux exclusion.  Allow new start only if currently not running.

func audiorecordcombomux() {

	JobIsrunningMu.Lock()
	start := !jobIsRunning
	jobIsRunning = true
	JobIsrunningMu.Unlock()

	if start {
		go func() {
			audiorecordcombo()
			JobIsrunningMu.Lock()
			jobIsRunning = false
			JobIsrunningMu.Unlock()
		}()
	} else {
		log.Println("info: Combo Audio Recording is Already Running. Please Wait.")
	}
}

// Record traffic and Mic.

func audiorecordcombo() {

	checkfile := isCommandAvailable("/usr/bin/sox")
	if !checkfile {
		log.Println("error: sox is Missing. Is the Package Installed?")
	}

	//Need apt-get install sox libsox-fmt-mp3 (lame)

	audrecfile := time.Now().Format("20060102150405") + "." + Config.Global.Hardware.AudioRecordFunction.RecordFileFormat
	log.Println("info: sox is Recording Traffic to", Config.Global.Hardware.AudioRecordFunction.RecordSavePath+"/"+audrecfile)
	log.Println("info: Audio Recording Mode:", Config.Global.Hardware.AudioRecordFunction.RecordMode)

	if Config.Global.Hardware.AudioRecordFunction.RecordTimeout != 0 { // Record traffic, but stop it after timeout, if specified. "0" no timeout.

		args := []string{"-m", "-t", Config.Global.Hardware.AudioRecordFunction.RecordSystem, Config.Global.Hardware.AudioRecordFunction.RecordFromOutput, "-t", Config.Global.Hardware.AudioRecordFunction.RecordSystem, Config.Global.Hardware.AudioRecordFunction.RecordFromInput, "-t", Config.Global.Hardware.AudioRecordFunction.RecordFileFormat, audrecfile, "silence", "-l", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0", Config.Global.Hardware.AudioRecordFunction.RecordChunkSize, ":", "newfile", ":", "restart"}

		log.Println("debug: sox Arguments: " + fmt.Sprint(strings.Trim(fmt.Sprint(args), "[]")))
		log.Println("info: Audio Combo Recording will Timeout After:", Config.Global.Hardware.AudioRecordFunction.RecordTimeout, "seconds")

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = Config.Global.Hardware.AudioRecordFunction.RecordSavePath
		err := cmd.Start()
		check(err)
		done := make(chan struct{})

		time.Sleep(time.Duration(Config.Global.Hardware.AudioRecordFunction.RecordTimeout) * time.Second) // let sox record for a time, then send kill signal

		go func() {
			err := cmd.Wait()
			status := cmd.ProcessState.Sys().(syscall.WaitStatus)
			exitStatus := status.ExitStatus()
			signaled := status.Signaled()
			signal := status.Signal()
			log.Println("error: sox Error:", err)
			if signaled {
				log.Println("debug: sox Signal:", signal)
			} else {
				log.Println("debug: sox Status:", exitStatus)
			}
			close(done)
			// Did sox close ?
			log.Println("info: sox Stopped Recording Traffic to", Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
		}()
		cmd.Process.Kill()
		<-done

	} else { // if AudioRecordTimeout is zero? Just keep recording until there is disk space on media.

		audrecfile := time.Now().Format("20060102150405") + "." + Config.Global.Hardware.AudioRecordFunction.RecordFileFormat // mp3, wav

		args := []string{"-m", "-t", Config.Global.Hardware.AudioRecordFunction.RecordSystem, Config.Global.Hardware.AudioRecordFunction.RecordFromOutput, "-t", Config.Global.Hardware.AudioRecordFunction.RecordSystem, Config.Global.Hardware.AudioRecordFunction.RecordFromInput, "-t", "mp3", audrecfile, "silence", "-l", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0", Config.Global.Hardware.AudioRecordFunction.RecordChunkSize, ":", "newfile", ":", "restart"}

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = Config.Global.Hardware.AudioRecordFunction.RecordSavePath
		err := cmd.Start()
		check(err)
		time.Sleep(2 * time.Second)

		emptydirchk, err := dirIsEmpty(Config.Global.Hardware.AudioRecordFunction.RecordSavePath) // If sox didn't start recording for wrong parameters or any reason...  No files.

		if err == nil && !emptydirchk {
			log.Println("info: sox is Recording Mixed Audio to", Config.Global.Hardware.AudioRecordFunction.RecordSavePath)
			log.Println("warn: sox will Go On Recording, Until it Runs out of Space or is Interrupted")

			starttime := time.Now()

			ticker := time.NewTicker(300 * time.Second) // Reminder if sox recordin program is still recording after ... 5 minutes (no timeout)

			go func() {
				for range ticker.C {
					checked := time.Since(starttime)
					checkedshort := fmt.Sprintf("%v", before(fmt.Sprint(checked), ".")) // trim  milliseconds after .  Format 00h00m00s.
					elapsed := fmtDuration(checked)                                     // hh:mm format
					log.Println("info: sox is Still Running After", checkedshort+"s", "|", elapsed)
				}
			}()

		} else {
			log.Println("error: Something Went Wrong... sox Traffic Recording was Launched but Encountered Some Problems")
			log.Println("warn: Check ALSA Sound Settings and sox Arguments")
		}
	}
}
