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
 * avrecord.go -> talkkonnect function to record audio and video with low cost USB web cameras.
 * Record incoming Mumble traffic with sox package. Record video and images with external
 * packages fswebcam, motion, ffmpeg or other.
 *
 */

package talkkonnect

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"sync"
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
		log.Println("info: No Old sox Instance is Running. It is OK to Start sox")
	} else {
		time.Sleep(1 * time.Second)
		log.Println("info: Old sox instance was Killed Before Running New")
	}

	CreateDirIfNotExist(AudioRecordSavePath)
	CreateDirIfNotExist(AudioRecordArchivePath)
	emptydirchk, err := DirIsEmpty(AudioRecordSavePath)
	if err == nil && emptydirchk == false {

		filezip := time.Now().Format("20060102150405") + ".zip"
		go zipit(AudioRecordSavePath+"/", AudioRecordArchivePath+"/"+filezip)
		log.Println("info: Archiving Old Audio Files to", AudioRecordArchivePath+"/"+filezip)
		time.Sleep(1 * time.Second)
		cleardir(AudioRecordSavePath)
	} else {
		log.Println("info: Audio Recording Folder Is Empty. No Old Files to Archive")
	}
	time.Sleep(1 * time.Second)
	//go audiorecordtrafficmux()
	go audiorecordtraffic()
	log.Println("info: sox is Recording Traffic to", AudioRecordSavePath)
	//go func() {
	//fileserve3mux()  //8085  mp3, wav files // mux?
	//fileserve3()   //8085  mp3, wav files
	//log.Println("info: Running http server on port 8085")
	//}()
	return
}

/*quitfs := make(chan struct{}) // for sanitizing fileserve
  go func() { // to run fileserve
  for {
  select {
  case <-quitfs:
      return
  default:
  fileserve3()   //8085  mp3, wav files
  close(quitfs) // stop fileserve too.
*/

// Record ambient audio from microphone with sox

func AudioRecordAmbient() {

	CreateDirIfNotExist(AudioRecordSavePath)
	CreateDirIfNotExist(AudioRecordArchivePath)
	emptydirchk, err := DirIsEmpty(AudioRecordSavePath)
	if err == nil && emptydirchk == false {
		filezip := time.Now().Format("20060102150405") + ".zip"
		go zipit(AudioRecordSavePath+"/", AudioRecordArchivePath+"/"+filezip) // path to end with "/" or not?
		log.Println("info: Archiving Old Audio Files to", AudioRecordArchivePath+"/"+filezip)
		time.Sleep(1 * time.Second)
		cleardir(AudioRecordSavePath) // Remove old files
	} else {
		log.Println("info: Audio Recording Folder Is Empty. No Old Files to Archive")
	}
	time.Sleep(1 * time.Second)
	go audiorecordambientmux()
	//go audiorecordambient()
	//log.Println("info: sox is Recording Ambient Audio from Mic to", AudioRecordSavePath )
	//go func() {
	//fileserve3()   //8085  mp3, wav files
	//log.Println("info: Running http server on port 8085")
	//}()
	return
}

// Record both incoming Mumble traffic and ambient audio with sox

func AudioRecordCombo() {

	CreateDirIfNotExist(AudioRecordSavePath)
	CreateDirIfNotExist(AudioRecordArchivePath)
	emptydirchk, err := DirIsEmpty(AudioRecordSavePath)
	if err == nil && emptydirchk == false {
		filezip := time.Now().Format("20060102150405") + ".zip"
		go zipit(AudioRecordSavePath+"/", AudioRecordArchivePath+"/"+filezip)
		log.Println("info: Archiving Old Audio Files to", AudioRecordArchivePath+"/"+filezip)
		time.Sleep(1 * time.Second)
		cleardir(AudioRecordSavePath)
	} else {
		log.Println("info: Audio Recording Folder Is Empty. No Old Files to Archive")
	}
	time.Sleep(1 * time.Second)
	go audiorecordcombomux()
	//go audiorecordcombo()
	//log.Println("info: sox is Recording Traffic and Ambient Audio Mix to", AudioRecordSavePath )
	//go func() {
	//fileserve3()   //8085  mp3, wav files
	//log.Println("info: Running http server on port 8085")
	//}()
	return
}

