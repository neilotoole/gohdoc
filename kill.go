package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/shirou/gopsutil/process"
)

// cmdKillAll attempts to kill running processes named "godoc" with arg "-http".
// That is, it attempts to kill all running godoc http servers.
func cmdKillAll() error {
	ps, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to list processes: %v", err)
	}

	for _, p := range ps {
		name, err := p.NameWithContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to get process [%d] name: %v", p.Pid, err)
		}

		if strings.HasPrefix(name, "godoc") == false {
			continue
		}

		args, err := p.CmdlineSliceWithContext(ctx)
		if err != nil {
			return fmt.Errorf("failed to get command line args for process [%d]: %v", p.Pid, err)
		}

		for _, a := range args {
			if strings.HasPrefix(a, "-http") {
				log.Printf("found process named godoc [%d] with http server flag [%s]\n",
					p.Pid, strings.Join(args, " "))
				log.Printf("will attempt to kill process %s [%d]\n", name, p.Pid)

				err := p.KillWithContext(ctx)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to kill process [%d] '%s': %v\n", p.Pid, strings.Join(args, " "), err)
				} else {
					fmt.Printf("Killed process [%d]: '%s'\n", p.Pid, strings.Join(args, " "))
				}

				break
			}
		}

	}

	return nil
}
