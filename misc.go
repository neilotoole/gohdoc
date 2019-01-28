package main

import (
	"fmt"
	"strconv"
)

// absPkgURL returns the godoc http server URL for the supplied pkg.
//
//	absPkgURL("sync/atomic") --> "http://localhost:6060/pkg/sync/atomic/"
func absPkgURL(app *App, fullPkgPath string) string {
	return fmt.Sprintf("http://localhost:%d/pkg/%s/", app.port, fullPkgPath)
}

// printPkgsWithLink will, for each pkg print a line with the pkg name and link.
func printPkgsWithLink(app *App, pkgs []string) {
	var width int
	for _, m := range pkgs {
		if len(m) > width {
			width = len(m)
		}
	}
	tpl := "%-" + strconv.Itoa(width) + "s    %s\n"

	for _, pkg := range pkgs {
		fmt.Printf(tpl, pkg, absPkgURL(app, pkg))
	}
}

//func splitFragment(s string) (path, fragment string) {
//
//}
