
# go-relax

CouchDB client written in Go.

work in progress ...

## Usage

Create a new client.

```go
client := Client{"http://127.0.0.1:5984/"}
```

## Structs

##### Server

```
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

Returns struct of type Server.

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

Returns a slice of strings.

```go
res, err := client.all()
// [_replicator _users]
```

## License

MIT
