package main

import (
	"fmt"
	"log"
	"os/exec"
)

// openBrowser opens a browser for url.
func openBrowser(url string) error {
	log.Println("attempting to open a browser for ", url)
	cmd := exec.Command("cmd", "/c", "start", url) // windows
	err := cmd.Run()
	if err != nil {
		log.Printf("failed to open browser for %s: %v", url, err)
		return err
	}

	if didStartServer() {
		fmt.Printf("Opening %s on GOPATH %s\n", url, gopath)
	} else {
		fmt.Printf("Opening %s on already-existing server\n", url)
	}

	return nil
}
