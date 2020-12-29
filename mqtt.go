package talkkonnect

import (
	"crypto/tls"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/talkkonnect/gpio"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//Variables used by Program
var (
	Iotuuid string
	relay1          gpio.Pin
	relay2          gpio.Pin
	relay3          gpio.Pin
	relay4          gpio.Pin
	relay5          gpio.Pin
	relay6          gpio.Pin
	relay7          gpio.Pin
	relay8          gpio.Pin
	relay1State     bool = false
	relay2State     bool = false
	relay3State     bool = false
	relay4State     bool = false
	relay5State     bool = false
	relay6State     bool = false
	relay7State     bool = false
	relay8State     bool = false
	relayAllState   bool = false
	RelayPulseMills time.Duration
	TotalRelays     uint
	RelayPins       = [9]uint{}
	Topic           string
	Broker          string
	MQTTPassword    string
	User            string
	Id              string
	Cleansess       bool
	Qos             int
	Num             int
	Payload         string
	Action          string
	Store           string
)

type dateTimeScheduleStruct struct {
	startDateTime string
	endDateTime   string
	matched       bool
	defaultLogic  bool
	stopOnMatch   bool
}

type dayScheduleStruct struct {
	dayint       int
	startTime    int
	endTime      int
	matched      bool
	defaultLogic bool
	stopOnMatch  bool
}

func relayAllPulse() {
	if relayAllState {
		relayCommand(0, "off")
	} else {
		relayCommand(0, "on")
	}
	relayAllState = !relayAllState
}

func relayCommand(relayNo int, command string) {
	// all relays (0)
	if relayNo == 0 {
		for i := 1; i <= int(TotalRelays); i++ {
			if command == "on" {
				log.Println("info: Relay ", i, "On")
				gpio.NewOutput(RelayPins[i], false)

			}
			if command == "off" {
				log.Println("info: Relay ", i, "Off")
				gpio.NewOutput(RelayPins[i], true)
			}
			if command == "pulse" {
				log.Println("info: Relay ", i, "Pulse")
				gpio.NewOutput(RelayPins[i], false)
				time.Sleep(RelayPulseMills * time.Millisecond)
				gpio.NewOutput(RelayPins[i], true)
			}
		}
		return
	}

	//specific relay (Number Between 1 and TotalRelays)
	if relayNo >= 0 && relayNo <= int(TotalRelays) {
		if command == "on" {
			log.Println("info: Relay ", relayNo, "On")
			gpio.NewOutput(RelayPins[relayNo], false)
		}
		if command == "off" {
			log.Println("info: Relay ", relayNo, "Off")
			gpio.NewOutput(RelayPins[relayNo], true)
		}
		if command == "pulse" {
			log.Println("info: Relay ", relayNo, "Pulse")
			gpio.NewOutput(RelayPins[relayNo], false)
			time.Sleep(RelayPulseMills * time.Millisecond)
			gpio.NewOutput(RelayPins[relayNo], true)
		}
	}
}

func dateTimeWithinRange(dateTimeSchedule dateTimeScheduleStruct) (bool, bool, bool, error) {
	var dateFormat string = "02/01/2006 15:04"
	startDateTime, err := time.Parse(dateFormat, dateTimeSchedule.startDateTime)
	if err != nil {
		return false, false, false, err
	}

	endDateTime, err := time.Parse(dateFormat, dateTimeSchedule.endDateTime)
	if err != nil {
		return false, false, false, err
	}

	checkDateTime, err := time.Parse(dateFormat, time.Now().Format("02/01/2006 15:04"))
	if err != nil {
		return false, false, false, err
	}
	log.Println("------")
	log.Println("debug: startdate is ", startDateTime, " enddate is ", endDateTime, " check date is ", checkDateTime)
	if startDateTime.Before(checkDateTime) && endDateTime.After(checkDateTime) {
		return true, dateTimeSchedule.defaultLogic, dateTimeSchedule.stopOnMatch, nil
	}
	return false, dateTimeSchedule.defaultLogic, dateTimeSchedule.stopOnMatch, nil
}

//func dayTimeWithinRange(startTime string, endTime string, dayCheck string, dateFormat string, defaultLogicDay string) (bool, error) {
func dayTimeWithinRange(dayTimeWithinRange dayScheduleStruct) (bool, bool, bool, error) {

	t1 := time.Now()
	t1Day := int(t1.Weekday())
	t1Minute := int((t1.Hour() * 60) + t1.Minute())

	log.Println("------")
	log.Println("debug: day is ", t1Day, " starttime is ", dayTimeWithinRange.startTime, " endtime is ", dayTimeWithinRange.endTime, " checkday is ", t1Day, " check time is ", t1Minute)

	if t1Day == dayTimeWithinRange.dayint && (dayTimeWithinRange.startTime <= t1Minute && dayTimeWithinRange.endTime >= t1Minute) {
		return true, dayTimeWithinRange.defaultLogic, dayTimeWithinRange.stopOnMatch, nil
	}
	return false, dayTimeWithinRange.defaultLogic, dayTimeWithinRange.stopOnMatch, nil
}

func mqtttestpub() {

	if Action != "pub" {
		log.Println("error: Invalid setting for -action, must be pub for test pub")
		return
	}

	if Topic == "" {
		log.Println("error: Invalid setting for -topic, must not be empty")
		return
	}

	opts := MQTT.NewClientOptions()
	opts.AddBroker(Broker)
	opts.SetClientID(Id)
	opts.SetUsername(User)
	opts.SetPassword(MQTTPassword)
	opts.SetCleanSession(Cleansess)

	if Store != ":memory:" {
		opts.SetStore(MQTT.NewFileStore(Store))
	}

	if Action == "pub" {

		log.Printf("info: action      : %s\n", Action)
		log.Printf("info: broker      : %s\n", Broker)
		log.Printf("info: clientid    : %s\n", Id)
		log.Printf("info: user        : %s\n", User)
		log.Printf("info: mqttpassword: %s\n", MQTTPassword)
		log.Printf("info: topic       : %s\n", Topic)
		log.Printf("info: message     : %s\n", Payload)
		log.Printf("info: qos         : %d\n", Qos)
		log.Printf("info: cleansess   : %v\n", Cleansess)
		log.Printf("info: num         : %d\n", Num)
		log.Printf("info: store       : %s\n", Store)

		client := MQTT.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}

		log.Println("info: Test MQTT Publisher Started")
		for i := 0; i < Num; i++ {
			log.Println("info: Publishing MQTT Message")
			token := client.Publish(Topic, byte(Qos), false, Payload)
			token.Wait()
		}

		client.Disconnect(250)

		log.Println("info: Test MQTT Publisher Disconnected")
	}
}

