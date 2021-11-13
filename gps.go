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
 * Zoran Dimitrijevic
 *
 * My Blog is at www.talkkonnect.com
 * The source code is hosted at github.com/talkkonnect
 *
 * gps.go -> talkkonnect function to interface to U-blox Neo-6M, Neo-7 (VK-172), Neo-8 and
 * possibly other low cost serial or USB GPS boards from other manufacturers widely used
 * with Arduino and Raspberry Pi.
 *
 * Integration for tracking talkkonnect GPS enabled devices with Traccar
 * server from https://www.traccar.org
 */

package talkkonnect

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/jacobsa/go-serial/serial"
	hd44780 "github.com/talkkonnect/go-hd44780"
	"github.com/talkkonnect/go-nmea"
)

//GPS Related Global Variables
var (
	TraccarPortT55     string = "5005" // Old Traccar Client port 5005 for working with T55 Protocol
	TraccarPortOpenGTS string = "5159" // Traccar Client port 5159 for for working OpenGTS Protocol
	TraccarPortOsmAnd  string = "5055" // Traccar Client port 5055 for working with OsmAnd Protocol
	GPSTime            string
	GPSDate            string
	GPSLatitude        float64
	GPSLongitude       float64
	GPSSpeed           float64
	GPSCourse          float64
	GPSVariation       float64
)

var goodGPSRead bool = false

