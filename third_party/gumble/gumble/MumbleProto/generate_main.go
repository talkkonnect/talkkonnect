// +build ignore

package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	// Get Mumble.proto
	resp, err := http.Get("https://raw.githubusercontent.com/mumble-voip/mumble/master/src/Mumble.proto")
	if err != nil {
		log.Fatalf("could not download Mumble.proto: %s\n", err)
	}
	defer resp.Body.Close()

	// Write Mumble.proto
	f, err := os.Create("Mumble.proto")
	if err != nil {
		log.Fatalf("could not create Mumble.proto: %s\n", err)
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		log.Fatalf("could not write Mumble.proto: %s\n", err)
	}
	if err := f.Close(); err != nil {
		log.Fatalf("could not close Mumble.proto: %s\n", err)
	}

	// Build proto-gen-go
	if err := exec.Command("go", "build", "-o", "protoc-gen-go", "github.com/golang/protobuf/protoc-gen-go").Run(); err != nil {
		log.Fatalf("could not build protoc-gen-go: %s\n", err)
	}

	// Generate code
	if err := exec.Command("protoc", "--go_out=.", "Mumble.proto").Run(); err != nil {
		log.Fatalf("could not run protoc: %s\n", err)
	}

	// Clean up
	os.Remove("Mumble.proto")
	os.Remove("protoc-gen-go")
}
