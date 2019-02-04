package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
)

func main() {
	app := newDefaultApp()
	err := initApp(app)
	exitOnErr(app, err)

	log.Printf("args: %s", strings.Join(os.Args, " "))

	switch {
	case app.flagHelp:
		cmdHelp()
		return
	case app.flagVersion:
		cmdVersion()
		return
	case app.flagServers:
		err = cmdServers(app)
	case app.flagKillAll:
		err = cmdKillAll(app)
	case app.flagList, app.flagListv:
		err = cmdList(app)
	case app.flagSearch, app.flagSearchv:
		err = cmdSearch(app)
	default:
		err = cmdOpen(app)
	}

	exitOnErr(app, err)
}

const (
	// envGodocPort is the envar used to override the
	// default godoc http server port (6060).
	envGodocPort = "GODOC_HTTP_PORT"
	version      = "0.1.2"
	helpText     = `gohdoc opens a package's godoc in the browser.

gohdoc (go http doc) looks for an existing godoc http server, and uses that if
available. If not, gohdoc will start a godoc http server on port 6060; override
with envar GODOC_HTTP_PORT. The godoc http server will continue to run after
gohdoc exits, but can be killed using gohdoc -killall.


Usage:

  gohdoc                                   open current pkg godoc in browser
  gohdoc .                                 same as above
  gohdoc my/sub/pkg                            
  gohdoc /go/src/github.com/my/pkg      
  gohdoc fmt                                   
  gohdoc fmt#Println                       open fmt#Println godoc
  gohdoc #MyFunc                           open current pkg #MyFunc godoc


Interrogate the godoc server's package list:

  gohdoc -list                             list all packages on the godoc http server
  gohdoc -listv                            same as -list, but also print pkg url
  gohdoc -search pkg/name                  list packages that match arg
  gohdoc -searchv pkg/name                 same as -search, but also print pkg url


List or kill running godoc servers:

  gohdoc -servers                          list godoc http server processes
  gohdoc -killall                          kill all godoc http server processes


For completeness:

  gohdoc -help                             print this help message
  gohdoc -version                          print gohdoc version


The -debug flag can be used to enable debug logging. If gohdoc spawns a godoc
http server, the -debug flag will also print that server's verbose output.

Note that a godoc http server is tied to a particular GOPATH. If your pkg is
unexpectedly not found, verify that the godoc http server is started on the
correct GOPATH. If necessary, use gohdoc -killall and rerun gohdoc inside the
appropriate GOPATH.

Feedback, bug reports etc to https://github.com/neilotoole/gohdoc
gohdoc was created by Neil O'Toole and is released under the MIT License.
`
)

// cmdHelp prints help.
func cmdHelp() {
	fmt.Println(helpText)
}

func cmdVersion() {
	fmt.Printf("gohdoc v%s\n", version)
}

// App holds program state.
type App struct {
	// port is the port to start a godoc http server on, if necessary to do so. Defaults
	// to 6060, but can be overridden by envar GODOC_HTTP_PORT.
	port int

	// cwd is the current working directory
	cwd string
	// cmd is the Cmd used to start a godoc http server, if necessary to do so.
	cmd *exec.Cmd
	ctx context.Context
	// serverPkgPageBody holds the contents of the godoc http server's /pkg/ page
	serverPkgPageBody []byte
	// serverPkgList holds the list of pkgs parsed from serverPkgPageBody
	serverPkgList []string

	flagHelp    bool
	flagVersion bool
	flagList    bool
	flagListv   bool
	flagSearch  bool
	flagSearchv bool
	flagServers bool
	flagKillAll bool

	flagDebug bool

	// args holds the processed value of flag.Args after flag.Parse is invoked.
	// Each element of args will have whitespace trimmed.
	args []string
}

// newDefaultApp returns a default App instance.
func newDefaultApp() *App {
	app := &App{port: 6060, ctx: context.Background()}

	var err error
	app.cwd, err = os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("failed to get current working directory: %v", err))
	}
	return app
}

// initApp initializes an App.
func initApp(app *App) error {
	flag.BoolVar(&app.flagHelp, "help", false, "print help")
	flag.BoolVar(&app.flagList, "list", false, "list all pkgs from godoc http server")
	flag.BoolVar(&app.flagListv, "listv", false, "like -list but with verbose output")
	flag.BoolVar(&app.flagSearch, "search", false, "search lists all pkgs that match pkg arg")
	flag.BoolVar(&app.flagSearchv, "searchv", false, "like -search but with verbose output")
	flag.BoolVar(&app.flagServers, "servers", false, "list all godoc http server processes")
	flag.BoolVar(&app.flagKillAll, "killall", false, "kill all godoc http server processes")
	flag.BoolVar(&app.flagDebug, "debug", false, "print debug messages")
	flag.BoolVar(&app.flagVersion, "version", false, "print gohdoc version")

	flag.Parse()

	for _, arg := range flag.Args() {
		// Process command line args.
		// Each element of app.args will have whitespace trimmed (and no empty strings).
		arg = strings.TrimSpace(arg)
		if len(arg) > 0 {
			app.args = append(app.args, arg)
		}
	}

	if !app.flagDebug {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetFlags(log.Ltime | log.Lshortfile)
	}

	envPortVal, ok := os.LookupEnv(envGodocPort)
	if ok {
		log.Printf("found envar %s: %s", envGodocPort, envPortVal)
	}
	if ok && len(strings.TrimSpace(envPortVal)) > 0 {
		var err error
		app.port, err = strconv.Atoi(envPortVal)
		if err != nil || app.port < 1 || app.port > 65535 {
			return fmt.Errorf("%s was set, but value is invalid: %s", envGodocPort, envPortVal)
		}
	}

	var cancelFn context.CancelFunc
	app.ctx, cancelFn = context.WithCancel(app.ctx)

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)

		<-stop
		log.Println("received interrupt/kill signal")
		cancelFn()
		if app.cmd != nil {
			_ = app.cmd.Process.Kill()
		}
	}()
	return nil
}

// exitOnErr prints err info and calls os.Exit(1) if err is not nil.
func exitOnErr(app *App, err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		if app != nil && app.cmd != nil {
			// If we're exiting due to an error, and we already started a godoc http server, kill it
			log.Printf("killing the godoc http server [%d] that gohdoc started\n", app.cmd.Process.Pid)
			_ = app.cmd.Process.Kill()
		}
		os.Exit(1)
	}
}
