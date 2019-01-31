package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestScrapePkgPage(t *testing.T) {

	var testCases []string
	versions := []string{"1.5", "1.6", "1.7", "1.8", "1.9", "1.10", "1.11"}
	for _, v := range versions {
		testCases = append(testCases, fmt.Sprintf("testdata/pkg_%s.html", v))
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			b, err := ioutil.ReadFile(tc)
			if err != nil {
				t.Error(err)
			}

			pkgs, err := scrapePkgPage(bytes.NewReader(b))
			if err != nil {
				t.Fatal(err)
			}

			if len(pkgs) < 150 { // approx 150 pkgs in stdlib
				t.Errorf("should have more than %d pkgs", len(pkgs))
			}

			// verify that we have the gohdoc pkg in the output
			found := false
			for _, pkg := range pkgs {
				if strings.HasSuffix(pkg, "neilotoole/gohdoc") {
					found = true
					break
				}
			}
			if !found {
				t.Error("didn't find gohdoc pkg in output")
			}
		})
	}

}

func TestDeterminePackage(t *testing.T) {

	testCases := []struct {
		gopath  string
		dirPath string
		wantPkg string
		wantOK  bool
	}{
		{"/go", "/go/src/github.com/neilotoole/gohdoc/", "github.com/neilotoole/gohdoc", true},
		{"/go", "/go/src/github.com/neilotoole/gohdoc", "github.com/neilotoole/gohdoc", true},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d %s", i, tc.dirPath), func(t *testing.T) {

			gotPkg, gotOK := determinePackageOnGopath(tc.gopath, tc.dirPath)
			if gotOK != tc.wantOK {
				t.Errorf("wantOK=%v but gotOK=%v", tc.wantOK, gotOK)
			}

			if gotPkg != tc.wantPkg {
				t.Errorf("wantPkg=%s but gotPkg=%s", tc.wantPkg, gotPkg)
			}

		})
	}
}