func getGpsPosition(verbose bool) (bool, error) {

	if Config.Global.Hardware.GPS.Enabled {

		if Config.Global.Hardware.GPS.Port == "" {
			return false, errors.New("you must specify port")
		}

		if Config.Global.Hardware.GPS.Even && Config.Global.Hardware.GPS.Odd {
			return false, errors.New("cant specify both even and odd parity")
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
			txData, err := hex.DecodeString(Config.Global.Hardware.GPS.TxData)

			if err != nil {
				Config.Global.Hardware.GPS.Enabled = false
				return false, errors.New("cannot decode hex data")
			}

			log.Println("info: Sending: ", hex.EncodeToString(txData))

			count, err := f.Write(txData)

			if err != nil {
				return false, errors.New("error writing to serial port")
			} else {
				log.Printf("info: Wrote %v bytes\n", count)
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

			goodGPSRead = false
			for scanner.Scan() {
				s, err := nmea.Parse(scanner.Text())

				if err == nil {

					if s.DataType() == nmea.TypeRMC {
						m := s.(nmea.RMC)
						if m.Latitude != 0 && m.Longitude != 0 {
							goodGPSRead = true

							FreqReport := float64(Config.Global.Hardware.GPSTrackingFunction.TraccarReportFrequency)                        // Reporting Frequency
							FreqReports := (time.Duration(Config.Global.Hardware.GPSTrackingFunction.TraccarReportFrequency) * time.Second) // Frequency of GPS Reporting. Minutes, Seconds or hours?

							if Config.Global.Hardware.GPSTrackingFunction.TrackEnabled && Config.Global.Hardware.GPSTrackingFunction.TraccarSendTo {
								if Config.Global.Hardware.GPSTrackingFunction.TraccarProto == "t55" {
									go tcpSendT55Traccar2() // Initial Send GPS position to Traccar with old T55 client protocol. No keep-alive.
								} else {
									go httpSendTraccar() // Initial Send GPS position to Traccar over http function for both OsmAnd or OpenGTS protocol.
								}
								log.Println("info: GPS Position Report Nr (1) Sent to Traccar Server")
							}

							PositionReporter := time.NewTicker(FreqReports)
							var TraccarCounter = 1
							go func() {
								for range PositionReporter.C {
									if Config.Global.Hardware.GPSTrackingFunction.TrackEnabled && Config.Global.Hardware.GPSTrackingFunction.TraccarSendTo {
										if Config.Global.Hardware.GPSTrackingFunction.TraccarProto == "t55" {
											tcpSendT55Traccar2() // Send GPS position to Traccar with old T55 client protocol. No keep-alive.
										} else {
											httpSendTraccar() // Send GPS position to Traccar over http function for both OsmAnd or OpenGTS protocol.
										}
									}
									TraccarCounter++

									if verbose {
										if Config.Global.Hardware.GPSTrackingFunction.TrackEnabled && Config.Global.Hardware.GPSTrackingFunction.TraccarSendTo {
											if Config.Global.Hardware.GPSTrackingFunction.TraccarProto == "osmand" {
												log.Println("info: OsmAnd: ", Config.Global.Hardware.GPSTrackingFunction.TraccarServerURL+":"+fmt.Sprint(TraccarPortOsmAnd)+"?"+"id="+Config.Global.Hardware.GPSTrackingFunction.TraccarClientId+"&"+"timestamp="+date2()+"%20"+time2()+"&"+"lat="+fmt.Sprint(m.Latitude)+"&"+"lon="+fmt.Sprint(m.Longitude)+"&"+"speed="+fmt.Sprint(m.Speed)+"&"+"course="+fmt.Sprint(m.Course)+"&"+"variation="+fmt.Sprint(m.Variation))
											} else if Config.Global.Hardware.GPSTrackingFunction.TraccarProto == "t55" {
												log.Println("info: T55: " + "Sending " + fmt.Sprint(m) + " to " + Config.Global.Hardware.GPSTrackingFunction.TraccarServerURL + ":" + fmt.Sprint(TraccarPortT55))
											} else if Config.Global.Hardware.GPSTrackingFunction.TraccarProto == "opengts" {
												log.Println("info: OpenGTS: ", Config.Global.Hardware.GPSTrackingFunction.TraccarServerURL+":"+fmt.Sprint(TraccarPortOpenGTS)+"/?"+"id="+Config.Global.Hardware.GPSTrackingFunction.TraccarClientId+"&"+fmt.Sprint(m))
											}
											log.Println("info: GPS Position Report Nr " + "(" + fmt.Sprint(TraccarCounter) + ")" + " Sent to Traccar Server. Next Position Report in " + fmt.Sprintf("%.2f", FreqReport/60) + " minute(s)")
										}
									}

									if Config.Global.Hardware.TargetBoard == "rpi" {
										if Config.Global.Hardware.GPSTrackingFunction.TrackEnabled {
											if Config.Global.Hardware.GPSTrackingFunction.TrackGPSShowLCD {
												log.Println("info: Showing GPS Info in LCD: " + "Lat: " + fmt.Sprint(m.Latitude) + " Long: " + fmt.Sprint(m.Longitude))
												time.Sleep(5 * time.Second)
												t := time.Now()
												if LCDEnabled {
													LcdText = [4]string{"nil", "GPS OK " + t.Format("15:04:05"), "lat:" + fmt.Sprintf("%f", m.Latitude), "lon:" + fmt.Sprintf("%f", m.Longitude) + " s:" + fmt.Sprintf("%.2f", m.Speed*1.852)} // 1 knot = 1.852 km.  Take LCD rows 1-3
													// Option: narrow down GPS LCD writing to rows 2-3. Row 2, status, row 3, grid.
													go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress) // Take LCD rows 2-3.
												}
												if OLEDEnabled {
													oledDisplay(false, 4, 1, "GPS OK "+t.Format("15:04:05"))
													oledDisplay(false, 5, 1, "lat: "+fmt.Sprintf("%f", m.Latitude))
													oledDisplay(false, 6, 1, "lon: "+fmt.Sprintf("%f", m.Longitude))
													oledDisplay(false, 7, 1, "sp: "+fmt.Sprintf("%.2f", (m.Speed*1.852)))
												}
											}
										}
									}

								}
							}()

							GPSDate := fmt.Sprintf("%v", m.Date)
							GPSTime := fmt.Sprintf("%v", m.Time)

							// Traccar needs timestamp in this format "timestamp=2019-01-15%2021:41:18"
							// for OsmAnd protocol. Date and time are concatenated in the same line with escape
							// %20 character in the middle. m.Date needs to change from
							// 19/10/04 to 2019-10-04, m.Time from GPS 23:27:15.0000s must have ms truncated.
							// date and time are parsed as timestapmp= ... and must use this format
							// "04-10-19 23:27:15" or "04-10-19%2023:27:15"

							//GPSDate = m.Date
							//GPSTime = m.Time

							GPSLatitude = m.Latitude
							GPSLongitude = m.Longitude
							GPSSpeed = m.Speed
							GPSCourse = m.Course
							GPSVariation = m.Variation

							Date1 := fmt.Sprint(gpsdatereorder()) // Reformatted date for Tracar
							Time1 := truncateString(GPSTime, 8)   // Truncate time for Traccar

							currentTime := time.Now()
							//Date2 := fmt.Sprintf(currentTime.Format("2006-01-02"))
							//Time2 := fmt.Sprintf("%s", currentTime.Format("15:04:05"))

							if verbose {
								// Testing date and time format
								log.Println("info: Date and Time from GPS: ", (GPSDate + " " + GPSTime))
								log.Println("info: Date and Time from GPS Reformatted for Traccar: ", (Date1 + " " + Time1)) // From GPS
								//log.Println("info: Date from Helper:", fmt.Sprint(gpsdatereorder()))                         // Need this date for Traccar
								log.Println("info: System Date with time.now Function:", currentTime.Format("2006-01-02"))
								log.Println("info: System Time with time.now function:", currentTime.Format("15:04:05"))
								log.Println("info: System Date & Time with time.now Function:", currentTime.Format("2006-01-02 15:04:05")) //currentTime.Format("2006-01-02 3:4:5")
								// Workaround for date time format. Preferablly always use GPS time, and only if not the system time.
								log.Println("info: GPS Date: ", m.Date)
								log.Println("info: GPS Time: ", m.Time)
								log.Println("info: Validity: ", m.Validity)
								log.Println("info: Latitude Decimal: ", m.Latitude)
								log.Println("info: Longitude Decimal: ", m.Longitude)
								log.Println("info: Latitude DMS: ", nmea.FormatDMS(m.Latitude))
								log.Println("info: Longitude DMS: ", nmea.FormatDMS(m.Longitude))
								log.Println("info: Latitude GPS: ", nmea.FormatGPS(m.Latitude))
								log.Println("info: Longitude GPS: ", nmea.FormatGPS(m.Longitude))
								log.Println("info: Speed: ", m.Speed) // Is this knots?
								log.Println("info: Course: ", m.Course)
								log.Println("info: Variation: ", m.Variation)
								log.Println("info: Traccar Cmd Osmand: " + fmt.Sprint(Config.Global.Hardware.GPSTrackingFunction.TraccarServerFullURL))
								log.Println("info: Traccar $GPRMC Sentence for T55/OpenGTS: " + fmt.Sprint(m))

								if Config.Global.Hardware.GPSTrackingFunction.TrackEnabled {
									//log.Println("info: GPS Tracking Enabled: " + fmt.Sprint(TrackEnabled))
									if Config.Global.Hardware.GPSTrackingFunction.TraccarSendTo {
										log.Println("info: Sending GPS Position to Traccar Server Enabled")
										log.Println("info: Traccar Protocol: " + strings.Title(strings.ToLower(Config.Global.Hardware.GPSTrackingFunction.TraccarProto)) + "; " + "Reporting Frequency: " + fmt.Sprintf("%.2f", FreqReport/60) + " minutes;")
										//Print GPS message format for sending to Traccar depending on tracking protocol.

										switch Config.Global.Hardware.GPSTrackingFunction.TraccarProto {
										case "osmand":
											log.Println("info: OsmAnd: ", Config.Global.Hardware.GPSTrackingFunction.TraccarServerFullURL)
										case "opengts":
											log.Println("info: OpenGTS: ", Config.Global.Hardware.GPSTrackingFunction.TraccarServerIP)
										case "t55":
											log.Println("info: T55:", fmt.Sprint(m), "...", Config.Global.Hardware.GPSTrackingFunction.TraccarServerURL+":"+TraccarPortT55)
										default:
											log.Println("info: OsmAnd: ", Config.Global.Hardware.GPSTrackingFunction.TraccarServerFullURL)
										}
									}
								}
							}

							break

						} else {
							log.Println("warn: Got Latitude 0 and Longtitude 0 from GPS")
						}
					} else {
						log.Println("warn: GPS Sentence Format Was not nmea.RMC")
					}
				} else {
					log.Println("warn: Scanner Function Error ", err)
				}
			}

			// } //

		} else {
			return false, errors.New("rx not set")
		}
		return goodGPSRead, nil
	}
	return false, errors.New("GPS Not Enabled. Or Not Connected")
}

