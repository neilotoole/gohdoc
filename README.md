# gohdoc
gohdoc opens a package's godoc in the browser.

## Overview

In your Go package, execute `gohdoc .` (or just `gohdoc`). This will open the current package's
godoc in the browser. You can also specify absolute and relative paths, or full or partial
package names, e.g. `gohdoc fmt` or `gohdoc encoding/jso`. See `gohdoc -help` for more.

## Install

`gohdoc` is installed in the usual fashion:

```bash

$ go get -u github.com/neilotoole/gohdoc
```


## Help

Use `gohdoc -help` to see this:

```
gohdoc 0.1 - https://github.com/neilotoole/gohdoc

gohdoc (go http doc) opens godoc for a pkg in the browser. gohdoc looks for an existing
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
  gohdoc -servers                              list godoc http server processes
  gohdoc -killall                              kill all godoc http server processes
  gohdoc -list                                 list all packages on the godoc http server
  gohdoc -listv                                same as -list, but prints additional detail
  gohdoc -search pkg/name                      list packages that match arg
  gohdoc -searchv pkg/name                     same as -search, but prints additional detail


The -debug flag can be used to enable debug logging. If gohdoc spawns a godoc
http server, the -debug flag will also print that server's verbose output.

Note that the godoc http server is tied to a particular GOPATH. If your pkg is not
found, verify that the godoc http server is using the correct GOPATH. If necessary,
use gohdoc -killall and rerun gohdoc inside the appropriate GOPATH.

```


