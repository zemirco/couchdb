
# go-relax

[![Build Status](https://travis-ci.org/zemirco/go-relax.svg)](https://travis-ci.org/zemirco/go-relax)
[![GoDoc](https://godoc.org/github.com/zemirco/go-relax?status.svg)](https://godoc.org/github.com/zemirco/go-relax)

CouchDB client for Go.

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

  // create a new client
  client, err := couchdb.NewClient("http://127.0.0.1:5984/")
  check(err)

  // get some information about your CouchDB
  info, err := client.Info()
  check(err)
  fmt.Println(info)

}
```

For more see
[example/example.go](https://github.com/zemirco/go-relax/blob/master/example/example.go).

## Test

`go test`

## License

MIT
