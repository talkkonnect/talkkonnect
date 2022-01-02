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
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/jacobsa/go-serial/serial"
	"github.com/adrianmo/go-nmea"
)

type GSVDataStruct struct {
	PRNNumber int64
	SNR       int64
	Azimuth   int64
	Elevation int64
}

type GNSSDataStruct struct {
	Time       string
	Validity   string
	Lattitude  float64
	Longitude  float64
	Speed      float64
	Course     float64
	Date       string
	Variation  float64
	FixQuality string
	SatsInUse  int64
	SatsInView int64
	GSVData    [4]GSVDataStruct
}

var (
	RMCSentenceValid bool
	GGASentenceValid bool
	GSVSentenceValid bool
	goodGPSRead      bool
	GNSSData         GNSSDataStruct
)

func getGpsPosition(verbose bool) (bool, error) {
	RMCSentenceValid = false
	GGASentenceValid = false
	GSVSentenceValid = false
	goodGPSRead = false

	if Config.Global.Hardware.GPS.Enabled {

		if Config.Global.Hardware.GPS.Port == "" {
			return false, errors.New("gnss port not specified")
		}

		if Config.Global.Hardware.GPS.Even && Config.Global.Hardware.GPS.Odd {
			return false, errors.New("can't specify both even and odd parity")
		}

		parity := serial.PARITY_NONE

		if Config.Global.Hardware.GPS.Even {
			parity = serial.PARITY_EVEN
		} else if Config.Global.Hardware.GPS.Odd {
			parity = serial.PARITY_ODD
		}

		options := serial.OpenOptions{
			PortName:               Config.Global.Hardware.GPS.Port,
			BaudRate:               Config.Global.Hardware.GPS.Baud,
			DataBits:               Config.Global.Hardware.GPS.DataBits,
			StopBits:               Config.Global.Hardware.GPS.StopBits,
			MinimumReadSize:        Config.Global.Hardware.GPS.MinRead,
			InterCharacterTimeout:  Config.Global.Hardware.GPS.CharTimeOut,
			ParityMode:             parity,
			Rs485Enable:            Config.Global.Hardware.GPS.Rs485,
			Rs485RtsHighDuringSend: Config.Global.Hardware.GPS.Rs485HighDuringSend,
			Rs485RtsHighAfterSend:  Config.Global.Hardware.GPS.Rs485HighAfterSend,
		}

		f, err := serial.Open(options)

		if err != nil {
			Config.Global.Hardware.GPS.Enabled = false
			return false, errors.New("cannot open serial port")
		} else {
			defer f.Close()
		}

		if Config.Global.Hardware.GPS.TxData != "" {
			txData_, err := hex.DecodeString(Config.Global.Hardware.GPS.TxData)

			if err != nil {
				Config.Global.Hardware.GPS.Enabled = false
				return false, errors.New("cannot decode hex data")
			}

			log.Println("Sending: ", hex.EncodeToString(txData_))

			count, err := f.Write(txData_)

			if err != nil {
				return false, errors.New("error writing to serial port")
			} else {
				log.Printf("Wrote %v bytes\n", count)
			}

		}

		if Config.Global.Hardware.GPS.Rx {
			serialPort, err := serial.Open(options)
			if err != nil {
				log.Println("warn: Unable to Open Serial Port Error ", err)
			}

			defer serialPort.Close()

			reader := bufio.NewReader(serialPort)
			scanner := bufio.NewScanner(reader)

			for scanner.Scan() {
				s, err := nmea.Parse(scanner.Text())

				if err == nil {

					switch s.DataType() {

					case nmea.TypeRMC:
						{
							m := s.(nmea.RMC)
							if m.Latitude != 0 && m.Longitude != 0 && !RMCSentenceValid {
								RMCSentenceValid = true
								GNSSData.Date = fmt.Sprintf("%v", m.Date)
								GNSSData.Time = fmt.Sprintf("%v", m.Time)
								GNSSData.Validity = fmt.Sprintf("%v", m.Validity)
								GNSSData.Lattitude = m.Latitude
								GNSSData.Longitude = m.Longitude
								GNSSData.Speed = m.Speed
								GNSSData.Course = m.Course
								GNSSData.Variation = m.Variation

							}
						}
					case nmea.TypeGGA:
						{
							m := s.(nmea.GGA)
							if m.Latitude != 0 && m.Longitude != 0 && !GGASentenceValid {
								GGASentenceValid = true
								GNSSData.FixQuality = m.FixQuality
								GNSSData.SatsInUse = m.NumSatellites
							}
						}

					case nmea.TypeGSV:
						{
							m := s.(nmea.GSV)

							for i := range m.Info {
								if m.Info[i].SNR > 0 && !GSVSentenceValid {
									GNSSData.GSVData[i].PRNNumber = s.(nmea.GSV).Info[i].SVPRNNumber
									GNSSData.GSVData[i].SNR = s.(nmea.GSV).Info[i].SNR
									GNSSData.GSVData[i].Azimuth = s.(nmea.GSV).Info[i].Azimuth
									GNSSData.GSVData[i].Elevation = s.(nmea.GSV).Info[i].Elevation
									if i >= 3 {
										GSVSentenceValid = true
										GNSSData.SatsInView = m.NumberSVsInView
									}
								}
							}
						}
					}
				}
			}
			if RMCSentenceValid && GGASentenceValid && GSVSentenceValid {
				goodGPSRead = true
				log.Println("info: RMC Date                    ", GNSSData.Date)
				log.Println("info: RMC Time                    ", GNSSData.Time)
				log.Println("info: RMC Validity                ", GNSSData.Validity)
				log.Println("info: RMC Latitude DMS            ", GNSSData.Longitude)
				log.Println("info: RMC Longitude DMS           ", GNSSData.Lattitude)
				log.Println("info: RMC Speed                   ", GNSSData.Speed)
				log.Println("info: RMC Course                  ", GNSSData.Course)
				log.Println("info: RMC Variation               ", GNSSData.Variation)
				log.Println("info: GGA GPS Quality Indicator   ", GNSSData.FixQuality)
				log.Println("info: GGA No of Satellites in Use ", GNSSData.SatsInUse)
				log.Println("info: GSV No of Satellites View   ", GNSSData.SatsInView)
				for i := range GNSSData.GSVData {
					log.Println("info: GSV SVPRNNumber Satellite   ", i, " ", GNSSData.GSVData[i].PRNNumber)
					log.Println("info: GSV SNR         Satellite   ", i, " ", GNSSData.GSVData[i].SNR)
					log.Println("info: GSV Azimuth     Satellite   ", i, " ", GNSSData.GSVData[i].Azimuth)
					log.Println("info: GSV Elevation   Satellite   ", i, " ", GNSSData.GSVData[i].Elevation)
				}
			}
		} else {
			return false, errors.New("error parsing gnss module")
		}
		return goodGPSRead, nil
	}
	return false, errors.New("gnss not enabled")
}
