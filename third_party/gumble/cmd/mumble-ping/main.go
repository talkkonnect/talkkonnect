package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/talkkonnect/gumble/gumble"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags] <destination>\n", os.Args[0])
		flag.PrintDefaults()
	}
	interval := flag.Duration("interval", time.Second*1, "ping packet retransmission interval")
	timeout := flag.Duration("timeout", time.Second*5, "ping timeout until failure")
	jsonOutput := flag.Bool("json", false, "output success response as JSON")
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	server := flag.Arg(0)

	host, port, err := net.SplitHostPort(server)
	if err != nil {
		host = server
		port = strconv.Itoa(gumble.DefaultPort)
	}

	resp, err := gumble.Ping(net.JoinHostPort(host, port), *interval, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], err)
		os.Exit(1)
	}
	major, minor, patch := resp.Version.SemanticVersion()

	if !*jsonOutput {
		fmt.Printf("Address:         %s\n", resp.Address)
		fmt.Printf("Ping:            %s\n", resp.Ping)
		fmt.Printf("Version:         %d.%d.%d\n", major, minor, patch)
		fmt.Printf("Connected Users: %d\n", resp.ConnectedUsers)
		fmt.Printf("Maximum Users:   %d\n", resp.MaximumUsers)
		fmt.Printf("Maximum Bitrate: %d\n", resp.MaximumBitrate)
	} else {
		output := map[string]interface{}{
			"address":         resp.Address.String(),
			"ping":            float64(resp.Ping) / float64(time.Millisecond),
			"version":         fmt.Sprintf("%d.%d.%d", major, minor, patch),
			"connected_users": resp.ConnectedUsers,
			"maximum_users":   resp.MaximumUsers,
			"maximum_bitrate": resp.MaximumBitrate,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.Encode(output)
	}
}
