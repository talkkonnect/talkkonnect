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
 * gps.go -> talkkonnect function to interface to USB GPS Neo6M
 */

package talkkonnect

import (
	"encoding/hex"
	"fmt"
	"github.com/talkkonnect/go-nmea"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"log"
	"errors"
	"strings"
)

func getGpsPosition(verbose bool) error {
	if GpsEnabled {

		if Port == "" {
			return errors.New("You Must Specify Port")
		}

		if Even && Odd {
			return errors.New("can't specify both even and odd parity")
		}

		parity := serial.PARITY_NONE

		if Even {
			parity = serial.PARITY_EVEN
		} else if Odd {
			parity = serial.PARITY_ODD
		}

		options := serial.OpenOptions{
			PortName:               Port,
			BaudRate:               Baud,
			DataBits:               DataBits,
			StopBits:               StopBits,
			MinimumReadSize:        MinRead,
			InterCharacterTimeout:  CharTimeOut,
			ParityMode:             parity,
			Rs485Enable:            Rs485,
			Rs485RtsHighDuringSend: Rs485HighDuringSend,
			Rs485RtsHighAfterSend:  Rs485HighAfterSend,
		}

		f, err := serial.Open(options)

		if err != nil {
			GpsEnabled = false
			return errors.New("Cannot Open Serial Port")
		} else {
			defer f.Close()
		}

		if TxData != "" {
			txData_, err := hex.DecodeString(TxData)

			if err != nil {
				GpsEnabled = false
				return errors.New("Cannot Decode Hex Data")
			}

			log.Println("Sending: ", hex.EncodeToString(txData_))

			count, err := f.Write(txData_)

			if err != nil {
				return errors.New("Error writing to serial port")
			} else {
				log.Println("Wrote %v bytes\n", count)
			}

		}

		if Rx {
			for {
				buf := make([]byte, 90)
				n, err := f.Read(buf)
				if err != nil {
					if err != io.EOF {
						return errors.New(fmt.Sprintf("Error reading from serial port: ", err))
					}
				} else {
					buf = buf[:n]
					// sentence format "$GPRMC,220516,A,5133.82,N,00042.24,W,173.8,231.8,130694,004.2,W*70"
					sentence := strings.TrimSpace(fmt.Sprintf("%s\n", buf))
					if len(sentence) > 0 && sentence[:6] == "$GPRMC" {
						s, err := nmea.Parse(sentence)
						if err != nil {
							log.Fatal(err)
						}
						m := s.(nmea.GPRMC)
						GPSTime = fmt.Sprintf("%v", m.Time)
						GPSDate = fmt.Sprintf("%v", m.Date)
						GPSLatitude = m.Latitude
						GPSLongitude = m.Longitude
						if verbose {
							log.Println("info: Time: ", m.Time)
							log.Println("info: Validity: ", m.Validity)
							log.Println("info: Latitude GPS: ", nmea.FormatGPS(m.Latitude))
							log.Println("info: Latitude DMS: ", nmea.FormatDMS(m.Latitude))
							log.Println("info: Longitude GPS: ", nmea.FormatGPS(m.Longitude))
							log.Println("info: Longitude DMS: ", nmea.FormatDMS(m.Longitude))
							log.Println("info: Speed: ", m.Speed)
							log.Println("info: Course: ", m.Course)
							log.Println("info: Date: ", m.Date)
							log.Println("info: Variation: ", m.Variation)
						}
						break
					}
				}
			}
		}
	} else {
		return errors.New("GPS Disabled in config")
	}
	return nil
}