func tcpSendT55Traccar2() {

	pgid := "$PGID" + "," + Config.Global.Hardware.GPSTrackingFunction.TraccarClientId + "*0F" + "\r" + "\n" // Unique Client ID (e.g. 12345). Follow with carriage return and line feed $
	gprmc := fmt.Sprint(m) + "\r" + "\n"
	log.Println("info: $GPRMC to send is: " + fmt.Sprint(m))
	fmt.Println(m)

	conn, _ := net.Dial("tcp", Config.Global.Hardware.GPSTrackingFunction.TraccarServerIP+":"+fmt.Sprint(TraccarPortT55)) // Use port 5005 for T55. Keep-alive.
	err := conn.(*net.TCPConn).SetKeepAlive(true)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = conn.(*net.TCPConn).SetKeepAlivePeriod(60 * time.Second)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = conn.(*net.TCPConn).SetNoDelay(false)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = conn.(*net.TCPConn).SetLinger(0)
	if err != nil {
		fmt.Println(err)
		return
	}

	log.Println("info: Traccar Client:", conn.LocalAddr().String(), "Connected to Server:", conn.RemoteAddr().String())

	fmt.Fprintf(conn, pgid) // Send ID
	time.Sleep(1 * time.Second)
	fmt.Fprintf(conn, gprmc) // send $GPRMC
	log.Println("info: Sending position message to Traccar over Protocol: " + strings.Title(strings.ToLower(Config.Global.Hardware.GPSTrackingFunction.TraccarProto)))

	notify := make(chan error)

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				notify <- err
				if io.EOF == err {
					close(notify)
					return
				}
			}

			if n > 0 {
				log.Printf("Unexpected Data: %s", buf[:n])
			}
		}
	}()

	for {
		select {
		case err := <-notify:
			log.Println("info: Traccar Server Connection dropped message", err)

			if err == io.EOF {
				log.Println("Connection to Traccar Server was closed")
				return
			}
		case <-time.After(time.Second * 60):
			log.Println("Traccar Server Connection Timeout 60. Still Alive")
		}
	}
}

