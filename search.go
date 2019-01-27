package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
)

// cmdList lists all pkgs on the godoc http server
func cmdList() error {
	err := ensureServer()
	if err != nil {
		return err
	}

	pkgs, err := extractPkgList()
	if err != nil {
		return err
	}

	if flagListv {
		// When verbose, also print a link to the pkg
		printPkgsWithLink(pkgs)
	} else {
		for _, pkg := range pkgs {
			fmt.Println(pkg)
		}
	}
	return nil
}

// cmdSearch lists all godoc http server packages that match the argument.
func cmdSearch() error {

	args := flag.Args()
	if len(args) != 1 {
		return fmt.Errorf("-match command takes exactly one arg")
	}

	err := ensureServer()
	if err != nil {
		return err
	}

	pkgs, err := extractPkgList()
	if err != nil {
		return err
	}

	matches := getPkgMatches(pkgs, args[0])
	if len(matches) == 0 {
		fmt.Fprintf(os.Stderr, "No package found matching %s", args[0])
		return nil
	}

	if flagSearchv {
		printPkgsWithLink(matches)
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
		}

		if strings.Contains(pkg, s) {
			containMatches = append(containMatches, pkg)
		}
	}

	sort.Strings(sufMatches)
	sort.Strings(preMatches)
	sort.Strings(containMatches)

	if exactMatch {
		matches = append(matches, s)
	}
	matches = append(matches, sufMatches...)
	matches = append(matches, preMatches...)
	matches = append(matches, containMatches...)

	return matches

}
