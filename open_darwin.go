package main

import (
	"os/exec"
)

func newOpenBrowserCmd(url string) *exec.Cmd {
	return exec.Command("open", url) // macOS
}
