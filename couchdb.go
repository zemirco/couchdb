package main

import (
  "bytes"
  "fmt"
  "net/http"
  "io/ioutil"
  "log"
  "encoding/json"
  "io"
  // "errors"
)



// STRUCTS



type Client struct {
  Url string
}

// http://docs.couchdb.org/en/latest/intro/api.html#server
type Server struct {
  Couchdb string
  Uuid string
  Vendor struct {
    Version string
    Name string
  }
  Version string
}

type Database struct {
  Url string
}

type DatabaseInfo struct {
  Name string `json:"db_name"`
  DocCount int `json:"doc_count"`
  DocDelCount int `json:"doc_del_count"`
}

type DbResponse struct {
  Ok bool
  Error string
  Reason string
}

type DocResponse struct {
  Ok bool
  Id string
  Rev string
  Error string
  Reason string
}

type Error struct {
  Method string
  Url string
  StatusCode int
  Type string `json:"error"`
  Reason string
}

// custom Error struct has to implement Error method
func (e *Error) Error() string {
  return "CouchDB: " + e.Type + " - " + e.Reason
}

// CLIENT OPERATIONS



/**
 * Get server information.
 */
func (c *Client) info() (*Server, error) {
  body, err := request("GET", c.Url, nil)
  if err != nil {
    return nil, err
  }
  var server *Server
  err = json.Unmarshal(body, &server)
  if err != nil {
    return nil, err
  }
  return server, nil
}

/**
 * Get all databases.
 */
func (c *Client) all() ([]string, error) {
  body, err := request("GET", c.Url + "_all_dbs", nil)
  if err != nil {
    return nil, err
  }
  var data []string
  return data, json.Unmarshal(body, &data)
}

/**
 * Get single database.
 */
func (c *Client) get(name string) (*DatabaseInfo, error) {
  body, err := request("GET", c.Url + name, nil)
  if err != nil {
    return nil, err
  }
  var dbInfo *DatabaseInfo
  return dbInfo, json.Unmarshal(body, &dbInfo)
}

/**
 * Create single database.
 */
func (c *Client) create(name string) (*DbResponse, error) {
  body, err := request("PUT", c.Url + name, nil)
  if err != nil {
    return nil, err
  }
  var DbResponse *DbResponse
  return DbResponse, json.Unmarshal(body, &DbResponse)
}

/**
 * Delete single database.
 */
func (c *Client) delete(name string) (*DbResponse, error) {
  body, err := request("DELETE", c.Url + name, nil)
  if err != nil {
    return nil, err
  }
  var DbResponse *DbResponse
  return DbResponse, json.Unmarshal(body, &DbResponse)
}

func (c *Client) use(name string) (Database) {
  return Database{c.Url + "/" + name + "/"}
}



// DATABASE OPERATIONS



/**
 * Head request.
 * http://docs.couchdb.org/en/latest/api/document/common.html#head--db-docid
 */
func (db *Database) head(id string) (*http.Response, error) {
  return http.Head(db.Url + id)
}

func (db *Database) get(id string) (map[string]interface{}, error) {
  body, err := request("GET", db.Url + id, nil)
  if err != nil {
    return nil, err
  }
  var data map[string]interface{}
  return data, json.Unmarshal(body, &data)
}

func (db *Database) put(id string, document interface{}) (*DocResponse, error) {
  data, err := marshal(document)
  if err != nil {
    return nil, err
  }
  body, err := request("PUT", db.Url + id, data)
  if err != nil {
    return nil, err
  }
  var res *DocResponse
  return res, json.Unmarshal(body, &res)
}


// FUNC MAIN



func main() {

  const url = "http://127.0.0.1:5984/"

  // create client
  client := &Client{url}

  // get server info
  // couch, err := client.info()
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(couch.Vendor.Version)

  // get all dbs
  // res, err := client.all()
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(res)

  // get db information
  // info, err := client.get("nice")
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(info)

  // use db
  db := client.use("nice")

  // get document head
  // head, err := db.head("awesome")
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(head.StatusCode)

  // get document
  // doc, err := db.get("awesome")
  // if err != nil {
  //   log.Fatal(err)
  // }
  // nested := doc["nested"].(map[string]interface{})
  // fmt.Println(nested["awesome"])

  // put document
  type MyDoc struct {
    Brand string `json:"brand"`
  }
  myDoc := MyDoc{"audi"}
  _, err := db.put("tight", myDoc)
  if err != nil {
    fmt.Println(err.Type)
    log.Fatal(err)
  }




  //
  // // create db
  // status, err := client.create("nice")
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(status)
  // fmt.Println(status.Ok)
  //
  // // delete database
  // status, err = client.delete("awesome")
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(status)
}

// PROBLEM
// return normal error vs return custom Error struct

// HELPER FUNCTIONS
func request(method, url string, data io.Reader) ([]byte, error) {
  client := &http.Client{}
  req, err := http.NewRequest(method, url, data)
  if err != nil {
    return nil, err
  }
  res, err := client.Do(req)
  if err != nil {
    return nil, err
  }
  defer res.Body.Close()
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return nil, err
  }
  // handle CouchDB http errors
  var error *Error
  err = json.Unmarshal(body, &error)
  if err != nil {
    return nil, err
  }
  if error.Type != "" && error.Reason != "" {
    fmt.Println(method)
    error.Method = method
    error.Url = url
    error.StatusCode = res.StatusCode
    return nil, error
  }
  return body, nil
}

func marshal(v interface{}) (io.Reader, error) {
  json, err := json.Marshal(v)
  if err != nil {
    return nil, err
  }
  return bytes.NewReader(json), nil
}
