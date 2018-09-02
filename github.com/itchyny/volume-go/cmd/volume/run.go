package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/itchyny/volume-go"
)

func run(args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("no arg")
	}
	switch args[0] {
	case "-v", "version", "-version", "--version":
		return printVersion(out)
	case "-h", "help", "-help", "--help":
		return printHelp(out)
	case "status":
		if len(args) == 1 {
			return printStatus(out)
		}
	case "get":
		if len(args) == 1 {
			return getVolume(out)
		}
	case "set":
		if len(args) == 2 {
			return setVolume(args[1])
		}
	case "up":
		if len(args) == 1 {
			return upVolume("6")
		} else if len(args) == 2 {
			return upVolume(args[1])
		}
	case "down":
		if len(args) == 1 {
			return downVolume("6")
		} else if len(args) == 2 {
			return downVolume(args[1])
		}
	case "mute":
		if len(args) == 1 {
			return volume.Mute()
		}
	case "unmute":
		if len(args) == 1 {
			return volume.Unmute()
		}
	}
	return fmt.Errorf("invalid argument for volume: %+v", args)
}

func printStatus(out io.Writer) error {
	vol, err := volume.GetVolume()
	if err != nil {
		return err
	}
	muted, err := volume.GetMuted()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "volume: %d\n", vol)
	fmt.Fprintf(out, "muted: %t\n", muted)
	return nil
}

func getVolume(out io.Writer) error {
	vol, err := volume.GetVolume()
	if err != nil {
		return err
	}
	fmt.Fprintln(out, vol)
	return nil
}

func setVolume(volStr string) error {
	vol, err := strconv.Atoi(volStr)
	if err != nil {
		return err
	}
	return volume.SetVolume(vol)
}

func upVolume(diffStr string) error {
	diff, err := strconv.Atoi(diffStr)
	if err != nil {
		return err
	}
	return volume.IncreaseVolume(diff)
}

func downVolume(diffStr string) error {
	diff, err := strconv.Atoi(diffStr)
	if err != nil {
		return err
	}
	return volume.IncreaseVolume(-diff)
}

func printVersion(out io.Writer) error {
	fmt.Fprintf(out, "%s version %s\n", name, version)
	return nil
}

func printHelp(out io.Writer) error {
	fmt.Fprintf(out, strings.Replace(`NAME:
   $NAME - %s

USAGE:
   $NAME command [argument...]

COMMANDS:
   status      prints the volume status
   get         prints the current volume
   set [vol]   sets the audio volume
   up [diff]   volume up by [diff]
   down [diff] volume down by [diff]
   mute        mutes the audio
   unmute      unmutes the audio
   version     prints the version
   help        prints this help

VERSION:
   %s

AUTHOR:
   %s
`, "$NAME", name, -1), description, version, author)
	return nil
}