func mqttsubscribe() {

	log.Printf("info: MQTT Subscription Information")
	log.Printf("info: MQTT Broker      : %s\n", Broker)
	log.Printf("info: MQTT clientid    : %s\n", Id)
	log.Printf("info: MQTT user        : %s\n", User)
	log.Printf("info: MQTT password    : %s\n", MQTTPassword)
	log.Printf("info: Subscribed topic : %s\n", Topic)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	connOpts := MQTT.NewClientOptions().AddBroker(Broker).SetClientID(Id).SetCleanSession(true)
	if User != "" {
		connOpts.SetUsername(User)
		if MQTTPassword != "" {
			connOpts.SetPassword(MQTTPassword)
		}
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: true, ClientAuth: tls.NoClientCert}
	connOpts.SetTLSConfig(tlsConfig)

	connOpts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(Topic, byte(Qos), onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := MQTT.NewClient(connOpts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Printf("info: Connected to     : %s\n", Broker)
	}

	<-c
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	log.Printf("info: Received MQTT message on topic: %s Payload: %s\n", message.Topic(), message.Payload())

	if string(message.Payload()) == "relay1:on" {
		relayCommand(1, "on")
		return
	}

	if string(message.Payload()) == "relay1:off" {
		relayCommand(1, "off")
		return
	}

	if string(message.Payload()) == "relay1:pulse" {
		relayCommand(1, "pulse")
		return
	}

	log.Printf("error: Undefined Command Received in MQTT message : %s \n", message.Payload())
	return

}
