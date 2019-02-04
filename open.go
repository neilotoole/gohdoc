package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

// cmdOpen is the primary functionality: it opens a browser for pkg in question.
func cmdOpen(app *App) error {

	if len(app.args) > 1 {
		return fmt.Errorf("must supply maximum one arg to gohdoc, but received %d: [%s]",
			len(app.args), strings.Join(app.args, " "))
	}

	err := requireServer(app)
	if err != nil {
		return err
	}
	err = loadServerPkgList(app)
	if err != nil {
		return err
	}
	// At this point, we know that the server is available, and app.serverPkgList is populated.
	return doCmdOpen(app)
}

// doCmdOpen does the main work of cmdOpen.
func doCmdOpen(app *App) error {
	pth, pkg, fragment := processCmdOpenArgs(app)

	// Try the path-based approach first.

	// pth looks something like /go/src/github.com/neilotoole/gohdoc
	// We'll iteratively look for a package that matches the path, trimming
	// a front segment each time. That is, we'll search for:
	//
	//   go/src/github.com/neilotoole/gohdoc
	//   src/github.com/neilotoole/gohdoc
	//   github.com/neilotoole/gohdoc
	//   neilotoole/gohdoc
	//   gohdoc

	// strip the leading slash, we don't need it
	pth = strings.TrimPrefix(pth, "/")

	parts := strings.Split(pth, "/")
	for {
		// reconstruct the path
		findPkg := strings.Join(parts, "/")
		log.Println("checking if pkg is listed:", findPkg)

		for _, serverPkg := range app.serverPkgList {
			if findPkg == serverPkg {
				log.Println("found in pkg list:", serverPkg)
				ok, err := serverPkgPageOK(app, serverPkg, true)
				if err != nil {
					return err
				}
				if !ok {
					return fmt.Errorf("should have been able to open pkg, but it seems not to exist: %s",
						serverPkg)
				}

				url := absPkgURL(app, serverPkg, fragment)
				return openBrowser(app, url)
			}
		}

		if len(parts) == 1 {
			break
		}

		parts = parts[1:]
	}

	// We weren't able to match the path (or subsections of it) against
	// serverPkgList, so we'll search for the pkg term.
	// When we get this far, we could be searching for partial
	// names like "byt", or "encoding/jso".
	matches, exactMatch := getPkgMatches(app.serverPkgList, pkg)
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

		return openBrowser(app, absPkgURL(app, matches[0], fragment))
	}

	// We don't have an exact match, so we'll iterate over the set of
	// possible matches and check if we can open that page.
	for _, match := range matches {
		ok, err := serverPkgPageOK(app, match, false)
		if err != nil {
			return err
		}
		if ok {
			err = openBrowser(app, absPkgURL(app, match, fragment))
			printPossibleMatches(app, pkg, matches)
			return err
		}
	}

	return fmt.Errorf("failed to find in server pkg list: %s", pkg)
}

// processCmdOpenArgs processes the command line args.
// This function returns a suggested absolute path, package name,
// and fragment (which may be empty). If the arg is relative, a
// suggested absolute path is constructed by joining with app.cwd.
// The returned path will always use forward slash (thus on Windows, the
// returned path is not a valid path).
func processCmdOpenArgs(app *App) (pkgpath, pkg, fragment string) {
	// There are several possibilities for args passed to the program, such as:
	// - no args                     = gohdoc .
	// - gohdoc .                    = gohdoc CWD
	// - gohdoc some/relative/path   = transformed to absolute path
	// - gohdoc arbitrary/pkg        = try relative path first, then search for arbitrary/pkg
	// - gohdoc /some/absolute/path  = passed through after path.Clean()
	//
	// - gohdoc fmt#Println          = open fmt with #fragment
	// - gohdoc fmt/#Println         = same as above
	// - gohdoc #Func                = open current dir godoc with #fragment
	// - gohdoc ./#Func              = same as above
	// - gohdoc .#Func               = same as above

	cwd := cleanFilePath(app.cwd)
	cwdBase := path.Base(cwd)

	if len(app.args) == 0 || strings.TrimSpace(app.args[0]) == "" {
		return cwd, cwdBase, ""
	}

	arg := strings.TrimSpace(app.args[0])
	arg = cleanFilePath(arg)

	if i := strings.IndexRune(arg, '#'); i >= 0 {
		if len(arg) == 1 {
			// i.e. arg is "#"
			return cwd, cwdBase, ""
		}

		if i < len(arg)-2 {
			frag := arg[i+1:]
			fragment = frag
		}
		arg = arg[0:i]
	}

	arg = path.Clean(arg)
	if arg == "." {
		return cwd, cwdBase, fragment
	}

	if path.IsAbs(arg) {
		return arg, path.Base(arg), fragment
	}

	return path.Join(cwd, arg), arg, fragment
}

// cleanFilePath strips any Windows volume name and converts
// to forward slash. This isn't particularly robust, but being
// that we don't need an actual working path, it should suffice.
// Also, too lazy to set project up to use platform-specific tests.
func cleanFilePath(p string) string {
	p = path.Clean(p)
	i := strings.IndexRune(p, ':')
	if i >= 1 {
		p = p[i+1:]
	}

	p = strings.Replace(p, `\`, "/", -1)
	p = path.Clean(p)
	return p
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
// The pkgPath arg should be a well-formed pkg path, e.g. "bytes", "encoding/json",
// or "github.com/neilotoole/gohdoc".
// An error is returned if a http failure occurs.
func serverPkgPageOK(app *App, pkgPath string, retry bool) (ok bool, err error) {
	if strings.HasPrefix(pkgPath, "/") || strings.HasSuffix(pkgPath, "/") {
		return false, fmt.Errorf("invalid pkg path (has '/' prefix or suffix): %s", pkgPath)
	}

	pageURL := absPkgURL(app, pkgPath, "")
	log.Println("verifying pkg page:", pageURL)
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

	cmd := openBrowserCmd(app.ctx, url) // openBrowserCmd is platform-specific
	err := cmd.Run()
	if err != nil {
		log.Printf("failed to open browser for %s: %v", url, err)
		return err
	}

	if app.cmd != nil {
		// if non-nil, we did start a server
		log.Printf("Opening %s on newly-started server\n", url)
	} else {
		log.Printf("Opening %s on pre-existing server\n", url)
	}

	fmt.Println(url)
	return nil
}

// absPkgURL returns the godoc http server URL for the supplied pkg.
func absPkgURL(app *App, fullPkgPath string, fragment string) string {

	fullPkgPath = strings.TrimPrefix(fullPkgPath, "/")
	fragment = strings.TrimSuffix(fragment, "#")
	if len(fragment) == 0 {
		return fmt.Sprintf("http://localhost:%d/pkg/%s/", app.port, fullPkgPath)
	}

	return fmt.Sprintf("http://localhost:%d/pkg/%s/#%s", app.port, fullPkgPath, fragment)
}

// printPkgsWithLink will - for each pkg - print a line with the pkg name and link.
func printPkgsWithLink(app *App, pkgs []string) {
	var width int
	for _, m := range pkgs {
		if len(m) > width {
			width = len(m)
		}
	}
	tpl := "%-" + strconv.Itoa(width) + "s    %s\n"

	for _, pkg := range pkgs {
		fmt.Printf(tpl, pkg, absPkgURL(app, pkg, ""))
	}
}
