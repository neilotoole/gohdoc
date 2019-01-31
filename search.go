package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// cmdList lists all pkgs on the godoc http server.
func cmdList(app *App) error {
	err := loadServerPkgList(app)
	if err != nil {
		return err
	}

	pkgs := app.serverPkgList

	if app.flagListv {
		// When verbose, also print a link to the pkg
		printPkgsWithLink(app, pkgs)
	} else {
		for _, pkg := range pkgs {
			fmt.Println(pkg)
		}
	}
	return nil
}

// cmdSearch lists all godoc http server packages that match the argument.
func cmdSearch(app *App) error {
	if len(app.args) != 1 {
		return fmt.Errorf("search command takes exactly one arg")
	}

	err := loadServerPkgList(app)
	if err != nil {
		return err
	}

	pkgs := app.serverPkgList
	term := app.args[0]
	log.Printf("searching %d pkg names for term %q", len(pkgs), term)

	matches, _ := getPkgMatches(pkgs, term)
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "No package found matching %s", term)
		return nil
	}

	if app.flagSearchv {
		printPkgsWithLink(app, matches)
	} else {
		for _, pkg := range matches {
			fmt.Println(pkg)
		}
	}

	return nil

}

// loadServerPkgList loads the list of pkgs from the server, and
// sets app.serverPkgList with that data.
func loadServerPkgList(app *App) error {
	err := ensureServer(app)
	if err != nil {
		return err
	}

	r, err := getPkgPageBodyReader(app)
	if err != nil {
		return err
	}

	pkgs, err := scrapePkgPage(r)
	if err != nil {
		return err
	}
	app.serverPkgList = pkgs
	return nil
}

// getPkgMatches returns the set of pkg names that match arg s,
// with the best match first. The "best match" algorithm is pretty trivial.
// If there's an exact match of s against pkgs, then exactMatch is returned
// true, and matches has minimum length 1 (although it may be larger).
func getPkgMatches(pkgs []string, s string) (matches []string, exactMatch bool) {
	if len(s) == 0 || len(pkgs) == 0 {
		return matches, false
	}

	// We match on whether s is a suffix, prefix, or is contained in pkg
	var sufMatches, preMatches, containMatches []string

	for _, pkg := range pkgs {
		if pkg == s {
			exactMatch = true
			continue
		}

		if strings.HasSuffix(pkg, s) {
			sufMatches = append(sufMatches, pkg)
			continue
		}
		if strings.HasPrefix(pkg, s) {
			preMatches = append(preMatches, pkg)
			continue
		}

		if strings.Contains(pkg, s) {
			containMatches = append(containMatches, pkg)
		}
	}

	sort.Strings(sufMatches)
	sort.Strings(preMatches)
	sort.Strings(containMatches)

	if exactMatch {
		// If there's an exact match, it should be the first result
		matches = append(matches, s)
	}
	matches = append(matches, sufMatches...)
	matches = append(matches, preMatches...)
	matches = append(matches, containMatches...)

	return matches, exactMatch

}

// scrapePkgPage scrapes the /pkg HTML, returning all pkg
// names listed on that page.
func scrapePkgPage(r io.Reader) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	var pkgs []string

	selector := ".pkg-dir td.pkg-name a[href]"

	doc.Find(selector).Each(func(_ int, s *goquery.Selection) {
		v, ok := s.Attr("href")
		if ok {
			// The link looks like "encoding/json/"
			if strings.HasSuffix(v, "/") {
				// Get rid of the trailing slash
				v = v[0 : len(v)-1]
			}
			pkgs = append(pkgs, v)
		}
	})

	return pkgs, nil
}

// determinePackageOnGopath attempts to determine if the supplied dirpath
// could be on our GOPATH, returning the pkg path relative to GOPATH.
//
// For example, if GOPATH is /go and dirpath is /go/src/github.com/neilotoole/gohdoc
// then the return value would be github.com/neilotoole/gohdoc.
func determinePackageOnGopath(gopath string, dirpath string) (pkg string, ok bool) {
	log.Printf("determinePackage: GOPATH=%q dirpath=%q", gopath, dirpath)

	gopath = filepath.Clean(gopath)

	if gopath == "" {
		log.Println("GOPATH is empty")
		return "", false
	}

	if !filepath.IsAbs(gopath) {
		log.Println("GOPATH is not an absolute path:", gopath)
		return "", false
	}

	gopathParts := strings.Split(gopath, string(filepath.Separator)) // this seems wrong
	if len(gopathParts) < 1 {
		// not planning on supporting a pathological case where GOPATH is /
		log.Println("GOPATH is too short: ", gopath)
		return "", false
	}

	// gopathPrefix, at a minimum, should be something like /go/src
	gopathPrefix := filepath.Join(gopath, "src")

	dirpath = filepath.Clean(dirpath)
	// At a minimum, dirpath should be something like /go/src/mypkg
	dirpathParts := strings.Split(dirpath, string(filepath.Separator))
	if len(dirpathParts) < 3 {
		log.Println("dirpath is too short:", dirpath, strings.Join(dirpathParts, " | "))
	}

	if dirpath == gopathPrefix {
		log.Println("dirpath is at GOPATH/src, needs to be in a sub-dir", dirpath)
		return "", false
	}

	if !strings.HasPrefix(dirpath, gopathPrefix) {
		log.Printf("dirpath %q is not beneath GOPATH/src %q\n", dirpath, gopathPrefix)
		return "", false
	}

	// strip out the gopath's path to just get the pkg path
	pkg = dirpath[len(gopathPrefix)+1:]
	if pkg == "" {
		// probably can never reach this due to checks above
		log.Printf("after stripping out GOPATH/src prefix %q from dirpath, empty path remaining: %s", gopathPrefix, pkg)
		return "", false
	}

	return pkg, true
}
