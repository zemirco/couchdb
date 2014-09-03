package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
  "log"
  "encoding/json"
)

type CouchDB struct {
  Uuid string
  Version string
}

type Database struct {
  DbName string `json:"db_name"`
  DocCount int `json:"doc_count"`
  DocDelCount int `json:"doc_del_count"`
}

type Result struct {
  Ok bool
  Error string
  Reason string
}

type Client struct {
  Url string
}

/**
 * Get all databases.
 */
func (c *Client) all() ([]string, error) {
  res, err := http.Get(c.Url + "_all_dbs")
  if err != nil {
    return nil, err
  }
  defer res.Body.Close()
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return nil, err
  }
  var data []string
  err = json.Unmarshal(body, &data)
  if err != nil {
    return nil, err
  }
  return data, nil
}

/**
 * Get single database.
 */
func (c *Client) get(name string) (*Database, error) {
  res, err := http.Get(c.Url + name)
  if err != nil {
    return nil, err
  }
  defer res.Body.Close()
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return nil, err
  }
  var database *Database
  err = json.Unmarshal(body, &database)
  if err != nil {
    return nil, err
  }
  return database, nil
}

/**
 * Create single database.
 */
func (c *Client) create(name string) (*Result, error) {
  client := &http.Client{}
  req, err := http.NewRequest("PUT", c.Url + name, nil)
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
  var result *Result
  err = json.Unmarshal(body, &result)
  if err != nil {
    return nil, err
  }
  return result, nil
}


func main() {

  const url = "http://127.0.0.1:5984/"

  // create client
  client := &Client{url}

  // get all dbs
  res, err := client.all()
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(res)

  // get dbs
  db, err := client.get("_users")
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(db)

  // create db
  status, err := client.create("awesome")
  if err != nil {
    log.Fatal(err)
  }
  fmt.Println(status)
  fmt.Println(status.Ok)
}
