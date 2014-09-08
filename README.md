
# go-relax

[![Build Status](https://travis-ci.org/zemirco/go-relax.svg)](https://travis-ci.org/zemirco/go-relax)

CouchDB client written in Go.

work in progress ...

## Usage

Create a new client.

```go
client := Client{"http://127.0.0.1:5984/"}
```

## Structs

##### Server

```go
type Server struct {
  Couchdb string
  Uuid string
  Vendor struct {
    Version string
    Name string
  }
  Version string
}
```

## [Server](http://docs.couchdb.org/en/latest/api/server/index.html)

##### [GET /](http://docs.couchdb.org/en/latest/api/server/common.html#get--)

Returns [`Server`](#server).

```go
couch, err := client.info()
// {
//   Welcome
//   6d4ef59395b6b2285fe12de8dd7af3a7
//   {
//     1.5.0
//     The Apache Software Foundation
//   }
//   1.5.0
// }
```

##### [GET /_all_dbs](http://docs.couchdb.org/en/latest/api/server/common.html#all-dbs)

Returns `[]string`.

```go
res, err := client.all()
// [_replicator _users]
```

## [Documents](http://docs.couchdb.org/en/latest/api/document/common.html)

Take your client and use a database.

```go
db := client.use("_users")
```

##### [HEAD /docid](http://docs.couchdb.org/en/latest/api/document/common.html#head--db-docid)

```go
head, err := db.head("_design/_auth")
```

Returns `http.Response`.

## Test

`go test`

## License

MIT
