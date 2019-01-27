This `testdata` dir contains output from the godoc http server's `/pkg/` path,
for several recent go versions.

Recreated the output for each go version (using the appropriate image, e.g. `golang:1.11`, `golang:1.10`, etc):


```
docker run -v $(pwd):/go/src/github.com/neilotoole/gohdoc -w /go/src/github.com/neilotoole/gohdoc -p 6060:6060 -it golang:1.11 godoc -index -http :6060

```

And then access the package page at `http://localhost:6060/pkg`.

Note that we mount the `gohdoc` src into the container, so that `gohdoc` should show up in the package list.


