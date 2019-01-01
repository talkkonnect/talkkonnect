package talkkonnect

import (
	"encoding/hex"
	"fmt"
	"github.com/adrianmo/go-nmea"
	"github.com/jacobsa/go-serial/serial"
	"io"
	"log"
	//	"os"
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
				//os.Exit(-1)
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
					//sentence := "$GPRMC,220516,A,5133.82,N,00042.24,W,173.8,231.8,130694,004.2,W*70"
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
