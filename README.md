
# go-relax

[![Build Status](https://travis-ci.org/zemirco/go-relax.svg)](https://travis-ci.org/zemirco/go-relax)

CouchDB client written in Go.

work in progress ...

Check out the [docs](https://godoc.org/github.com/zemirco/go-relax).

## Example

```go
package main

import "github.com/zemirco/couchdb"

func main() {
  client := couchdb.Client{"http://127.0.0.1:5984/"}

  info, err := client.Info()
  if err != nil {
    panic(err)
  }
}
```

## Test

`go test`

## License

MIT
