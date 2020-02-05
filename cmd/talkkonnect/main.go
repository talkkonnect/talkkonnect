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
	"path/filepath"
)

var exec string

func main() {
	// NOTE: If using symlinks and you've found yourself here, see the
	//  info in os.Executable about how its inconsistant with symlinks
	//   and implement a fix for your problem. Its better that way.
	var err error
	if exec, err = os.Executable(); err != nil {
		panic(err)
	}
	// Avoid hardcoding by using the executable name (nothing should change though, unless the user likes a different name)
	configPath := os.Getenv("HOME") + "/.config/" + filepath.Base(exec) + "/" + filepath.Base(exec) + ".xml"
	if _, err := os.Stat(configPath); err != nil {
		//Well, maybe we'll try the same dir as the executable...
		if _, err := os.Stat(filepath.Dir(exec) + "/" + filepath.Base(exec) + ".xml"); err == nil {
			configPath = filepath.Dir(exec) + "/" + filepath.Base(exec) + ".xml"
		} else if _, err := os.Stat(filepath.Dir(exec) + "/talkkonnect.xml"); err == nil {
			// Renamed binary, not renamed xml
			configPath = filepath.Dir(exec) + "/talkkonnect.xml"
		} else {
			_, err := os.Stat(os.Getenv("HOME") + "/.config/talkkonnect/talkkonnect.xml")
			if err == nil {
				configPath = os.Getenv("HOME") + "/.config/talkkonnect/talkkonnect.xml"
			} else {
				//I guess just set it here since it doesn't exist in any of our search paths
				//Hopefully the user set something that actually exists
				configPath = filepath.Dir(exec) + "/" + filepath.Base(exec) + ".xml"
			}

		}
	}
	config := flag.String("config", configPath, "full path to "+filepath.Base(exec)+".xml configuration file")
	flag.Usage = talkkonnectusage
	flag.Parse()

	log.Println("info: Reading Config File: ", *config)
	talkkonnect.PreInit0(*config)
}

func talkkonnectusage() {
	fmt.Println("---------------------------------------------------------------------------------------")
	fmt.Println("Usage: talkkonnect [-config=[full path and file to " + filepath.Base(exec) + ".xml configuration file]]")
	fmt.Println("By Suvir Kumar <suvir@talkkonnect.com>")
	fmt.Println("For more information visit http://www.talkkonnect.com and github.com/talkkonnect")
	fmt.Println("---------------------------------------------------------------------------------------")
	fmt.Println("-config=$HOME/.config/" + filepath.Base(exec) + "/" + filepath.Base(exec) + ".xml")
	fmt.Println("-version for the version")
	fmt.Println("-help for this screen")
}
