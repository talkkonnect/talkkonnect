package main

import (
	"crypto/rand"
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/talkkonnect/gumble/gumble"
	_ "github.com/talkkonnect/gumble/opus"
	"github.com/talkkonnect/talkkonnect"
	"os"
	"os/signal"
	"syscall"
	"log"
)

func talkkonnectusage() {
	fmt.Println("Usage: talkkonnect -server= -username= -password=")
	fmt.Println("talkkonnect is a headless mumble client with lcd and channel up down buttons made for raspberry pi 3 b+")
	fmt.Println("By Suvir Kumar <suvir@talkkonnect.com>")
	fmt.Println("For more information visit http://www.talkkonnect.com")
	fmt.Println("- server")
	fmt.Println("	string is the server to connect to (default \"218.219.148.36:8000\")")
	fmt.Println("- username")
	fmt.Println("        the username of the client (If not provided talkkonnect will generate a random username prefixed by talkkonnect-xx:xx:xx:xx:xx")
	fmt.Println("- password")
	fmt.Println("        the password of the server")
	fmt.Println("-insecure")
	fmt.Println("        skip server certificate verification (default true)")
	fmt.Println("-certificate string")
	fmt.Println("        PEM encoded certificate and private key")
	fmt.Println("-channel string")
	fmt.Println("        mumble channel to join by default (default \"Root\")")
	fmt.Println("-logging string")
	fmt.Println("        Select Logging to screen or file or both (default \"screen\")")
	fmt.Println("-daemonize")
	fmt.Println("        Select daemonize as no to run in foreground or yes to run in background as daemon (default \"no\")")
}

func main() {

	// Command line flags
	server      := flag.String("server", "218.219.148.36:8000", "the server to connect to")
	username    := flag.String("username", "", "the username of the client")
	password    := flag.String("password", "", "the password of the server")
	insecure    := flag.Bool("insecure", true, "skip server certificate verification")
	certificate := flag.String("certificate", "", "PEM encoded certificate and private key")
	channel     := flag.String("channel", "Root", "mumble channel to join by default")
	logging     := flag.String("logging", "screen", "Select Logging to screen or file or both")
	daemonize   := flag.String("daemonize", "no", "Select daemonize as no to run in foreground or yes to run in background as daemon")

	flag.Usage = talkkonnectusage

	flag.Parse()

	// Initialize
	b := talkkonnect.Talkkonnect{
		Config:      gumble.NewConfig(),
		Address:     *server,
		ChannelName: *channel,
		Logging:     *logging,
		Daemonize:   *daemonize,
	}

	// if no username specified, lets just autogen a random one
	if len(*username) == 0 {
		buf := make([]byte, 6)
		_, err := rand.Read(buf)
		if err != nil {
			log.Println("error: ", err)
			os.Exit(1)
		}

		buf[0] |= 2
		b.Config.Username = fmt.Sprintf("talkkonnect-%02x%02x%02x%02x%02x%02x", buf[0], buf[1], buf[2], buf[3], buf[4], buf[5])
	} else {
		b.Config.Username = *username
	}

	b.Config.Password = *password

	if *insecure {
		b.TLSConfig.InsecureSkipVerify = true
	}
	if *certificate != "" {
		cert, err := tls.LoadX509KeyPair(*certificate, *certificate)
		if err != nil {
			log.Println("error: ", err)
			os.Exit(1)
		}
		b.TLSConfig.Certificates = append(b.TLSConfig.Certificates, cert)
	}

	b.Init()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	exitStatus := 0

	<-sigs
	b.CleanUp()

	os.Exit(exitStatus)
}
