package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// ensureServer checks if there's an existing godoc http server, or
// starts one if not. If ensureServer returns without an error, var pkgPageBody
// will have been set to the contents of the godoc http server's /pkg/ page.

func ensureServer(app *App) (err error) {

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
		app.pkgPageBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read body from %s: %v", pingURL, err)
		}
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
					app.pkgPageBody, err = ioutil.ReadAll(resp.Body)
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

// didStartServer returns true if gohdoc did start a server.
func didStartServer(app *App) bool {
	return app != nil && app.cmd != nil
}
