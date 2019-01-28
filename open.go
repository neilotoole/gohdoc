package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"path"
	"strings"
	"time"
)

// cmdOpen is the primary functionality: it opens a browser for pkg in question.
func cmdOpen(app *App) error {

	// There are several possibilities for args passed to the program
	// - no args                     = gohdoc .
	// - gohdoc .                    = gohdoc CWD
	// - gohdoc some/relative/path   = transformed to absolute path
	// - gohdoc arbitrary/pkg        = try relative path first, then search for arbitrary/pkg
	// - gohdoc /some/absolute/path  = passed through after path.Clean()
	// - gohdoc more than one arg    = error

	var arg string
	var originalArg *string

	switch len(app.args) {
	case 0:
		// If no arg supplied, then assume current working directory
		arg = app.cwd

	case 1:
		arg = app.args[0]
		if arg == "." {
			// Special case, we expand "." to cwd
			arg = app.cwd
		}
		originalArg = &app.args[0]
	default:
		return fmt.Errorf("must supply maximum one arg to gohdoc, but received [%s]", strings.Join(app.args, ","))
	}

	pkgPath := path.Clean(arg)
	if !path.IsAbs(pkgPath) {
		pkgPath = path.Join(app.cwd, pkgPath)
	}

	log.Println("using pkg path:", pkgPath)

	err := ensureServer(app)
	if err != nil {
		return err
	}

	// We know that the server is accessible, check that we can actually access our pkg page
	pkg, err := determinePackage(app.gopath, pkgPath)
	if err != nil {
		return err
	}
	log.Println("attempting pkg:", pkg)

	pageURL := absPkgURL(app, pkg)
	timeout := time.Now().Add(time.Millisecond * 500)
	var resp *http.Response
	for {
		// Not sure if we still need this, but I occasionally ran into
		// the situation where godoc had not yet indexed
		if time.Now().After(timeout) {
			break
		}

		resp, err = http.Head(pageURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if app.cmd == nil {
			// Well, if we didn't start the server ourselves, it should
			// already be indexed, no need to wait.
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	if err != nil {
		return fmt.Errorf("failed to access godoc http server: %v", err)
	}

	if resp.StatusCode == http.StatusOK {
		// happy path, we have our page
		return openBrowser(app, pageURL)
	}

	if originalArg != nil && !path.IsAbs(*originalArg) {
		log.Printf("failed to find %q at %s", *originalArg, pageURL)
		// Alternative strategy for args such as "bytes" or "crypto/dsa"

		log.Printf("will attempt to search for %q", *originalArg)

		r, err := getPkgPageBodyReader(app)
		if err != nil {
			return err
		}

		pkgList, err := scrapePkgPage(r)
		if err != nil {
			return err
		}

		matches := getPkgMatches(pkgList, *originalArg)
		log.Printf("search %q returned %d results\n", *originalArg, len(matches))
		if len(matches) > 0 {
			log.Printf("results: %s\n", strings.Join(matches, " "))
			// Let's use the best match
			u := absPkgURL(app, matches[0])
			resp, err := http.Head(u)
			if err == nil && resp.StatusCode == http.StatusOK {
				log.Printf("probable match at %s", u)

				const maxPkgList = 10

				err := openBrowser(app, u)

				if err == nil && len(matches) > 1 {
					if len(matches) > maxPkgList {
						const tpl = "Found %d possible matches; showing first %d. To see full set: gohdoc -searchv %s\n"

						fmt.Printf(tpl, len(matches), maxPkgList, *originalArg)
						matches = matches[0:maxPkgList]
					} else {
						fmt.Printf("Found %d possible matches:\n", len(matches))
					}
					printPkgsWithLink(app, matches)
				}

				return err
			}
		}
	}

	return fmt.Errorf("got %s from %s", resp.Status, pageURL)
}

func getPkgPageBodyReader(app *App) (io.Reader, error) {
	if app == nil || len(app.serverPkgPageBody) == 0 {
		return nil, errors.New("apparently no data from godoc http server /pkg")
	}
	return bytes.NewReader(app.serverPkgPageBody), nil
}

// newOpenBrowserCmdFn is set by platform-specific go files
var newOpenBrowserCmdFn func(ctx context.Context, url string) *exec.Cmd

// openBrowser opens a browser for url.
func openBrowser(app *App, url string) error {
	log.Println("attempting to open a browser for ", url)

	//cmd := newOpenBrowserCmd(app.ctx, url) // newOpenBrowserCmd is platform-specific
	cmd := newOpenBrowserCmdFn(app.ctx, url) // newOpenBrowserCmd is platform-specific
	err := cmd.Run()
	if err != nil {
		log.Printf("failed to open browser for %s: %v", url, err)
		return err
	}

	if app.cmd != nil {
		// if non-nil, we did start a server
		fmt.Printf("Opening %s on GOPATH %s\n", url, app.gopath)
	} else {
		fmt.Printf("Opening %s on already-existing server\n", url)
	}

	return nil
}