/*
// TO DO. helper to break sox arguments to different "profiles" and select slices with settings to command sox.

func getsoxargs() []string {

	audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat

	args0 := []string{"-t", "alsa", AudioRecordFromOutput, "-t", AudioRecordFileFormat, audrecfile} // "standard". Just record.

	silence := []string{"silence", "1", "1", "2%", "-1", "0.5", "2%"}      // Detect and omit silence from recording

	chunks := []string{"trim", "0", "300", ":", "newfile", ":", "restart"} // Break recording to chunks

	args1 := append(args0, silence...)                                     // "vox-trimsilence". Detect vox and trim silence
	args2 := append(args0, chunks...)                                      // "chunks". Make file chunks
	args3 := append(args1, chunks...)                                      // "vox-trimsilence-chunks". vox, trim silence and make chunks

	if AudioRecordMode == "standard" {
		//args := args0
		fmt.Print(args0)
	}

	if AudioRecordMode == "vox-trimsilence" {
		args := args1
		fmt.Print(args)

	}
	if AudioRecordMode == "chunks" {
		args := args2
		fmt.Println(args)
	}
	if AudioRecordMode == "vox-trimsilence-chunks" {
		args := args3
		fmt.Println(args)
	}

	return args
}

// test: getsoxargs() should return sox arguments.
// it should return something like this [-t alsa hw:1,0 -t mp3 20091110230000.mp3 trim 0 300 : newfile : restart]

*/

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
		log.Println("info: sox is Missing. Is the Package Installed?")
	}

	//Need apt-get install sox libsox-fmt-mp3 (lame) (or pulseaudio and libsox-fmt-pulse)

	/*Need snd-aloop Alsa module. Should be standard kernel module. Just load snd-aloop
		and define loopback, dsnoop and dmix in .asoundrc or /usr/share/alsa/alsa.conf
	        modprobe modprobe snd-aloop (add to rc.local)
	        or echo "modprobe snd-aloop index=0 pcm_substreams=1" | sudo tee -a /etc/modules
	        or echo 'snd-aloop' >> /etc/modules
		or create snd-aloop.conf in /etc/modprobe.d/
	        to load snd-aloop on startup. For quick test just load with modprobe, remove with modprobe -r.
	*/

	audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat
	log.Println("info: sox is Recording Traffic to", AudioRecordSavePath+"/"+audrecfile)
	log.Println("info: Audio Recording Mode:", AudioRecordMode)

	if AudioRecordTimeout != 0 { // Record traffic, but stop it after timeout, if specified. "0" no timeout.

		/*	//TEST. Can switch be used for sox args? Need function to return args?
			//var args []string
			switch AudioRecordMode  {
			case "standard":
			if AudioRecordMode == "standard" {
			args := []string{"-t", "alsa", AudioRecordFromOutput, "-t", AudioRecordFileFormat, audrecfile}
			log.Println("info: standard", args)
			}
			case "vox-trimsilence":
			if AudioRecordMode == "vox-trimsilence" {
			args := []string{"-t", "alsa", AudioRecordFromOutput, "-t", AudioRecordFileFormat, audrecfile, "silence", "1", "1", `2%`, "-1", "0.5", `2%`}
			log.Println("info: vox-trimsilence", args)
			}
			case "chunks":
			if AudioRecordMode == "chunks" {
			args := []string{"-t", "alsa", AudioRecordFromOutput, "-t", AudioRecordFileFormat, audrecfile, "trim", "0", "300", ":", "newfile", ":", "restart"}
			log.Println("info: chunks", args)
			}
			case "vox-trimsilence-chunks":
			if AudioRecordMode == "vox-trimsilence-chunks" {
			args := []string{"-t", "alsa", AudioRecordFromOutput, "-t", AudioRecordFileFormat, audrecfile, "silence", "1", "1", `2%`, "-1", "0.5", `2%`,  "trim", "0", "300", ":", "newfile", ":", "restart"}
			log.Println("info: vox-trimsilence-chunks", args)
			}
			default:
			args := []string{"-t", "alsa", AudioRecordFromOutput, "-t", AudioRecordFileFormat, audrecfile}
			log.Println("info: standard", args)
			}
		*/
		args := []string{"-t", "alsa", AudioRecordFromOutput, "-t", AudioRecordFileFormat, audrecfile, "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"}

		log.Println("info: sox Arguments: " + fmt.Sprint(strings.Trim(fmt.Sprint(args), "[]")))
		log.Println("info: Traffic Recording will Timeout After:", AudioRecordTimeout, "seconds")

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
			log.Println("info: sox Error:", err)
			if signaled {
				log.Println("info: sox Signal:", signal)
			} else {
				log.Println("info: sox Status:", exitStatus)
			}
			close(done)
			// Did sox close ?
			log.Println("info: sox Stopped Recording Traffic to", AudioRecordSavePath)
		}()
		cmd.Process.Kill()
		<-done

	} else { // if AudioRecordTimeout is zero? Just keep recording until there is disk space on media.

		audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat // mp3, wav

		args := []string{"-t", "alsa", AudioRecordFromOutput, "-t", "mp3", audrecfile, "silence", "1", "1", "2%", "-1", "0.5", "2%", "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"}

		//getsoxargs()
		//fmt.Println(getsoxargs())
		//fmt.Sprint(strings.Trim(fmt.Sprint(getsoxargs()),"[]"))

		//args := []string{"-t", "alsa", "hw:Loopback,1,0", "-t", "mp3", audrecfile}         				               // Continous
		//test args := []string{"-t", "alsa", "hw:Loopback,1,0", "-t", "mp3", test.mp3, "silence", "1", "1", "2%", "-1", "0.5", "2%", "trim" ,"0" , "300", ":", "newfile", ":", "restart"}
		//Break recording to 5 min audio chunks. Detect silence.

		//devices "loopin", "loopout", "plughw:1,0, hw:Loopback,1,0, plug:dsnoop(er)...
		//args := []string{"-t", "alsa", AudioRecordFromOutput, "-t", AudioRecordFileFormat, audrecfile, "trim", "0" ,"300",":", "newfile", ":", "restart"}  // Continous with 5 min audio chunks
		//args := []string{"-t", "alsa", "loopout", "-t", "mp3", audrecfile,"silence", "1", "1", `2%`, "-1", "0.5", `2%`}      // skip silence.
		//Break recording to 5 min audio chunks.
		//args := []string{"-t", "alsa", "loopout", "-t", "mp3", audrecfile, "silence", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0" ,"300",":", "newfile", ":", "restart"}

		//Break recording to 5 minute chunks. And omitt silence. Requires approximatelly 1MB / minute of disk space.
		//args := []string{`--combine`, "mix", "-t", "alsa", "loopout", "-t", "plug:dsnoop", "-t", "mp3", audrecfile, "silence", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0" ,"300",":", "newfile", ":", "restart"}

		//Break recording to 5 minute chunks, omit silence, record both incomming traffic and mic input. For Alsa test config 1. (multi and dsnoop)
		//args := []string{"-m", "-t", "alsa", "loopout", "-t", "alsa", "plug:dsnooped", "-t", "mp3", audrecfile, "silence", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0" , "300",":", "newfile", ":", "restart"}

		//Break recording to 5 minute chunks, omit silence, record both incomming traffic and mic input. For Alsa test config 2 (dmix and multi)
		//args := []string{"-m", "-t", "alsa", "AudioRecordFromOutput", "-t", "alsa", "AudioRecordFromInput", "-t", "mp3", audrecfile, "silence", "-l", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0" , "300",":", "newfile", ":", "restart"}

		// if Alsa device paramerers are wrong or have changed, sox process will launch, but it
		// won't start recording to file. No file wils be created. Check for new file and log print if it doesn's start?
		// Check if recording dir is still empty after launching sox?

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = AudioRecordSavePath
		err := cmd.Start()
		check(err)
		time.Sleep(2 * time.Second)

		emptydirchk, err := DirIsEmpty(AudioRecordSavePath) // If sox didn't start recording for wrong parameters or any reason...  No  file.

		if err == nil && emptydirchk == false {
			log.Println("info: sox is Recording Traffic to", AudioRecordSavePath)
			log.Println("warn: sox will Go On Recording, Until it Runs out of Space or is Interrupted")

			starttime := time.Now()
			ticker := time.NewTicker(300 * time.Second) // Reminder if sox recording program is still recording after ... 5 minutes (no timeout)

			go func() {
				for range ticker.C {
					checked := time.Since(starttime)
					checkedshort := fmt.Sprintf(before(fmt.Sprint(checked), ".")) // trim  milliseconds after .  Format 00h00m00s.
					//elapsed := checked.Sub(starttime)
					//log.Println("info: sox is Still Running. Time:", elapsed)
					elapsed := fmtDuration(checked) // hh:mm format
					//fmt.Println(elapsedn)
					//log.Println("info: sox is Still Running. Time:", elapsed[:9])
					log.Println("info: sox is Still Running After", checkedshort+"s", "|", elapsed)
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
		log.Println("info: sox is Missing. Is the Package Installed?")
	}

	//Need apt-get install sox libsox-fmt-mp3 (lame)

	audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat // mp3, wav

	log.Println("info: sox is Recording Ambient Audio to", AudioRecordSavePath+"/"+audrecfile)
	log.Println("info: sox Audio Recording will Stop After", fmt.Sprint(AudioRecordMicTimeout), "seconds")

	if AudioRecordMicTimeout != 0 { // Record ambient audio, but stop it after timeout, if specified. "0" no timeout.

		args := []string{"-t", "alsa", AudioRecordFromInput, "-t", "mp3", audrecfile, "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"}

		cmd := exec.Command("/usr/bin/sox", args...)

		//cmd := exec.Command("/root/soxrecord1.sh")
		//cmd := exec.Command("bash", "-c", "soxrecord1.sh")

		/*Some cmd examples:
		sox -t alsa default -t test.wav
		sox -t alsa dsnoop -t test.wav
		sox -t pulse default -t mp3 test.mp3 silence 1 1 2% -1 0.5 2%
		sox -t alsa plughw:1,0 -c 1 -r 16k r.wav silence 1 0.1 0.3% 1 3.0 0.3% : newfile : restart
		sox -t alsa default -c 1 -r 16k r.wav
		sox --combine mix -t alsa loopout -t alsa plug:dsnoop -t mp3 test.mp3 silence 1 1 2% -1 0.5 2% trim 0 30 : newfile : restart
		arecord -D plughw:1,0 -f S16_LE $(date +%Y%m%d-%H%M%S).wav
		*/

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
			log.Println("info: sox Error:", err)
			if signaled {
				log.Println("info: sox Signal:", signal)
			} else {
				log.Println("info: sox Status:", exitStatus)
			}
			close(done)
			// Did sox close ?
			log.Println("info: sox Stopped Recording Traffic to", AudioRecordSavePath)
		}()
		cmd.Process.Kill()
	} else {
		audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat // mp3, wav
		//args := []string{"-t", "alsa", "plughw:1,0", "-t", "mp3", audrecfile} 	// Continous
		//args := []string{"-t", "alsa", "plug:dsnooper", "-t", "mp3", audrecfile,"silence", "1", "1", `2%`, "-1", "0.5", `2%`}   // skip silence.
		//args := []string{"-t", "alsa", "plug:dsnooper", "-t", "mp3", audrecfile, "trim", "0" ,"300", ":", "newfile", ":", "restart"} 	// Continous with 5 min audio chunks

		args := []string{"-t", "alsa", "plug:dsnooper", "-t", "mp3", audrecfile, "silence", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"} // voice detect, trim silence with 5 min audio chunks

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = AudioRecordSavePath // save audio recording to dir
		err := cmd.Start()
		check(err)

		emptydirchk, err := DirIsEmpty(AudioRecordSavePath) // If sox didn't start recording for wrong parameters or any reason...  No file.

		if err == nil && emptydirchk == false {
			log.Println("info: sox is Recording Ambient Audio to", AudioRecordSavePath)
			log.Println("warn: sox will Go On Recording, Until it Runs out of Space or is Interrupted")

			starttime := time.Now()

			ticker := time.NewTicker(300 * time.Second) // reminder if program is still recording after ... 5 minutes

			go func() {
				for range ticker.C {
					checked := time.Since(starttime)
					checkedshort := fmt.Sprintf(before(fmt.Sprint(checked), ".")) // trim  milliseconds after .  Format 00h00m00s.
					//elapsed := checked.Sub(starttime)
					//log.Println("info: sox is Still Running. Time:", elapsed)
					elapsed := fmtDuration(checked) // hh:mm format
					//fmt.Println(elapsedn)
					//log.Println("info: sox is Still Running. Time:", elapsed[:9])
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
		log.Println("info: sox is Missing. Is the Package Installed?")
	}

	//Need apt-get install sox libsox-fmt-mp3 (lame)

	audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat
	log.Println("info: sox is Recording Traffic to", AudioRecordSavePath+"/"+audrecfile)
	log.Println("info: Audio Recording Mode:", AudioRecordMode)

	if AudioRecordTimeout != 0 { // Record traffic, but stop it after timeout, if specified. "0" no timeout.

		args := []string{"-m", "-t", "alsa", AudioRecordFromOutput, "-t", "alsa", AudioRecordFromInput, "-t", AudioRecordFileFormat, audrecfile, "silence", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"}

		log.Println("info: sox Arguments: " + fmt.Sprint(strings.Trim(fmt.Sprint(args), "[]")))
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
			log.Println("info: sox Error:", err)
			if signaled {
				log.Println("info: sox Signal:", signal)
			} else {
				log.Println("info: sox Status:", exitStatus)
			}
			close(done)
			// Did sox close ?
			log.Println("info: sox Stopped Recording Traffic to", AudioRecordSavePath)
		}()
		cmd.Process.Kill()
		<-done

	} else { // if AudioRecordTimeout is zero? Just keep recording until there is disk space on media.

		audrecfile := time.Now().Format("20060102150405") + "." + AudioRecordFileFormat // mp3, wav

		args := []string{"-m", "-t", "alsa", AudioRecordFromOutput, "-t", "alsa", AudioRecordFromInput, "-t", "mp3", audrecfile, "silence", "1", "1", `2%`, "-1", "0.5", `2%`, "trim", "0", AudioRecordChunkSize, ":", "newfile", ":", "restart"}

		cmd := exec.Command("/usr/bin/sox", args...)
		cmd.Dir = AudioRecordSavePath
		err := cmd.Start()
		check(err)
		time.Sleep(2 * time.Second)

		emptydirchk, err := DirIsEmpty(AudioRecordSavePath) // If sox didn't start recording for wrong parameters or any reason...  No files.

		if err == nil && emptydirchk == false {
			log.Println("info: sox is Recording Mixed Audio to", AudioRecordSavePath)
			log.Println("warn: sox will Go On Recording, Until it Runs out of Space or is Interrupted")

			starttime := time.Now()

			ticker := time.NewTicker(300 * time.Second) // Reminder if sox recordin program is still recording after ... 5 minutes (no timeout)

			go func() {
				for range ticker.C {
					checked := time.Since(starttime)
					checkedshort := fmt.Sprintf(before(fmt.Sprint(checked), ".")) // trim  milliseconds after .  Format 00h00m00s.
					//elapsed := checked.Sub(starttime)
					//log.Println("info: sox is Still Running. Time:", elapsed)
					elapsed := fmtDuration(checked) // hh:mm format
					//fmt.Println(elapsedn)
					//log.Println("info: sox is Still Running. Time:", elapsed[:9])
					log.Println("info: sox is Still Running After", checkedshort+"s", "|", elapsed)
				}
			}()

		} else {
			log.Println("error: Something Went Wrong... sox Traffic Recording was Launched but Encountered Some Problems")
			log.Println("warn: Check ALSA Sound Settings and sox Arguments")
		}
	}
}

//

func clearfiles() { // Testing os.Remove to delete files
	//err := os.Remove("/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/img/test.file")
	err := os.RemoveAll(`/avrec`)
	if err != nil {
		fmt.Println(err)
		return
	}
}

// http server test
/*func fileserve() {
	port := flag.String("p", "8083", "port to serve on")
	directory := flag.String("d", "./img", "the directory of static file to host")
	//. or /
	flag.Parse()

	http.Handle("/", http.FileServer(http.Dir(*directory)))
	//http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./img/"))))
	// in case of problem with img dir
        time.Sleep(3 * time.Second)
	log.Println("info: Serving Location", *directory, "over HTTP port:", *port)
	log.Println("info: HTTP Server Waiting")
	log.Fatal(http.ListenAndServe(":" + *port, nil))

// Error with Talkkonnect when serving files
// panic: http: multiple registrations for /
// When running alone with "go run" not a problem.
// Problem: defaultHTTPMux, doesn’t support multiple registrations.
// Try to fix with mux exclusion.
*/

// mux for server

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

// Serve audio recordings over 8085

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
	log.Println("info: Serving Audio Files", *directory, "over HTTP port:", *port)
	log.Println("info: HTTP Server Waiting")
	// log.Fatal(http.ListenAndServe(":" + *port, nil))
	log.Fatal(http.ListenAndServe(":"+*port, mux))
	//return
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
	log.Println("info: Serving Directory", *directory, "over HTTP port:", *port)
	log.Println("info: HTTP Server Waiting")
	// log.Fatal(http.ListenAndServe(":" + *port, nil))
	log.Fatal(http.ListenAndServe(":"+*port, mux))
	//return
}

// ZIP Compressing function

/* Usage:
zipit("/tmp/documents", "/tmp/backup.zip")
zipit("/tmp/report.txt", "/tmp/report-2015.zip")
unzip("/tmp/report-2015.zip", "/tmp/reports/")
Example from: https://gist.github.com/svett/424e6784facc0ba907ae
Reuse this for compressing logs, backup , etc.
*/

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

// Unzip. For future use.

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

// Helper to check if dirs for working with images/video exist. If not create.

func CreateDirIfNotExist(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			panic(err)
		}
	}
}

// Helper to Clear files from work dir

func ClearDir(dir string) error {
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

// Another function to os.Remove, delete all files in dir.

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

// Helper to check is directory empty?

func DirIsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err // Not Empty
		fmt.Println("Dir is Not Empty", "%t")
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)  // empty
	if err == io.EOF {
		return true, nil
		fmt.Println("Dir is Empty", "%t")
	}
	return false, err // Either not empty or error, suits both cases
}

// Check if some file exists or not. Maybe use later.

func FileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		// exist
		return true
	}
	// not exist
	return false
}

func FileNotExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// not exist
		return true
	}
	// exist
	return false
}

// Check if fswebcam, motion or other bin is available in system?
// Dont start function if they ar not installed.

func isCommandAvailable(name string) bool {
	cmd := exec.Command("/bin/sh", "-c", "command -v "+name)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// Simple err check help for cmd
func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//  Helper to round up duration time to 1h1m45s / 01:02 format
//  use when fmt printing sox recording times
func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	//d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	//s := m / time.Second
	return fmt.Sprintf("%02d:%02d", h, m) // show sec’s also?
}

// try to use time.Duration() and time.ParseDuration().time.String()
// instead to round up time format?

// Return before, between or after some strings.
// Trim for extracting desired values.

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

// /|\- Spinner for indicating program running
//go spinner(time.Duration(AudioRecordTimeout)*time.Millisecond)

func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}
