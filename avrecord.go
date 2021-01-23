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

	createDirIfNotExist(AudioRecordSavePath)
	createDirIfNotExist(AudioRecordArchivePath)
	emptydirchk, err := dirIsEmpty(AudioRecordSavePath)
	if err == nil && emptydirchk == false {

		filezip := time.Now().Format("20060102150405") + ".zip"
		go zipit(AudioRecordSavePath+"/", AudioRecordArchivePath+"/"+filezip)
		log.Println("debug: Archiving Old Audio Files to", AudioRecordArchivePath+"/"+filezip)
		time.Sleep(1 * time.Second)
		cleardir(AudioRecordSavePath)
	} else {
		log.Println("debug: Audio Recording Folder Is Empty. No Old Files to Archive")
	}
	time.Sleep(1 * time.Second)
	go audiorecordtraffic()
	log.Println("debug: sox is Recording Traffic to", AudioRecordSavePath)

	return
}

// Record ambient audio from microphone with sox

func AudioRecordAmbient() {

	createDirIfNotExist(AudioRecordSavePath)
	createDirIfNotExist(AudioRecordArchivePath)
	emptydirchk, err := dirIsEmpty(AudioRecordSavePath)
	if err == nil && emptydirchk == false {
		filezip := time.Now().Format("20060102150405") + ".zip"
		go zipit(AudioRecordSavePath+"/", AudioRecordArchivePath+"/"+filezip) // path to end with "/" or not?
		log.Println("info: Archiving Old Audio Files to", AudioRecordArchivePath+"/"+filezip)
		time.Sleep(1 * time.Second)
		cleardir(AudioRecordSavePath) // Remove old files
	} else {
		log.Println("debug: Audio Recording Folder Is Empty. No Old Files to Archive")
	}
	time.Sleep(1 * time.Second)
	go audiorecordambientmux()

	return
}

// Record both incoming Mumble traffic and ambient audio with sox

func AudioRecordCombo() {

	createDirIfNotExist(AudioRecordSavePath)
	createDirIfNotExist(AudioRecordArchivePath)
	emptydirchk, err := dirIsEmpty(AudioRecordSavePath)
	if err == nil && emptydirchk == false {
		filezip := time.Now().Format("20060102150405") + ".zip"
		go zipit(AudioRecordSavePath+"/", AudioRecordArchivePath+"/"+filezip)
		log.Println("info: Archiving Old Audio Files to", AudioRecordArchivePath+"/"+filezip)
		time.Sleep(1 * time.Second)
		cleardir(AudioRecordSavePath)
	} else {
		log.Println("debug: Audio Recording Folder Is Empty. No Old Files to Archive")
	}
	time.Sleep(1 * time.Second)
	go audiorecordcombomux()

	return
}

//Record traffic with mux exclusion. Allow new start only if currently not running.

func audiorecordtrafficmux() { // check if mux for this is working?

	JobIsrunningMu.Lock()
	start := !jobIsRunning
	jobIsRunning = true
	JobIsrunningMu.Unlock()

	if start {
		go func() {
			audiorecordtraffic()
			JobIsrunningMu.Lock()
			jobIsRunning = false
			JobIsrunningMu.Unlock()
		}()
	} else {
		log.Println("info: Traffic Audio Recording is Already Running. Please Wait.")
	}
}

//  sox function for traffic recording

