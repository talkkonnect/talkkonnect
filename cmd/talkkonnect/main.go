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
)

func main() {

	config := flag.String("config", "/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml", "full path to talkkonnect.xml configuration file")

	flag.Usage = talkkonnectusage
	flag.Parse()

	log.Println("info: Reading Config File: ", *config)
	talkkonnect.PreInit0(*config)
}

func talkkonnectusage() {
	fmt.Println("---------------------------------------------------------------------------------------")
	fmt.Println("Usage: talkkonnect [-config=[full path and file to talkkonnect.xml configuration file]]")
	fmt.Println("By Suvir Kumar <suvir@talkkonnect.com>")
	fmt.Println("For more information visit http://www.talkkonnect.com and github.com/talkkonnect")
	fmt.Println("---------------------------------------------------------------------------------------")
	fmt.Println("-config=/home/talkkonnect/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml")
	fmt.Println("-version for the version")
	fmt.Println("-help for this screen")
}
