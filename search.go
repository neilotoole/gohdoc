package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// cmdList lists all pkgs on the godoc http server.
func cmdList(app *App) error {
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

	term := app.args[0]
	log.Printf("searching %d pkg names for term %q", len(pkgs), term)

	matches := getPkgMatches(pkgs, term)
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

// getPkgMatches returns the set of pkg names that match arg s,
// with the best match first. The "best match" algorithm is pretty trivial.
func getPkgMatches(pkgs []string, s string) []string {
	var matches []string
	if len(s) == 0 || len(pkgs) == 0 {
		return matches
	}

	// exactMatch is set to true if an exact match is found
	var exactMatch bool
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

	return matches

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

// determinePackage returns the package path of the supplied dir (relative to the GOPATH).
// For example, the return value could be "github.com/neilotoole/gohdoc".
func determinePackage(gopath string, dirPath string) (pkg string, err error) {
	const tpl = "dir %s does not appear to be a valid package on GOPATH %s"

	dirPath = path.Clean(dirPath)
	log.Println("using path dir:", dirPath)

	gopathPrefix := path.Join(gopath, "src")

	if !strings.HasPrefix(dirPath, gopathPrefix) || dirPath == gopathPrefix {
		return "", fmt.Errorf(tpl, dirPath, gopath)
	}

	// strip out the gopath's path to just get the pkg path
	pkg = dirPath[len(gopathPrefix)+1:]
	if pkg == "" {
		return "", fmt.Errorf(tpl, dirPath, gopath)
	}
	return pkg, nil
}