func audiorecordtraffic() {

	// check if external program is installed?
	checkfile := isCommandAvailable("/usr/bin/sox")
	if checkfile == false {
		log.Println("error: sox is Missing. Is the Package Installed?")
	}

	audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat
	log.Println("info: sox is Recording Traffic to", AudioRecordSavePath+"/"+audrecfile)
	log.Println("info: Audio Recording Mode:", AudioRecordMode)

	if AudioRecordTimeout != 0 { // Record traffic, but stop it after timeout, if specified. "0" for no timeout.

		args := []string{"-t", AudioRecordSystem, AudioRecordFromOutput, "-t", AudioRecordFileFormat, audrecfile, "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"}

		log.Println("debug: sox Arguments: " + fmt.Sprint(strings.Trim(fmt.Sprint(args), "[]")))
		log.Println("debug: Traffic Recording will Timeout After:", AudioRecordTimeout, "seconds")

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = AudioRecordSavePath
		err := cmd.Start()
		check(err)
		done := make(chan struct{})

		time.Sleep(time.Duration(AudioRecordTimeout) * time.Second) // let sox record for a time, then send kill signal
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
			log.Println("info: sox Stopped Recording Traffic to", AudioRecordSavePath)
		}()
		cmd.Process.Kill()
		<-done

	} else { // if AudioRecordTimeout is zero? Just keep recording until there is disk space on media.

		audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat // mp3, wav

		args := []string{"-t", AudioRecordSystem, AudioRecordFromOutput, "-t", "mp3", audrecfile, "silence", "-l", "1", "1", "2%", "-1", "0.5", "2%", "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"}

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = AudioRecordSavePath
		err := cmd.Start()
		check(err)
		time.Sleep(2 * time.Second)

		emptydirchk, err := dirIsEmpty(AudioRecordSavePath) // If sox didn't start recording for wrong parameters or any reason...  No  file.

		if err == nil && emptydirchk == false {
			log.Println("info: sox is Recording Traffic to", AudioRecordSavePath)
			log.Println("info: sox will Go On Recording, Until it Runs out of Space or is Interrupted")

			starttime := time.Now()
			ticker := time.NewTicker(300 * time.Second) // Reminder if sox recording program is still recording after ... 5 minutes (no timeout)

			go func() {
				for range ticker.C {
					checked := time.Since(starttime)
					checkedshort := fmt.Sprintf(before(fmt.Sprint(checked), ".")) // trim  milliseconds after.  Format 00h00m00s.
					elapsed := fmtDuration(checked)                               // hh:mm format
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
	if checkfile == false {
		log.Println("error: sox is Missing. Is the Package Installed?")
	}

	//Need apt-get install sox libsox-fmt-mp3 (lame)

	audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat // mp3, wav

	log.Println("info: sox is Recording Ambient Audio to", AudioRecordSavePath+"/"+audrecfile)
	log.Println("info: sox Audio Recording will Stop After", fmt.Sprint(AudioRecordMicTimeout), "seconds")

	if AudioRecordMicTimeout != 0 { // Record ambient audio, but stop it after timeout, if specified. "0" no timeout.

		args := []string{"-t", AudioRecordSystem, AudioRecordFromInput, "-t", "mp3", audrecfile, "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"}

		cmd := exec.Command("/usr/bin/sox", args...)

		cmd.Dir = AudioRecordSavePath // save audio recording
		err := cmd.Start()
		check(err)
		done := make(chan struct{})
		time.Sleep(time.Duration(AudioRecordMicTimeout) * time.Second) // let sox record for a time, then signal kill

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
			log.Println("info: sox Stopped Recording Traffic to", AudioRecordSavePath)
		}()
		cmd.Process.Kill()
	} else {
		audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat // mp3, wav

		args := []string{"-t", AudioRecordSystem, AudioRecordFromInput, "-t", "mp3", audrecfile, "silence", "-l", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"} // voice detect, trim silence with 5 min audio chunks

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = AudioRecordSavePath // save audio recording to dir
		err := cmd.Start()
		check(err)

		emptydirchk, err := dirIsEmpty(AudioRecordSavePath) // If sox didn't start recording for wrong parameters or any reason...  No file.

		if err == nil && emptydirchk == false {
			log.Println("info: sox is Recording Ambient Audio to", AudioRecordSavePath)
			log.Println("warn: sox will Go On Recording, Until it Runs out of Space or is Interrupted")

			starttime := time.Now()

			ticker := time.NewTicker(300 * time.Second) // reminder if program is still recording after ... 5 minutes

			go func() {
				for range ticker.C {
					checked := time.Since(starttime)
					checkedshort := fmt.Sprintf(before(fmt.Sprint(checked), ".")) // trim  milliseconds after .  Format 00h00m00s.
					elapsed := fmtDuration(checked)                               // hh:mm format
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
	if checkfile == false {
		log.Println("error: sox is Missing. Is the Package Installed?")
	}

	//Need apt-get install sox libsox-fmt-mp3 (lame)

	audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat
	log.Println("info: sox is Recording Traffic to", AudioRecordSavePath+"/"+audrecfile)
	log.Println("info: Audio Recording Mode:", AudioRecordMode)

	if AudioRecordTimeout != 0 { // Record traffic, but stop it after timeout, if specified. "0" no timeout.

		args := []string{"-m", "-t", AudioRecordSystem, AudioRecordFromOutput, "-t", AudioRecordSystem, AudioRecordFromInput, "-t", AudioRecordFileFormat, audrecfile, "silence", "-l", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"}

		log.Println("debug: sox Arguments: " + fmt.Sprint(strings.Trim(fmt.Sprint(args), "[]")))
		log.Println("info: Audio Combo Recording will Timeout After:", AudioRecordTimeout, "seconds")

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = AudioRecordSavePath
		err := cmd.Start()
		check(err)
		done := make(chan struct{})

		time.Sleep(time.Duration(AudioRecordTimeout) * time.Second) // let sox record for a time, then send kill signal

		go func() {
			err := cmd.Wait()
			status := cmd.ProcessState.Sys().(syscall.WaitStatus)
			exitStatus := status.ExitStatus()
			signaled := status.Signaled()
			signal := status.Signal()
			log.Println("erroe: sox Error:", err)
			if signaled {
				log.Println("debug: sox Signal:", signal)
			} else {
				log.Println("debug: sox Status:", exitStatus)
			}
			close(done)
			// Did sox close ?
			log.Println("info: sox Stopped Recording Traffic to", AudioRecordSavePath)
		}()
		cmd.Process.Kill()
		<-done

	} else { // if AudioRecordTimeout is zero? Just keep recording until there is disk space on media.

		audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat // mp3, wav

		args := []string{"-m", "-t", AudioRecordSystem, AudioRecordFromOutput, "-t", AudioRecordSystem, AudioRecordFromInput, "-t", "mp3", audrecfile, "silence", "-l", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"}

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = AudioRecordSavePath
		err := cmd.Start()
		check(err)
		time.Sleep(2 * time.Second)

		emptydirchk, err := dirIsEmpty(AudioRecordSavePath) // If sox didn't start recording for wrong parameters or any reason...  No files.

		if err == nil && emptydirchk == false {
			log.Println("info: sox is Recording Mixed Audio to", AudioRecordSavePath)
			log.Println("warn: sox will Go On Recording, Until it Runs out of Space or is Interrupted")

			starttime := time.Now()

			ticker := time.NewTicker(300 * time.Second) // Reminder if sox recordin program is still recording after ... 5 minutes (no timeout)

			go func() {
				for range ticker.C {
					checked := time.Since(starttime)
					checkedshort := fmt.Sprintf(before(fmt.Sprint(checked), ".")) // trim  milliseconds after .  Format 00h00m00s.
					elapsed := fmtDuration(checked)                               // hh:mm format
					log.Println("info: sox is Still Running After", checkedshort+"s", "|", elapsed)
				}
			}()

		} else {
			log.Println("error: Something Went Wrong... sox Traffic Recording was Launched but Encountered Some Problems")
			log.Println("warn: Check ALSA Sound Settings and sox Arguments")
		}
	}
}

func fileserve3mux() {

	JobIsrunningMu.Lock()
	start := !jobIsRunning
	jobIsRunning = true
	JobIsrunningMu.Unlock()

	if start {
		go func() {
			fileserve3()
			JobIsrunningMu.Lock()
			jobIsRunning = false
			JobIsrunningMu.Unlock()
		}()
	} else {
		log.Println("info: Ambient Audio Recording is Already Running. Please Wait.")
	}
}
