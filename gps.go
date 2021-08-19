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
	//"encoding/json" // used for testing http api for flespi
)

var (
	TraccarPortOsmAnd  string = "5055" // Traccar Client port 5055 for working with OsmAnd Protocol
	TraccarPortT55     string = "5005" // Old Traccar Client port 5005 for working with T55 Protocol
	TraccarPortOpenGTS string = "5159" // Traccar Client port 5159 for for working OpenGTS Protocol
)

//A test http request for OsmAnd protocol and passing basic $GPRMC information to Traccar.
//http://10.8.0.1:5055?id=12345&timestamp=2019-10-05%20&lat=44.000000&lon=20.000000&speed=0.000&course=0.000000&variation=0.000000
//More info about OsmAnd for Traccar: https://www.traccar.org/osmand

//A request example for OpenGTS protocol in Traccar...
//http://10.8.0.1:5159/?id=12345&gprmc=$GPRMC,094852,A,4446.8347,N,02030.9393,E,0.0147,69.114,041019,,*10

// Supported protocols for communicating with Traccar: OsmAnd, T55 and OpenGTS.
// OsmAnd and OpenGTS use http. T55 use tcp socket connection (udp is also possible).

var goodGPSRead bool = false

// GPS Serial reading.

func getGpsPosition(verbose bool) (bool, error) {

	if GpsEnabled {

		if Port == "" {
			return false, errors.New("you must specify port")
		}

		if Even && Odd {
			return false, errors.New("cant specify both even and odd parity")
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
			return false, errors.New("cannot open serial port")
		} else {
			defer f.Close()
		}

		if TxData != "" {
			txData, err := hex.DecodeString(TxData)

			if err != nil {
				GpsEnabled = false
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

		if Rx {
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

					/*if s.DataType() == nmea.TypeGGA {
											g := s.(nmea.RMC)
					                                                if g.Latitude != 0 && g.Longitude != 0 {
												goodGPSRead = true
												fmt.Println(...)
					*/
					// Try to read other sentences containing useful info like $GPGGA, ...
					// and  print useful info, altitude, number of satellites, fix quality, etc. To Do.

					if s.DataType() == nmea.TypeRMC {
						m := s.(nmea.RMC)
						if m.Latitude != 0 && m.Longitude != 0 {
							goodGPSRead = true

							//if m.Speed != 0 && m.Course != 0 {
							//isVehicleMoving = true  // Report if vehicle moves or overspeeds?
							//if m.Speed >> some km/h
							// Send a text alert to Mumble or email?

							// Read GGA Info for FixQuality, NumSatellites, Altitude ... To Do..
							// https://github.com/adrianmo/go-nmea/blob/master/gga.go
							//...

							/*if TrackEnabled == true {
							  if TraccarSendTo == true {
							  if TraccarProto == "osmand" {
							  TraccarServerFullURL = (fmt.Sprint(TraccarServerURL) + ":" + fmt.Sprint(TraccarPortOsmAnd) + "?" + "id=" + TraccarClientId + "&" + "timestamp=" + date2() + "%20" + time2() + "&" + "lat=" + fmt.Sprintf("%f", m.Latitude)+ "&" + "lon=" + fmt.Sprintf("%f", m.Longitude) + "&" + "speed=" + fmt.Sprintf("%f", m.Speed) + "&" + "course=" + fmt.Sprintf("%f", m.Course) + "&" + "variation=" + fmt.Sprintf("%f", m.Variation))
							  } else if TraccarProto == "opengts" {
							  TraccarServerFullURL = fmt.Sprint(TraccarServerURL) + ":" + fmt.Sprint(TraccarPortOpenGTS) + "/?" + "id=" + TraccarClientId + "&" + "gprmc=" + fmt.Sprint(m)
							  //TraccarServerFullURL = fmt.Sprint(TraccarServerURL) + ":" + fmt.Sprint(TraccarPortOpenGTS) + "?" + "id=" + TraccarClientId + "&" + "timestamp=" + Date2 + "%20" + Time2 + "&" + "lat=" + fmt.Sprintf("%f", m.Latitude) + "&" + "lon=" + fmt.Sprintf("%f", m.Longitude) + "&" + "speed=" + fmt.Sprint(m.Speed) + "&" + "course=" + fmt.Sprintf("%f", m.Course) + "&" + "variation=" + fmt.Sprintf("%f", m.Variation)
							  } else if TraccarProto == "t55" {
							  TraccarServerFullURL = ""
							  }
							  }
							  }
							  }
							*/

							// This is a request example for OpenGTS protocol in Traccar. Use port 5159.
							// http://10.8.0.1:5159/?id=12345&gprmc=$GPRMC,094852,A,4446.8347,N,02030.9393,E,0.0147,69.114,041019,,*10
							// http://10.8.0.1:5159/?id=12345&gprmc=$GPRMC,114614.00,A,4446.82735,N,02030.94387,E,0.030,,151019,,,A*70

							FreqReport := float64(TraccarReportFrequency)                        // Reporting Frequency
							FreqReports := (time.Duration(TraccarReportFrequency) * time.Second) // Frequency of GPS Reporting. Minutes, Seconds or hours?

							//FreqReportm := (FreqReports / 60)  // in minutes
							//FreqReport := 60 // Reporting Frequency
							//FreqReports := (time.Duration(FreqReport) * time.Second) // Frequency of GPS Reporting. Minutes, Seconds or hours?
							//FreqReportm := (FreqReport / 60)  // minutes
							//FreqReportH := (FreqReports/(60*60)) // hours
							//FreqReportD := (FreqReports/(60*60*24)) // days
							//Improve calc with time/ string to print reporting freq in minutes or seconds, as needed.

							// Position Reporter

							// send GPS position once immediately on start
							if TrackEnabled && TraccarSendTo {
								if TraccarProto == "t55" {
									go tcpSendT55Traccar2() // Initial Send GPS position to Traccar with old T55 client protocol. No keep-alive.
									//tcpSendT55Traccar1()  // Initial Send GPS position to Traccar with old T55 client protocol. Keep-alive.
								} else {
									go httpSendTraccar() // Initial Send GPS position to Traccar over http function for both OsmAnd or OpenGTS protocol.
									//flespi() // Test flespi
								}
								log.Println("info: GPS Position Report Nr (1) Sent to Traccar Server")
							}

							// Now, keep sending GPS position and counting how many reports were sent?

							PositionReporter := time.NewTicker(FreqReports)
							var TraccarCounter = 1
							go func() {
								for range PositionReporter.C {
									if TrackEnabled && TraccarSendTo {
										if TraccarProto == "t55" {
											tcpSendT55Traccar2() // Send GPS position to Traccar with old T55 client protocol. No keep-alive.
											//tcpSendT55Traccar1()  // Send GPS position to Traccar with old T55 client protocol. Keep-alive.
										} else {
											httpSendTraccar() // Send GPS position to Traccar over http function for both OsmAnd or OpenGTS protocol.
											//flespi() // Test flespi
										}
									}
									TraccarCounter++

									if verbose {
										if TrackEnabled && TraccarSendTo {
											if TraccarProto == "osmand" {
												log.Println("info: OsmAnd: ", TraccarServerURL+":"+fmt.Sprint(TraccarPortOsmAnd)+"?"+"id="+TraccarClientId+"&"+"timestamp="+date2()+"%20"+time2()+"&"+"lat="+fmt.Sprint(m.Latitude)+"&"+"lon="+fmt.Sprint(m.Longitude)+"&"+"speed="+fmt.Sprint(m.Speed)+"&"+"course="+fmt.Sprint(m.Course)+"&"+"variation="+fmt.Sprint(m.Variation))
											} else if TraccarProto == "t55" {
												log.Println("info: T55: " + "Sending " + fmt.Sprint(m) + " to " + TraccarServerURL + ":" + fmt.Sprint(TraccarPortT55))
											} else if TraccarProto == "opengts" {
												log.Println("info: OpenGTS: ", TraccarServerURL+":"+fmt.Sprint(TraccarPortOpenGTS)+"/?"+"id="+TraccarClientId+"&"+fmt.Sprint(m))
											}
											log.Println("info: GPS Position Report Nr " + "(" + fmt.Sprint(TraccarCounter) + ")" + " Sent to Traccar Server. Next Position Report in " + fmt.Sprintf("%.2f", FreqReport/60) + " minute(s)")
										}
									}

									//Display Show GPS Position.

									if TargetBoard == "rpi" {
										if TrackEnabled {
											if TrackGPSShowLCD {
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

							Date1 := fmt.Sprint(gpsdatereorder())                  // Reformatted date for Tracar
							Time1 := fmt.Sprintf("%s", truncateString(GPSTime, 8)) // Truncate time for Traccar

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
								log.Println("info: Traccar Cmd Osmand: " + fmt.Sprint(TraccarServerFullURL))
								log.Println("info: Traccar $GPRMC Sentence for T55/OpenGTS: " + fmt.Sprint(m))

								if TrackEnabled {
									//log.Println("info: GPS Tracking Enabled: " + fmt.Sprint(TrackEnabled))
									if TraccarSendTo {
										log.Println("info: Sending GPS Position to Traccar Server Enabled")
										log.Println("info: Traccar Protocol: " + strings.Title(strings.ToLower(TraccarProto)) + "; " + "Reporting Frequency: " + fmt.Sprintf("%.2f", FreqReport/60) + " minutes;")
										//Print GPS message format for sending to Traccar depending on tracking protocol.

										switch TraccarProto {
										case "osmand":
											log.Println("info: OsmAnd: ", TraccarServerFullURL)
										case "opengts":
											log.Println("info: OpenGTS: ", TraccarServerIP)
										case "t55":
											log.Println("info: T55:", fmt.Sprint(m), "...", TraccarServerURL+":"+TraccarPortT55)
										default:
											log.Println("info: OsmAnd: ", TraccarServerFullURL)
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
			return false, errors.New("Rx Not Set")
		}
		return goodGPSRead, nil
	}
	return false, errors.New("GPS Not Enabled. Or Not Connected")
}

/*
GPS Reporter function. Need more work.
func gpsReporter() {
//ticker := time.NewTicker(60 * time.Second) // send every 60 seconds.
ticker := time.NewTicker(time.Duration(int64(TraccarReportFrequency)) * time.Second) // send every ... seconds.
done := make(chan bool)
go func() {
for {
select {
case <-done:
return
case t := <-ticker.C:
//httpSendTraccar()
//tcpSendT55Traccar1()
//tcpSendT55Traccar2()
log.Println("info: GPS Position Sent to Traccar Server over" + strings.Title(strings.ToLower(TraccarProto)) + "protocol", t)
//log.Println("info: OsmAnd: ", fmt.Sprintf(TraccarServerURL + ":" + TraccarPortOsmAnd + "?" + "id=" + TraccarClientId) + "&" + "timestamp=" + Date2 + "%" + "20" + Time2 + "&" + "lat=" + fmt.Sprintf("%f", m.Latitude) + "&" + "lon=" + fmt.Sprintf("%f", m.Longitude) + "&" + "speed=" + fmt.Sprint(m.Speed) + "&" + "course=" + fmt.Sprintf("%f", m.Course) + "&" + "variation=" + fmt.Sprintf("%f", m.Variation))
//log.Println("info: GPS Position Sent to Traccar Server. Next Position Report in " + fmt.Sprintf(strconv.Itoa(FreqReportm)) + " minute(s)")
}
}
}()
}

*/

/* T55 Protocol.
Info: https://www.traccar.org/traccar-client-protocol
T55 protocol is used by old version of Traccar Client.
New version use OsmAnd protocol. T55 Protocol uses TCP/IP as a transport layer.
It is more simple. Can be UDP also, but it is less reliable. Messages are separated
simply by carriage return and line feed characters (\r\n).

How to format T55 message for sending to Traccar?

Login (send once the TCP connection is establshed):
$PGID,12345*0F\r\n (where 12345 - IMEI or other unique id)

Simple location report format is just a standard NMEA GPRMC sentence:
$GPRMC,225446,A,4916.45,N,12311.12,W,000.5,054.7,191194,020.3,E*68\r\n

225446 - Time of fix 22:54:46 UTC
A - Navigation receiver warning A = OK, V = warning
4916.45,N - Latitude 49 deg. 16.45 min North
12311.12,W - Longitude 123 deg. 11.12 min West
000.5 - Speed over ground, Knots
054.7 - Course Made Good, True
191194 - Date of fix 19 November 1994
020.3,E - Magnetic variation 20.3 deg East
*68 - Checksum

Extended location report format:

$TRCCR,20140111000000.000,A,60.000000,60.000000,0.00,0.00,0.00,50,*3a\r\n

20140111000000.000 - Date and tim of fix 2014-01-11 00:00:00.000 UTC
A - Navigation receiver warning A = OK, V = warning
60.000000 - Latitude in degrees 60 deg (negative for south hemisphere)
60.000000 - Longitude in degrees 60 deg (negative for west hemisphere)
0.00 - Speed over ground, Knots
0.00 - Course Made Good, True
0.00 - Altitude in meters
50 - Battery level
*/

//T55(1)

//T55(2)
//Another T55 TCP Connection with keep-alive. EOF for indicating a connection drop.
//Keep-alive. Client stays connected between sending position.

func tcpSendT55Traccar2() {

	//Hard coded for testing
	//pgid := "$PGID,12345*0F\r\n"  // Unique Client ID (e.g. 12345). Follow with carriage return and line feed characters (\r\n).
	//gprmc := "$GPRMC,114614.00,A,4446.82735,N,02030.94387,E,0.030,,151019,,,A*70\r\n" //  Test sentence. Send after pgid. Follow \r\n.

	pgid := "$PGID" + "," + TraccarClientId + "*0F" + "\r" + "\n" // Unique Client ID (e.g. 12345). Follow with carriage return and line feed $
	gprmc := fmt.Sprint(m) + "\r" + "\n"
	log.Println("info: $GPRMC to send is: " + fmt.Sprint(m))
	fmt.Println(m)

	conn, _ := net.Dial("tcp", TraccarServerIP+":"+fmt.Sprint(TraccarPortT55)) // Use port 5005 for T55. Keep-alive.
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

	//log.Println("info: Sending T55 Position Report to Traccar ")
	//log.Println("info: Connection established with Traccar Server :", TraccarServerIP)
	//log.Println("info: Traccar Server Address :", conn.RemoteAddr().String())
	//log.Println("info: Traccar Client Address :", conn.LocalAddr().String())

	log.Println("info: Traccar Client:", conn.LocalAddr().String(), "Connected to Server:", conn.RemoteAddr().String())

	// Send a T55 position report...

	fmt.Fprintf(conn, pgid) // Send ID
	time.Sleep(1 * time.Second)
	fmt.Fprintf(conn, gprmc) // send $GPRMC
	log.Println("info: Sending position message to Traccar over Protocol: " + strings.Title(strings.ToLower(TraccarProto)))

	notify := make(chan error)

	// error checking for a closed connection. dispatch a goroutine to read from the connection until there's an EOF error.

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
				//fmt.Println("Unexpected Data: %s", buf[:n])
			}
		}
	}()

	for {
		select {
		case err := <-notify:
			log.Println("info: Traccar Server Connection dropped message", err)
			//fmt.Println("Traccar Server Connection dropped message", err)

			if err == io.EOF {
				log.Println("Connection to Traccar Server was closed")
				//fmt.Println("Connection to Traccar Server was closed")
				return
			}
		case <-time.After(time.Second * 60):
			log.Println("Traccar Server Connection Timeout 60. Still Alive")
			//fmt.Println("Traccar Server Connection Timeout 60. Still Alive")
		}
	}
}

// T55(2) End

// Sending over OsmAnd and OpenGTS.
// OsmAnd is primary Traccar protocol for sending position over http api.

// OpenGTS currently not working. Need to check if format parsing is correct?
// Traccar returns 400 http error.
// A request example for OpenGTS protocol in Traccar is...
// http://10.8.0.1:5159/?id=12345&gprmc=$GPRMC,094852,A,4446.8347,N,02030.9393,E,0.0147,69.114,041019,,*10
// This is correct format for OpenGTS. Works directly from a web browser.
// For troubleshooting this problem check /opt/traccar/logs/tracker-server.log in the Traccar server
// Message will be HEX coded. Decode with calculator
// OsmAnd is working OK.

func httpSendTraccar() {

	//Date1 := fmt.Sprint(gpsdatereorder())                 // Reformatted date for Tracar
	//Time1 := fmt.Sprint("%s", truncateString(GPSTime, 8)) // Truncate time for Traccar

	if TrackEnabled {
		if TraccarSendTo {
			if TraccarProto == "osmand" {
				TraccarServerFullURL = (fmt.Sprint(TraccarServerURL) + ":" + fmt.Sprint(TraccarPortOsmAnd) + "?" + "id=" + TraccarClientId + "&" + "timestamp=" + date2() + "%20" + time2() + "&" + "lat=" + fmt.Sprintf("%f", GPSLatitude) + "&" + "lon=" + fmt.Sprintf("%f", GPSLongitude) + "&" + "speed=" + fmt.Sprintf("%f", GPSSpeed) + "&" + "course=" + fmt.Sprintf("%f", GPSCourse) + "&" + "variation=" + fmt.Sprintf("%f", GPSVariation))
			} else if TraccarProto == "opengts" {
				TraccarServerFullURL = fmt.Sprint(TraccarServerURL) + ":" + fmt.Sprint(TraccarPortOpenGTS) + "/" + "?" + "id=" + TraccarClientId + "&" + "gprmc=" + fmt.Sprint(m)

				//gprmctest = "$GPRMC,081753.00,A,4446.82690,N,02030.95217,E,0.285,,011219,,,A*73"
				//TraccarServerFullURL := `http://10.8.0.1:5055?id=tk-demo-04&timestamp=2019-10-05%2011:30:00&lat=44.000000&lon=20.000000&speed=0.000&course=0.000000&variation=0.000000`
				//TraccarServerFullURL := (fmt.Sprint(TraccarServerURL) + ":" + fmt.Sprint(TraccarPortOsmAnd) + "?" + "id=" + TraccarClientId + "&" + "timestamp=" + fmt.Sprint(gpsdatereorder()) + "%20" + time2() + "&" + "lat=" + fmt.Sprint(GPSLatitude) + "&" + "lon=" + fmt.Sprint(GPSLongitude) + "&" + "speed=" + fmt.Sprint(GPSSpeed) + "&" + "course=" + fmt.Sprint(GPSCourse) + "&" + "variation=" + fmt.Sprint(GPSVariation))
				//http://10.8.0.1:5159/?id=12345&gprmc=$GPRMC,094852,A,4446.8347,N,02030.9393,E,0.0147,69.114,041019,,*10 //URL for testing openGTS

			}
		}
	} //

	response, err := http.Get(TraccarServerFullURL)

	if err != nil {

		log.Println("error: Cannot Establish Connection with Traccar Server! Error ", err)
		// Print to LCD "TRACK ERR 1". Error 1 for network connectivity problems with Traccar server. Check why the connectivity failed?
		currentTime := time.Now()
		if TargetBoard == "rpi" {
			if TrackEnabled {
				if TrackGPSShowLCD {
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
		// LCD TRACK ERR1 to screen
		return

	} else {
		contents, err := ioutil.ReadAll(response.Body)
		// if it's no error then defer the call for closing body
		defer response.Body.Close()

		if err != nil {
			log.Println("error: Error Sending Data to Traccar Server!")
		}

		// if http request response body is empty
		if response.ContentLength == 0 {
			log.Println("info: Empty Request Response Body")
		} else {
			//
			log.Println("info: Traccar Web Server Response -->\n" + "-------------------------------------------------------------\n" + string(contents) + "-------------------------------------------------------------")
		}
		// Added code to read http response status code, 2xx, 4xx.
		log.Println("info: HTTP Response Status from Traccar:", response.StatusCode, http.StatusText(response.StatusCode))
		if response.StatusCode >= 200 && response.StatusCode <= 299 {
			log.Println("info: HTTP Status Code from Traccar is in the 2xx range. This is OK.")

			// "TRACK OK" Message To LCD.
			// Print LCD message "TRACK OK" to LCD, overwrite row with "GPS OK" + timestamp to report a
			// position had been successfully transferred to Traccar (after 200 OK).
			// time.Sleep(1 * time.Second)
			currentTime := time.Now()
			if TargetBoard == "rpi" {
				if TrackEnabled {
					if TrackGPSShowLCD {
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
			// "TRACK OK" Message.
		} else {
			// "TRACK ERR 2" Message to LCD. Error 2 means connectivity is working, but something else went wrong (Response Status 400), like data is corrupted. HEX decode Traccar Server log message. Keep trying.
			currentTime := time.Now()
			if TargetBoard == "rpi" {
				if TrackEnabled {
					if TrackGPSShowLCD {
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
		// "TRACK ERR2" Message.
	} //
	// Note: When sending over Traccar client ports 5055, 5159 response code should be 200 OK. No "body text" is visible.
	// If using Trccar web port 8082 to test, code 200 OK and "body text" from Traccar server will show.
}

// OsmAnd End

// Some helper functons to use with GPS

// Truncate function. Need to truncate .0000 / milliseconds time from GPSTime. Is there a more simple way?
// example: Time := fmt.Sprint(truncateString(GPSTime,8))  -> show only the first 8 characters, ending with seconds.
// Should gps time be different type? int64?

func truncateString(str string, num int) string {
	shortstr := str
	if len(str) > num {
		shortstr = str[0:num]
	}
	return shortstr
}

// Or trim string like this fmt... (string[:num])... num from which character.

// helper to reorder gps date from dd/mm/yy to yyyy/mm/dd format

func gpsdatereorder() string {
	GPSDate := "01/12/19" // Works with hard coded date for test. How to make date from GPS visible to helper?
	//GPSDate := fmt.Sprintf("%s", m.Date)
	dd := GPSDate[0:2]
	//fmt.Println(yy) //test
	//runtime error: slice bounds out of range! But works in Go Playground. Need to fix this error.
	mm := GPSDate[3:5]
	//fmt.Println(mm) // test
	yy := GPSDate[6:8]
	//fmt.Println(dd) //test
	yyyy := "20" + yy
	//fmt.Println(yyyy) //test
	Date1reorder := yyyy + "-" + mm + "-" + dd
	//fmt.Println(Date1reorder)
	return Date1reorder
}

// "errchk" checks if an error is nil and if it isn't calls log.Fatalln(err)

func errchk(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// current time system stamp
func time2() string {
	currentTime := time.Now()
	Time2 := fmt.Sprintf("%s", currentTime.Format("15:04:05"))
	return Time2
}

// current date system
func date2() string {
	currentTime := time.Now()
	Date2 := fmt.Sprintf("%s", currentTime.Format("2006-01-02"))
	return Date2
}
