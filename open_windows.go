package main

import (
	"context"
	"os/exec"
)

//func init() {
//	newOpenBrowserCmdFn = func(ctx context.Context, url string) *exec.Cmd {
//		return exec.CommandContext(ctx, "cmd", "/c", "start", url) // windows
//	}
//}

func openBrowserCmd(ctx context.Context, url string) *exec.Cmd {
	return exec.CommandContext(ctx, "cmd", "/c", "start", url) // windows
}
