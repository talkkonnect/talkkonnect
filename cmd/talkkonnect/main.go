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
 * main.go -- The Main Function of talkkonnect project
 *
 */

package main

import (
	"flag"
	"github.com/talkkonnect/talkkonnect"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"runtime"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var serverindex = flag.String("serverindex", "0", "jump to server index [n]")


func main() {

	config := flag.String("config", "/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml", "full path to talkkonnect.xml configuration file")

	flag.Usage = talkkonnectusage
	flag.Parse()

  if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal("could not create CPU profile: ", err)
        }
        if err := pprof.StartCPUProfile(f); err != nil {
            log.Fatal("could not start CPU profile: ", err)
        }
        defer pprof.StopCPUProfile()
    }

    if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            log.Fatal("could not create memory profile: ", err)
        }
        runtime.GC() // get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            log.Fatal("could not write memory profile: ", err)
        }
        f.Close()
    }

	talkkonnect.Init(*config, *serverindex)



}

func talkkonnectusage() {
	fmt.Println("---------------------------------------------------------------------------------------")
	fmt.Println("Usage: talkkonnect [-config=[full path and file to talkkonnect.xml configuration file]]")
	fmt.Println("By Suvir Kumar <suvir@talkkonnect.com>")
	fmt.Println("For more information visit http://www.talkkonnect.com and github.com/talkkonnect")
	fmt.Println("---------------------------------------------------------------------------------------")
	fmt.Println("-config=/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml")
	fmt.Println("-serverindex=[n] for the index of the enabled server to connect to in XML file")
	fmt.Println("-version for the version")
	fmt.Println("-help for this screen")
}
