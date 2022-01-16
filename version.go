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
 * The Initial Developer of the Original Code is Suvir Kumar <suvir@talkkonnect.com>
 *
 * Portions created by the Initial Developer are Copyright (C) Suvir Kumar. All Rights Reserved.
 *
 * Contributor(s):
 *
 * Suvir Kumar <suvir@talkkonnect.com>
 * Zoran Dimitrijevic
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * xmlparser.go -> talkkonnect functionality to read from XML file and populate global variables
 */

package talkkonnect

const (
	talkkonnectVersion  string = "2.08.03"
	talkkonnectReleased string = "Jan 16 2022"
)

/* Release Notes
1. In Client.go moved the stream on start and tx on start with possible delay to the last item to run before keylistener
2. Now we can check board type and version and show in the talkkonnect banner screen
3. Fixed State Tracking of Cancellable Stream
4. Client.go stricter checking of gps enabled and traccar enabled before firing off go routines for handling gpstracking and traccar
5. Fix the update-talkkonnect.sh script to delete only the talkkonnect binary and not whole binary folder for faster use with vs code
*/
