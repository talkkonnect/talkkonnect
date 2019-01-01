package talkkonnect

import (
	"errors"
	"fmt"
	hd44780 "github.com/talkkonnect/go-hd44780"
	"github.com/xackery/gomail"
	"log"
)

func sendviagmail(username string, password string, receiver string, subject string, message string) error {

	err := gomail.Send(username, password, []string{receiver}, subject, message)
	if err != nil {
		return errors.New(fmt.Sprintf("Sending Email Via GMAIL Error: ", err.Error()))
	}

	log.Println("Info: Email Sent Successfully to ", receiver)
	LcdText = [4]string{"nil", "nil", "nil", "Email Sent!"}
	go hd44780.LcdDisplay(LcdText, RSPin, EPin, D4Pin, D5Pin, D6Pin, D7Pin)

	return nil
}
