package main

import (
	"os/exec"
)

func newOpenBrowserCmd(url string) *exec.Cmd {
	return exec.Command("cmd", "/c", "start", url) // windows
}