func httpSendTraccar() {

	if Config.Global.Hardware.GPSTrackingFunction.TrackEnabled {
		if Config.Global.Hardware.GPSTrackingFunction.TraccarSendTo {
			if Config.Global.Hardware.GPSTrackingFunction.TraccarProto == "osmand" {
				Config.Global.Hardware.GPSTrackingFunction.TraccarServerFullURL = (fmt.Sprint(Config.Global.Hardware.GPSTrackingFunction.TraccarServerURL) + ":" + fmt.Sprint(TraccarPortOsmAnd) + "?" + "id=" + Config.Global.Hardware.GPSTrackingFunction.TraccarClientId + "&" + "timestamp=" + date2() + "%20" + time2() + "&" + "lat=" + fmt.Sprintf("%f", GPSLatitude) + "&" + "lon=" + fmt.Sprintf("%f", GPSLongitude) + "&" + "speed=" + fmt.Sprintf("%f", GPSSpeed) + "&" + "course=" + fmt.Sprintf("%f", GPSCourse) + "&" + "variation=" + fmt.Sprintf("%f", GPSVariation))
			} else if Config.Global.Hardware.GPSTrackingFunction.TraccarProto == "opengts" {
				Config.Global.Hardware.GPSTrackingFunction.TraccarServerFullURL = fmt.Sprint(Config.Global.Hardware.GPSTrackingFunction.TraccarServerURL) + ":" + fmt.Sprint(TraccarPortOpenGTS) + "/" + "?" + "id=" + Config.Global.Hardware.GPSTrackingFunction.TraccarClientId + "&" + "gprmc=" + fmt.Sprint(m)

			}
		}
	}

	response, err := http.Get(Config.Global.Hardware.GPSTrackingFunction.TraccarServerFullURL)

	if err != nil {

		log.Println("error: Cannot Establish Connection with Traccar Server! Error ", err)
		currentTime := time.Now()
		if Config.Global.Hardware.TargetBoard == "rpi" {
			if Config.Global.Hardware.GPSTrackingFunction.TrackEnabled {
				if Config.Global.Hardware.GPSTrackingFunction.TrackGPSShowLCD {
					if LCDEnabled {
						LcdText = [4]string{"nil", "TRACK ERR1 " + currentTime.Format("15:04:05"), "nil", "nil"}
						go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
					}
					if OLEDEnabled {
						oledDisplay(false, 4, 1, "TRACK ERR1 "+currentTime.Format("15:04:05"))
					}
				}
			}
		}
		return

	} else {
		contents, err := ioutil.ReadAll(response.Body)
		defer response.Body.Close()

		if err != nil {
			log.Println("error: Error Sending Data to Traccar Server!")
		}

		if response.ContentLength == 0 {
			log.Println("info: Empty Request Response Body")
		} else {
			//
			log.Println("info: Traccar Web Server Response -->\n" + "-------------------------------------------------------------\n" + string(contents) + "-------------------------------------------------------------")
		}
		log.Println("info: HTTP Response Status from Traccar:", response.StatusCode, http.StatusText(response.StatusCode))
		if response.StatusCode >= 200 && response.StatusCode <= 299 {
			log.Println("info: HTTP Status Code from Traccar is in the 2xx range. This is OK.")

			currentTime := time.Now()
			if Config.Global.Hardware.TargetBoard == "rpi" {
				if Config.Global.Hardware.GPSTrackingFunction.TrackEnabled {
					if Config.Global.Hardware.GPSTrackingFunction.TrackGPSShowLCD {
						if LCDEnabled {
							LcdText = [4]string{"nil", "TRACK OK " + currentTime.Format("15:04:05"), "nil", "nil"}
							go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled {
							oledDisplay(false, 4, 1, "TRACK OK "+currentTime.Format("15:04:05"))
						}
					}
				}
			}
		} else {
			currentTime := time.Now()
			if Config.Global.Hardware.TargetBoard == "rpi" {
				if Config.Global.Hardware.GPSTrackingFunction.TrackEnabled {
					if Config.Global.Hardware.GPSTrackingFunction.TrackGPSShowLCD {
						if LCDEnabled {
							LcdText = [4]string{"nil", "TRACK ERR2 " + currentTime.Format("15:04:05"), "nil", "nil"}
							go hd44780.LcdDisplay(LcdText, LCDRSPin, LCDEPin, LCDD4Pin, LCDD5Pin, LCDD6Pin, LCDD7Pin, LCDInterfaceType, LCDI2CAddress)
						}
						if OLEDEnabled {
							oledDisplay(false, 4, 1, "TRACK ERR2 "+currentTime.Format("15:04:05"))
						}
					}
				}
			}
		}
	}
}

func truncateString(str string, num int) string {
	shortstr := str
	if len(str) > num {
		shortstr = str[0:num]
	}
	return shortstr
}

func gpsdatereorder() string {
	GPSDate := "01/12/19" // Works with hard coded date for test. How to make date from GPS visible to helper?
	dd := GPSDate[0:2]
	mm := GPSDate[3:5]
	yy := GPSDate[6:8]
	yyyy := "20" + yy
	Date1reorder := yyyy + "-" + mm + "-" + dd
	return Date1reorder
}

func time2() string {
	currentTime := time.Now()
	Time2 := currentTime.Format("15:04:05")
	return Time2
}

// current date system
func date2() string {
	currentTime := time.Now()
	Date2 := currentTime.Format("2006-01-02")
	return Date2
}
