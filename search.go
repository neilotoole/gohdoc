package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
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
		log.Printf("No package found matching %s\n", term)
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
	err := requireServer(app)
	if err != nil {
		return err
	}

	if app == nil || len(app.serverPkgPageBody) == 0 {
		return errors.New("apparently no data from godoc http server /pkg")
	}
	r := bytes.NewReader(app.serverPkgPageBody)

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
			v = strings.TrimSuffix(v, "/")
			pkgs = append(pkgs, v)
		}
	})

	return pkgs, nil
}
