// Package main is the gohdoc implementation.
package main

// TODO:
// - Windows and Linux support
// - Support arbitrary packages (e.g. gohdoc fmt or gohdoc sql/driver)

import (
	"context"
	"flag"
	"fmt"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
)

// envGodocPort ("GODOC_HTTP_PORT") is the envar used to override the
// default godoc http server port (6060).
const envGodocPort = "GODOC_HTTP_PORT"

const Version = "0.1"

// port is the port to start a godoc http server on, if necessary to do so. Defaults
// to 6060, but can be overridden by envar GODOC_HTTP_PORT.
var port = 6060

// gopath is the calculated value of GOPATH.
var gopath string

// cwd is the current working directory
var cwd string

// cmd is the Cmd used to start a godoc http server, if necessary to do so.
var cmd *exec.Cmd

var ctx context.Context

// pkgPageBody holds the contents of the godoc http server's /pkg/ page
var pkgPageBody []byte

var flagHelp bool
var flagList bool
var flagListv bool
var flagSearch bool
var flagSearchv bool
var flagKillAll bool
var flagDebug bool

func init() {

	var err error
	cwd, err = os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get current working directory: %v\n", err)
		os.Exit(1)
	}

	flag.BoolVar(&flagHelp, "help", false, "print help")
	flag.BoolVar(&flagList, "list", false, "list all pkgs from godoc http server")
	flag.BoolVar(&flagListv, "listv", false, "like -list but with verbose output")
	flag.BoolVar(&flagSearch, "search", false, "search lists all pkgs that match pkg arg")
	flag.BoolVar(&flagSearchv, "searchv", false, "like -search but with verbose output")
	flag.BoolVar(&flagKillAll, "killall", false, "kill any running godoc http servers")
	flag.BoolVar(&flagDebug, "debug", false, "print debug messages")
}

func main() {
	if runtime.GOOS != "darwin" {
		exitOnErr(fmt.Errorf("gohdoc only works on macOS"))
	}

	flag.Parse()

	if flagDebug == false {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetFlags(log.Ltime | log.Lshortfile)
	}

	gopath = os.Getenv("GOPATH")
	if gopath == "" {
		gopath = build.Default.GOPATH
	}
	log.Println("using GOPATH:", gopath)

	if flagHelp {
		cmdHelp()
		os.Exit(0)
	}

	ctx = context.Background()
	var cancelFn context.CancelFunc
	ctx, cancelFn = context.WithCancel(ctx)

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, os.Kill)

		<-stop
		log.Println("received interrupt/kill signal")
		cancelFn()
	}()

	if flagKillAll {
		err := cmdKillAll()
		exitOnErr(err)
		os.Exit(0)
	}

	if flagList || flagListv {
		err := cmdList()
		exitOnErr(err)
		os.Exit(0)
	}

	if flagSearch || flagSearchv {
		err := cmdSearch()
		exitOnErr(err)
		os.Exit(0)
	}

	err := cmdOpen()
	exitOnErr(err)

}

// cmdHelp prints help.
func cmdHelp() {
	const helpText = `gohdoc (go html doc) opens godoc for a pkg in the browser. gohdoc looks for an existing
godoc http server, and uses that if available. If not, gohdoc will start a godoc http
server on port 6060 (port can be overridden using envar GODOC_HTTP_PORT). The godoc http
server will continue to run after gohdoc exits, but can be killed using gohdoc -killall.

Usage:

  gohdoc                                       open godoc in browser for current pkg
  gohdoc .                                     same as above
  gohdoc my/sub/pkg                            open godoc in browser for pkg at relative path
  gohdoc ~/go/src/github.com/ksoze/myproj      open godoc in browser for pkg at this path
  gohdoc fmt                                   open godoc in browser for pkg fmt
  gohdoc -help                                 print this help message
  gohdoc -killall                              kill any running godoc http servers
  gohdoc -list                                 list all packages on the godoc http server
  gohdoc -listv                                same as -list, but prints additional detail
  gohdoc -search pkg/name                      list packages that match arg
  gohdoc -searchv pkg/name                     same as -search, but prints additional detail


The -debug flag can be used to enable debug logging. If gohdoc spawns a godoc
http server, the -debug flag will also print that server's verbose output.

Note that the godoc http server is tied to a particular GOPATH. If your pkg is not
found, verify that the godoc http server is using the correct GOPATH. If necessary,
use gohdoc -killall and rerun gohdoc inside the appropriate GOPATH.`

	fmt.Printf("gohdoc version %s - neilotoole@apache.org\n\n", Version)
	fmt.Println(helpText)
}
