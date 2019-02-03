package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// cmdOpen is the primary functionality: it opens a browser for pkg in question.
func cmdOpen(app *App) error {

	if len(app.args) > 1 {
		return fmt.Errorf("must supply maximum one arg to gohdoc, but received %d [%s]",
			len(app.args), strings.Join(app.args, ","))
	}

	err := ensureServer(app)
	if err != nil {
		return err
	}
	err = loadServerPkgList(app)
	if err != nil {
		return err
	}
	// At this point, we know that the server is available, and app.serverPkgList is populated.
	return doOpen(app, openBrowser)

}

// processCmdOpenArgs processes the command line args.
//
// There are several possibilities for args passed to the program
// - no args                     = gohdoc .
// - gohdoc .                    = gohdoc CWD
// - gohdoc some/relative/path   = transformed to absolute path
// - gohdoc arbitrary/pkg        = try relative path first, then search for arbitrary/pkg
// - gohdoc /some/absolute/path  = passed through after path.Clean()
// - gohdoc more than one arg    = error
//
// - gohdoc fmt#Println          = open fmt with #fragment
// - gohdoc fmt/#Println         = same as above
// - gohdoc #Func                = open current dir godoc with #fragment
// - gohdoc ./#Func              = same as above
// - gohdoc .#Func               = same as above
func processCmdOpenArgs(app *App) (path string, pkg string, fragment *string) {

	cwd := filepath.Clean(app.cwd)
	cwdBase := filepath.Base(cwd)
	if len(app.args) == 0 || app.args[0] == "" {
		return app.cwd, cwdBase, nil
	}

	raw := app.args[0]
	arg := strings.TrimSpace(raw)

	if i := strings.IndexRune(arg, '#'); i >= 0 {
		if len(arg) == 1 {
			// i.e. arg is "#"
			return cwd, cwdBase, nil
		}

		if i < len(arg)-2 {
			// e.g. arg is "blah#"
			frag := arg[i+1:]
			fragment = &frag
		}

		if i > 0 {
			arg = arg[0:i]
		} else {
			arg = ""
		}
	}

	arg = filepath.Clean(arg)
	if arg == "." {
		return cwd, cwdBase, fragment
	}

	base := filepath.Base(arg)
	if filepath.IsAbs(arg) {
		return arg, base, fragment
	}

	// arg is not absolute
	return filepath.Join(cwd, arg), arg, fragment

}

// doOpen does the main work of cmdOpen.
func doOpen(app *App, fnOpenBrowser func(app *App, url string) error) error {

	var arg string
	var originalArg *string
	var fragment *string

	if len(app.args) == 0 || app.args[0] == "" {
		// If no arg supplied, then assume current working directory
		arg = app.cwd
	} else {
		arg = app.args[0]

		if i := strings.IndexRune(arg, '#'); i >= 0 {
			frag := arg[i:]
			fragment = &frag
			if i > 0 {
				arg = arg[0 : i-1]
			} else {
				arg = ""
			}
		}

		if arg == "." {
			// Special case, we expand "." to cwd
			arg = app.cwd
		}
		originalArg = &app.args[0]
	}

	log.Println("#fragment:", fragment)

	tentativePkgPath := filepath.Clean(arg)
	if !path.IsAbs(tentativePkgPath) {
		tentativePkgPath = filepath.Join(app.cwd, tentativePkgPath)
	}

	log.Println("using tentativePkgPath:", tentativePkgPath)

	// We know that the server is accessible, check that we can actually access our pkg page
	pkgPath, tentativeOnGopath := determinePackageOnGopath(app.gopath, tentativePkgPath)
	if tentativeOnGopath {
		log.Println("attempting tentativePkgPath:", pkgPath)
		ok, err := serverPkgPageOK(app, pkgPath, true)
		if err != nil {
			return err
		}
		if ok {
			return fnOpenBrowser(app, absPkgURL(app, pkgPath, nil)) // TODO: #fragment
		}

		log.Println("server doesn't have pkgPath:", pkgPath)

	}

	if originalArg != nil && !filepath.IsAbs(arg) {
		// If originalArg is absolute, then we've got no business searching for it

		// But if we get this far, we could be searching for something like "fmt",
		// or partial names like "byt", or "encoding/jso".
		matches, exactMatch := getPkgMatches(app.serverPkgList, arg)
		if exactMatch {
			// If we've got an exact match, we only want to open that page
			ok, err := serverPkgPageOK(app, matches[0], true)
			if err != nil {
				return err
			}
			if !ok {
				// shouldn't happen
				return fmt.Errorf("should have been able to open this, but it seems not to exist: %s", matches[0])
			}

			return fnOpenBrowser(app, absPkgURL(app, matches[0], nil)) // TODO: #fragment
		}

		// We don't have an exact match
		for _, match := range matches {
			ok, err := serverPkgPageOK(app, match, false)
			if err != nil {
				return err
			}
			if ok {
				printPossibleMatches(app, arg, matches)
				return fnOpenBrowser(app, absPkgURL(app, match, nil)) // TODO: #fragment
			}
		}

	}

	return fmt.Errorf("failed to find %s", arg)
}

func printPossibleMatches(app *App, arg string, matches []string) {
	const maxPkgList = 10
	if len(matches) > maxPkgList {
		const tpl = "Found %d possible matches; showing first %d. To see full set: gohdoc -searchv %s\n"

		fmt.Printf(tpl, len(matches), maxPkgList, arg)
		matches = matches[0:maxPkgList]
	} else {
		fmt.Printf("Found %d possible matches:\n", len(matches))
	}
	printPkgsWithLink(app, matches)
}

// serverPkgPageOK returns true, nil if pkgPath exists on the server.
// The pkgPath arg must be a well-formed pkg path, e.g. "bytes", "encoding/json",
// or "github.com/neilotoole/gohdoc".
// An error is returned if a http failure occurs.
func serverPkgPageOK(app *App, pkgPath string, retry bool) (ok bool, err error) {
	log.Printf("serverPkgPageOK: pkgPath=%q\n", pkgPath)
	if strings.HasPrefix(pkgPath, "/") || strings.HasSuffix(pkgPath, "/") {
		return false, fmt.Errorf("invalid pkg path (has '/' prefix or suffix): %s", pkgPath)
	}

	pageURL := absPkgURL(app, pkgPath, nil)
	log.Println("serverPkgPageOK: attempting pageURL:", pageURL)
	timeout := time.Now()
	if retry {
		timeout = time.Now().Add(time.Millisecond * 500)
	}
	var resp *http.Response
	for {

		resp, err = http.Head(pageURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			return true, nil
		}

		if !retry || time.Now().After(timeout) {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	if err != nil {
		return false, fmt.Errorf("failed to access godoc http server: %v", err)
	}

	return false, nil
}

// openBrowser opens a browser for url. It delegates creation of the platform-specific
// exec.Cmd to build tag-gated implementations of openBrowserCmd.
func openBrowser(app *App, url string) error {
	log.Println("attempting to open a browser for:", url)

	//cmd := newOpenBrowserCmd(app.ctx, url) // newOpenBrowserCmd is platform-specific
	cmd := openBrowserCmd(app.ctx, url) // openBrowserCmd is platform-specific
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
