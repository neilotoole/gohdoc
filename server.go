package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/shirou/gopsutil/process"
)

// cmdServers lists godoc http server processes.
func cmdServers(app *App) error {
	ctx := app.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	ps, err := listServerProcesses(ctx)
	if err != nil {
		return err
	}

	for _, p := range ps {
		fmt.Println(p)
	}
	return nil
}

// cmdKillAll attempts to kill running processes named "godoc" with arg "-http".
// That is, it attempts to kill all running godoc http servers.
func cmdKillAll(app *App) error {
	ctx := app.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	ps, err := listServerProcesses(ctx)
	if err != nil {
		return err
	}

	var errCount int

	for _, p := range ps {
		err := p.process.KillWithContext(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s  :  %s\n", p, err)
			errCount++

		} else {
			fmt.Println(p)
		}
	}

	switch errCount {
	case 0:
		return nil
	case 1:
		if len(ps) == 1 {
			return fmt.Errorf("failed to kill 1 process")
		}
		fallthrough
	default:
		return fmt.Errorf("failed to kill %d of %d processes", errCount, len(ps))
	}
}

// processMeta encapsulate a process.Process and some already-loaded
// metadata, to avoid having to load the metadata again.
type processMeta struct {
	process *process.Process
	pid     int32
	// name is process name
	name     string
	username string
	cmdline  []string
}

func (p processMeta) String() string {
	username := p.username
	if username == "" {
		username = "UNKNOWN_USER"
	}

	return fmt.Sprintf("%-16s  %-6d  %s", username, p.pid, strings.Join(p.cmdline, " "))
}

func listServerProcesses(ctx context.Context) ([]processMeta, error) {
	var matches []processMeta

	ps, err := process.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list processes: %v", err)
	}

	for _, p := range ps {
		name, err := p.NameWithContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get process [%d] name: %v", p.Pid, err)
		}

		if strings.HasPrefix(name, "godoc") == false {
			continue
		}

		args, err := p.CmdlineSliceWithContext(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get command line args for process [%d]: %v", p.Pid, err)
		}

		uname, _ := p.Username() // not critical that we get the uname

		for _, a := range args {
			if strings.HasPrefix(a, "-http") {
				// TODO: should refine this to only kill servers on same port as us?
				log.Printf("found process named godoc [%d] with http server flag [%s]\n",
					p.Pid, strings.Join(args, " "))

				match := processMeta{process: p, pid: p.Pid, name: name, username: uname, cmdline: args}
				matches = append(matches, match)
				break
			}
		}

	}
	return matches, nil
}

// ensureServer checks if there's an existing godoc http server, or starts one if
// not. If ensureServer returns without an error, app.serverPkgPageBody will have
// been set to the contents of the godoc http server's /pkg page.
func ensureServer(app *App) (err error) {
	if len(app.serverPkgPageBody) > 0 {
		// If this is already set, then we've already determined that a server exists.
		return nil
	}

	serverExisted := false

	pingURL := fmt.Sprintf("http://localhost:%d/pkg", app.port)

	resp, err := http.Get(pingURL)
	if err != nil {
		log.Printf("apparently there's no existing godoc http server at %s: %v", pingURL, err)
		serverExisted = false
	} else if resp.StatusCode == http.StatusOK {
		serverExisted = true
		log.Println("found existing godoc server at", pingURL)
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read body from %s: %v", pingURL, err)
		}
		app.serverPkgPageBody = b
	}

	if !serverExisted {
		log.Println("no existing godoc server, will attempt to start one, which will continue to run in background after gohdoc exits")

		err = startServer(app)
		if err != nil {
			return err
		}
	}

	log.Printf("godoc server is running at %s", pingURL)
	var timeout time.Time

	if !serverExisted {
		// Check that the newly-started server is accessible
		timeout = time.Now().Add(time.Second * 2)

		for {
			if time.Now().After(timeout) {
				break
			}

			resp, err = http.Get(pingURL)
			if err == nil {
				if resp.StatusCode == http.StatusOK {
					app.serverPkgPageBody, err = ioutil.ReadAll(resp.Body)
					if err != nil {
						return fmt.Errorf("failed to read body from %s: %v", pingURL, err)
					}
					_ = resp.Body.Close()
					break
				}
				_ = resp.Body.Close()
			}

			time.Sleep(time.Millisecond * 100)
		}

		if err != nil {
			return fmt.Errorf("failed to access godoc http server: %v", err)
		} else if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("got %s from %s", resp.Status, pingURL)
		}
	}

	return nil
}

// startServer starts a godoc http server. On success, the app.Cmd field will
// be set to the exec.Cmd used to start the server.
func startServer(app *App) error {
	var cmd *exec.Cmd

	ctx := app.ctx
	if ctx == nil {
		ctx = context.Background()
	}

	flagDebug := app.flagDebug
	flagDebug = false // TODO: get rid of this line, or introduce -vv flag
	if flagDebug {
		cmd = exec.CommandContext(ctx, "godoc", fmt.Sprintf("-http=:%d", app.port), "-v", "-index", "-index_throttle=0.5")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

	} else {
		cmd = exec.CommandContext(ctx, "godoc", fmt.Sprintf("-http=:%d", app.port), "-index", "-index_throttle=0.5")
	}

	err := cmd.Start()
	if err != nil {
		return err
	}
	// If the cmd started successfully, assign it to the app.
	app.cmd = cmd

	fmt.Printf("Started godoc server [%d] for GOPATH %s at http://localhost:%d\n", cmd.Process.Pid, app.gopath, app.port)
	fmt.Printf("Server will continue to run in the background. Kill with: gohdoc -killall\n\n")

	return nil

}
