package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

// absPkgURL returns the godoc http server URL for the supplied pkg.
//
//	absPkgURL("sync/atomic") --> "http://localhost:6060/pkg/sync/atomic/"
func absPkgURL(fullPkgPath string) string {
	return fmt.Sprintf("http://localhost:%d/pkg/%s/", port, fullPkgPath)
}

// determinePackage returns the package path of the supplied dir (relative to the GOPATH).
// For example, the return value could be "github.com/neilotoole/gohdoc".
func determinePackage(pkgPath string) (pkg string, err error) {
	log.Println("using path dir:", pkgPath)

	gopathPrefix := path.Join(gopath, "src")

	if !strings.HasPrefix(pkgPath, gopathPrefix) || pkgPath == gopathPrefix {
		return "", fmt.Errorf("current dir does not appear to be a valid package on GOPATH")
	}

	pkg = pkgPath[len(gopathPrefix)+1:]
	if pkg == "" {
		return "", fmt.Errorf("current dir does not appear to be a valid package on GOPATH")
	}
	return pkg, nil
}

// exitOnErr prints err info and calls os.Exit(1) if err is not nil.
func exitOnErr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		if gopath != "" {
			fmt.Fprintln(os.Stderr, " info: was using GOPATH:", gopath)
		}
		if cmd != nil {
			// If we're exiting due to an error, and we already started a godoc http server, kill it
			log.Printf("killing the godoc http server [%d] that gohdoc started\n", cmd.Process.Pid)
			_ = cmd.Process.Kill()
		}
		os.Exit(1)
	}
}

// printPkgsWithLink will, for each pkg print a line with the pkg name and link.
func printPkgsWithLink(pkgs []string) {
	var width int
	for _, m := range pkgs {
		if len(m) > width {
			width = len(m)
		}
	}
	tpl := "%-" + strconv.Itoa(width) + "s    %s\n"

	for _, pkg := range pkgs {
		fmt.Printf(tpl, pkg, absPkgURL(pkg))
	}
}

//func splitFragment(s string) (path, fragment string) {
//
//}
