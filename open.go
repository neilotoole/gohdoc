package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// cmdOpen is the primary functionality: it opens a browser for pkg in question.
func cmdOpen() error {

	// There are several possibilities for args passed to the program
	// - no args                     = gohdoc .
	// - gohdoc .                    = gohdoc CWD
	// - gohdoc some/relative/path   = transformed to absolute path
	// - gohdoc arbitrary/pkg        = try relative path first, then search for arbitrary/pkg
	// - gohdoc /some/absolute/path  = passed through after path.Clean()
	// - gohdoc more than one arg    = error

	args := flag.Args()
	var pkgPath string
	var originalArg *string

	switch len(args) {
	case 0:
		// If no arg supplied, then assume current working directory
		pkgPath = cwd

	case 1:
		pkgPath = args[0]
		if pkgPath == "." {
			// Special case, we expand "." to cwd
			pkgPath = cwd
		}
		originalArg = &args[0]
	default:
		exitOnErr(fmt.Errorf("must supply maximum one arg to gohdoc, but received [%s]", strings.Join(args, ",")))
	}

	pkgPath = path.Clean(pkgPath)
	if !path.IsAbs(pkgPath) {
		pkgPath = path.Join(cwd, pkgPath)
	}

	log.Println("using pkg path:", pkgPath)
	envar, ok := os.LookupEnv(envGodocPort)
	if ok {
		log.Printf("found envar %s: %s", envGodocPort, envar)
	}
	if ok && len(strings.TrimSpace(envar)) > 0 {
		var err error
		port, err = strconv.Atoi(envar)
		if err != nil || port < 1 || port > 65536 {
			return fmt.Errorf("%s was set, but value is invalid: %s", envGodocPort, envar)
		}
	}

	var err error
	err = ensureServer()
	if err != nil {
		return err
	}

	// We know that the server is accessible, check that we can actually access our pkg page
	pkg, err := determinePackage(pkgPath)
	if err != nil {
		return err
	}
	log.Println("attempting pkg:", pkg)

	pageURL := absPkgURL(pkg)
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

		if cmd == nil {
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
		err = openBrowser(pageURL)
		if err != nil {
			return err
		}

		return nil
	}

	if originalArg != nil && !path.IsAbs(*originalArg) {
		log.Printf("failed to find %q at %s", *originalArg, pageURL)
		// Alternative strategy for args such as "bytes" or "crypto/dsa"

		log.Printf("will attempt to search for %q", *originalArg)
		pkgList, err := extractPkgList()
		if err != nil {
			return err
		}

		matches := getPkgMatches(pkgList, *originalArg)
		log.Printf("search %q returned %d results\n", *originalArg, len(matches))
		if len(matches) > 0 {
			log.Printf("results: %s\n", strings.Join(matches, " "))
			// Let's use the best match
			u := absPkgURL(matches[0])
			resp, err := http.Head(u)
			if err == nil && resp.StatusCode == http.StatusOK {
				log.Printf("probable match at %s", u)

				const maxPkgList = 10

				err := openBrowser(u)

				if err == nil && len(matches) > 1 {
					if len(matches) > maxPkgList {
						const tpl = "Found %d possible matches; showing first %d. To see full set: gohdoc -searchv %s\n"

						fmt.Printf(tpl, len(matches), maxPkgList, *originalArg)
						matches = matches[0:maxPkgList]
					} else {
						fmt.Printf("Found %d possible matches:\n", len(matches))
					}
					printPkgsWithLink(matches)
				}

				return err
			}
		}
	}

	return fmt.Errorf("got %s from %s", resp.Status, pageURL)
}
