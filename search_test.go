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
		tc := tc
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
