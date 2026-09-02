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
 * Library golang-evdev Copyright (c) 2016 Georgi Valkov. All rights reserved. See Copyright Message in Library source code.
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * usbkeyboard.go -> function in talkkonnect for reading from external usb keyboard
 */

package talkkonnect

func (b *Talkkonnect) USBKeyboard() {
	for _, devicePath := range Config.Global.Hardware.USBKeyboard.USBKeyboardPaths {
		go b.evdevKeyboardListener(devicePath, "USB Keyboard", USBKeyMap, Config.Global.Hardware.USBKeyboard.NumlockScanID)
	}
}
