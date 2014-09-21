
# go-relax

[![Build Status](https://travis-ci.org/zemirco/go-relax.svg)](https://travis-ci.org/zemirco/go-relax)

CouchDB client for Go.

Check out the [docs](https://godoc.org/github.com/zemirco/go-relax).

## Example

```go
package main

import "github.com/zemirco/couchdb"

func check(err error) {
  if err != nil {
    panic(err)
  }
}

func main() {
  client, err := NewClient("http://127.0.0.1:5984/")
  check(err)

  info, err := client.Info()
  check(err)
}
```

## Test

`go test`

## License

MIT
