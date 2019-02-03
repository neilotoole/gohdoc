package main

import (
	"fmt"
	"testing"
)

func TestProcessCmdOpenArgs(t *testing.T) {
	const cwd = "/go/src/github.com/neilotoole/gohdoc"

	testCases := []struct {
		arg0     string
		wantPath string
		wantPkg  string
		wantFrag string
	}{
		{arg0: "", wantPath: cwd, wantPkg: "gohdoc"},
		{arg0: "#Frag", wantPath: cwd, wantPkg: "gohdoc", wantFrag: "Frag"},
		{arg0: "#", wantPath: cwd, wantPkg: "gohdoc"},
		{arg0: ".", wantPath: cwd, wantPkg: "gohdoc"},
		{arg0: ".#Frag", wantPath: cwd, wantPkg: "gohdoc", wantFrag: "Frag"},
		{arg0: "./#Frag", wantPath: cwd, wantPkg: "gohdoc", wantFrag: "Frag"},
		{arg0: "./#", wantPath: cwd, wantPkg: "gohdoc"},
		{arg0: "/go/src/github.com/neilotoole/gohdoc", wantPath: "/go/src/github.com/neilotoole/gohdoc", wantPkg: "gohdoc"},
		{arg0: "/go/src/github.com/neilotoole/gohdoc/", wantPath: "/go/src/github.com/neilotoole/gohdoc", wantPkg: "gohdoc"},
		{arg0: "/go/src/github.com/neilotoole/../neilotoole/./gohdoc", wantPath: "/go/src/github.com/neilotoole/gohdoc", wantPkg: "gohdoc"},
		{arg0: "/go/src/github.com/neilotoole/gohdoc/sub/pkg", wantPath: "/go/src/github.com/neilotoole/gohdoc/sub/pkg", wantPkg: "pkg"},
		{arg0: "/go/src/github.com/neilotoole/gohdoc/#", wantPath: "/go/src/github.com/neilotoole/gohdoc", wantPkg: "gohdoc"},
		{arg0: "/go/src/github.com/neilotoole/gohdoc/#", wantPath: "/go/src/github.com/neilotoole/gohdoc", wantPkg: "gohdoc"},
		{arg0: "/go/src/github.com/neilotoole/gohdoc/.#", wantPath: "/go/src/github.com/neilotoole/gohdoc", wantPkg: "gohdoc"},
		{arg0: "/go/src/github.com/neilotoole/gohdoc/#Frag", wantPath: "/go/src/github.com/neilotoole/gohdoc", wantPkg: "gohdoc", wantFrag: "Frag"},
		{arg0: "github.com/neilotoole/gohdoc", wantPath: "/go/src/github.com/neilotoole/gohdoc/github.com/neilotoole/gohdoc", wantPkg: "github.com/neilotoole/gohdoc"},
		{arg0: "sub/pkg", wantPath: "/go/src/github.com/neilotoole/gohdoc/sub/pkg", wantPkg: "sub/pkg"},
		{arg0: "sub/pkg/", wantPath: "/go/src/github.com/neilotoole/gohdoc/sub/pkg", wantPkg: "sub/pkg"},
		{arg0: "sub/pkg#", wantPath: "/go/src/github.com/neilotoole/gohdoc/sub/pkg", wantPkg: "sub/pkg"},
		{arg0: "sub/pkg/#", wantPath: "/go/src/github.com/neilotoole/gohdoc/sub/pkg", wantPkg: "sub/pkg"},
		{arg0: "sub/pkg/.#", wantPath: "/go/src/github.com/neilotoole/gohdoc/sub/pkg", wantPkg: "sub/pkg"},
		{arg0: "sub/pkg#Frag", wantPath: "/go/src/github.com/neilotoole/gohdoc/sub/pkg", wantPkg: "sub/pkg", wantFrag: "Frag"},
		{arg0: "sub/pkg/#Frag", wantPath: "/go/src/github.com/neilotoole/gohdoc/sub/pkg", wantPkg: "sub/pkg", wantFrag: "Frag"},
		{arg0: "sub/pkg/.#Frag", wantPath: "/go/src/github.com/neilotoole/gohdoc/sub/pkg", wantPkg: "sub/pkg", wantFrag: "Frag"},
		{arg0: "fmt", wantPath: "/go/src/github.com/neilotoole/gohdoc/fmt", wantPkg: "fmt"},
		{arg0: "fmt#Println", wantPath: "/go/src/github.com/neilotoole/gohdoc/fmt", wantPkg: "fmt", wantFrag: "Println"},
	}

	for i, tc := range testCases {
		tc := tc
		t.Run(fmt.Sprintf("%d__%s", i, tc.arg0), func(t *testing.T) {
			app := &App{cwd: cwd, args: []string{tc.arg0}}

			gotPath, gotPkg, gotFrag := processCmdOpenArgs(app)

			if gotPath != tc.wantPath || gotPkg != tc.wantPkg || gotFrag != tc.wantFrag {
				t.Errorf("Wanted {%q %q %q} but got {%q %q %q}",
					tc.wantPath, tc.wantPkg, tc.wantFrag, gotPath, gotPkg, gotFrag)
			}
		})
	}
}
