package couchdb

import (
  "net/http"
  "io/ioutil"
  "encoding/json"
  "io"
)

// Get server information.
func (c *Client) Info() (*Server, error) {
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

// Get all databases.
func (c *Client) All() ([]string, error) {
  body, err := request("GET", c.Url + "_all_dbs", nil)
  if err != nil {
    return nil, err
  }
  var data []string
  return data, json.Unmarshal(body, &data)
}

// Get database.
func (c *Client) Get(name string) (*DatabaseInfo, error) {
  body, err := request("GET", c.Url + name, nil)
  if err != nil {
    return nil, err
  }
  var dbInfo *DatabaseInfo
  return dbInfo, json.Unmarshal(body, &dbInfo)
}

// Create database.
func (c *Client) Create(name string) (*DbResponse, error) {
  body, err := request("PUT", c.Url + name, nil)
  if err != nil {
    return nil, err
  }
  var DbResponse *DbResponse
  return DbResponse, json.Unmarshal(body, &DbResponse)
}

// Delete database.
func (c *Client) Delete(name string) (*DbResponse, error) {
  body, err := request("DELETE", c.Url + name, nil)
  if err != nil {
    return nil, err
  }
  var DbResponse *DbResponse
  return DbResponse, json.Unmarshal(body, &DbResponse)
}

// Use database.
func (c *Client) Use(name string) (Database) {
  return Database{c.Url + name + "/"}
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
