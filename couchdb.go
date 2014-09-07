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

type Error struct {
  Method string
  Url string
  StatusCode int
  Type string `json:"error"`
  Reason string
}

// custom Error struct has to implement Error method
func (e *Error) Error() string {
  return fmt.Sprintf("CouchDB - %s %s, Status Code: %d, Error: %s, Reason: %s", e.Method, e.Url, e.StatusCode, e.Type, e.Reason)
}

// leave out _rev when empty otherwise "Invalid rev format"
type Document struct {
  Id string `json:"_id,omitempty"`
  Rev string `json:"_rev,omitempty"`
}

type DocResponse struct {
  Ok bool
  Id string
  Rev string
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
    fmt.Println("here")
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
  return Database{c.Url + name + "/"}
}



// DATABASE OPERATIONS



/**
 * Head request.
 * http://docs.couchdb.org/en/latest/api/document/common.html#head--db-docid
 */
func (db *Database) head(id string) (*http.Response, error) {
  return http.Head(db.Url + id)
}

func (db *Database) get(document interface{}, id string) error {
  body, err := request("GET", db.Url + id, nil)
  if err != nil {
    return err
  }
  return json.Unmarshal(body, &document)
}

func (db *Database) put(doc interface{}) error {
  res, err := json.Marshal(doc)
  if err != nil {
    return err
  }
  var document *Document
  err = json.Unmarshal(res, &document)
  if err != nil {
    return err
  }
  data := bytes.NewReader(res)
  if err != nil {
    return err
  }
  _, err = request("PUT", db.Url + document.Id, data)
  if err != nil {
    return err
  }
  return db.get(doc, document.Id)
}

func (db *Database) post(doc interface{}) error {
  res, err := json.Marshal(doc)
  if err != nil {
    return err
  }
  data := bytes.NewReader(res)
  if err != nil {
    return err
  }
  body, err := request("POST", db.Url, data)
  if err != nil {
    return err
  }
  var response *DocResponse
  err = json.Unmarshal(body, &response)
  if err != nil {
    return err
  }
  return db.get(doc, response.Id)
}

// FUNC MAIN



func main() {

  const url = "http://127.0.0.1:5984/"

  // create client
  client := Client{url}

  // get server info
  // couch, err := client.info()
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(couch)

  // get all dbs
  // res, err := client.all()
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(res)

  // get db information
  // info, err := client.get("_users")
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(info)

  // use db
  db := client.use("_users")

  // get document head
  head, err := db.head("_design/_auth")
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(head.StatusCode)

  // type MyDoc struct {
  //   Document
  //   Brand string `json:"brand"`
  //   Name string
  //   Nested struct {
  //     Awesome string
  //   }
  // }
  //
  // // create document
  // myDoc := MyDoc{
  //   Brand: "volvo",
  //   Name: "Pickup",
  // }
  //
  // err := db.post(&myDoc)
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(myDoc.Name)

  // get document
  // var myDoc *MyDoc
  // err := db.get(&myDoc, "awesome")
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(myDoc)
  //
  // myDoc.Name = "sour"
  // err = db.put(&myDoc)
  // if err != nil {
  //   log.Fatal(err)
  // }
  // fmt.Println(myDoc)

  // doc["foo"] = "bar"
  // nested := doc["nested"].(map[string]interface{})
  // fmt.Println(nested["awesome"])



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

// HELPER FUNCTIONS
func request(method, url string, data io.Reader) ([]byte, error) {
  client := &http.Client{}
  req, err := http.NewRequest(method, url, data)
  // for post request
  req.Header.Set("Content-Type", "application/json")
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
  if res.StatusCode < 200 || res.StatusCode >= 300 {
    var error *Error
    err = json.Unmarshal(body, &error)
    if err != nil {
      return nil, err
    }
    if error.Type != "" && error.Reason != "" {
      error.Method = method
      error.Url = url
      error.StatusCode = res.StatusCode
      return nil, error
    }
  }
  return body, nil
}

// func marshal(v interface{}) (io.Reader, error) {
//   res, err := json.Marshal(v)
//   var document *Document
//   json.Unmarshal(res, &document)
//   fmt.Println(document.Id)
//   if err != nil {
//     return nil, err
//   }
//   return bytes.NewReader(res), nil
// }
