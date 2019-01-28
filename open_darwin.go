package main

import (
	"context"
	"os/exec"
)

func init() {
	newOpenBrowserCmdFn = func(ctx context.Context, url string) *exec.Cmd {
		return exec.CommandContext(ctx, "open", url) // macOS
	}
}

//newOp(ctx context.Context, url string) *exec.Cmd {
//
//	return exec.CommandContext(ctx, "open", url) // macOS
//}
