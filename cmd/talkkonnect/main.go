package main

import (
	"flag"
	"github.com/talkkonnect/talkkonnect"
	"fmt"
	"log"
)

func main() {

	// Command line flags

	config := flag.String("config", "/home/mumble/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml", "full path to talkkonnect.xml configuration file")

	flag.Usage = talkkonnectusage

	flag.Parse()

	log.Println("info: Reading Config File: ", *config)
	talkkonnect.PreInit(*config)
}

func talkkonnectusage() {
	fmt.Println("---------------------------------------------------------------------------------------")
	fmt.Println("Usage: talkkonnect [-config=[full path and file to talkkonnect.xml configuration file]]")
	fmt.Println("By Suvir Kumar <suvir@talkkonnect.com>")
	fmt.Println("For more information visit http://www.talkkonnect.com and github.com/talkkonnect")
	fmt.Println("---------------------------------------------------------------------------------------")
	fmt.Println("-config=/home/mumble/gocode/src/github.com/talkkonnect/talkkonnect/talkkonnect.xml")
	fmt.Println("-version for the version")
	fmt.Println("-help for this screen")
}
