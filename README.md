# gohdoc

`gohdoc` opens a package's godoc in the browser.

[![Go Report Card](https://goreportcard.com/badge/github.com/neilotoole/gohdoc)](https://goreportcard.com/report/github.com/neilotoole/gohdoc)

## Overview

In your package dir, execute `gohdoc .` (or just `gohdoc`). This will open the current package's
godoc in the browser, starting a godoc http server if necessary. You can  specify absolute and
relative paths, or full or partial package names, e.g. `gohdoc fmt` or `gohdoc encoding/jso`.
Fragments are preserved, so `gohdoc fmt#Println` will work. Use `gohdoc -search` or `gohdoc -list`
to interrogate the set of packages on the godoc server.

## Install

`gohdoc` is installed in the usual Go fashion:

```bash

$ go get -u github.com/neilotoole/gohdoc
```


## Usage

Use `gohdoc -help` to see this:

```
gohdoc opens a package's godoc in the browser.

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
```


