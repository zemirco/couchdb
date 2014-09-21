
# couchdb

[![Build Status](https://travis-ci.org/zemirco/couchdb.svg)](https://travis-ci.org/zemirco/couchdb)
[![GoDoc](https://godoc.org/github.com/zemirco/couchdb?status.svg)](https://godoc.org/github.com/zemirco/couchdb)

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

More
[examples](https://github.com/zemirco/couchdb/blob/master/example/example.go).

## Test

`go test`

## License

MIT
