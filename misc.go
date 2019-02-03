package main

import (
	"fmt"
	"strconv"
)

// absPkgURL returns the godoc http server URL for the supplied pkg.
func absPkgURL(app *App, fullPkgPath string, fragment *string) string {
	if fragment == nil || len(*fragment) == 0 {
		return fmt.Sprintf("http://localhost:%d/pkg/%s/", app.port, fullPkgPath)
	}

	frag := *fragment
	if frag[0] != '#' {
		frag = "#" + frag
	}

	return fmt.Sprintf("http://localhost:%d/pkg/%s/%s", app.port, fullPkgPath, frag)
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
		fmt.Printf(tpl, pkg, absPkgURL(app, pkg, nil))
	}
}
