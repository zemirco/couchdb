package couchdb

import (
  "net/http"
  "encoding/json"
  "bytes"
)

type Database struct {
  Url string
}

/**
 * Head request.
 */

func (db *Database) Head(id string) (*http.Response, error) {
  return http.Head(db.Url + id)
}

/**
 * Get document
 */
func (db *Database) Get(doc CouchDoc, id string) error {
  body, err := request("GET", db.Url + id, nil)
  if err != nil {
    return err
  }
  return json.Unmarshal(body, doc)
}

/**
 * Put document.
 */
func (db *Database) Put(doc CouchDoc) (*DocumentResponse, error) {
  res, err := json.Marshal(doc)
  if err != nil {
    return nil, err
  }
  id := doc.GetId()
  body, err := request("PUT", db.Url + id, bytes.NewReader(res))
  if err != nil {
    return nil, err
  }
  var response *DocumentResponse
  return response, json.Unmarshal(body, &response)
}

/**
 * Post document.
 */
func (db *Database) Post(doc CouchDoc) (*DocumentResponse, error) {
  res, err := json.Marshal(doc)
  if err != nil {
    return nil, err
  }
  body, err := request("POST", db.Url, bytes.NewReader(res))
  if err != nil {
    return nil, err
  }
  var response *DocumentResponse
  return response, json.Unmarshal(body, &response)
}

/**
 * Delete document.
 */
func (db *Database) Delete(doc CouchDoc) (*DocumentResponse, error) {
  id := doc.GetId()
  rev := doc.GetRev()
  body, err := request("DELETE", db.Url + id + "?rev=" + rev, nil)
  if err != nil {
    return nil, err
  }
  var response *DocumentResponse
  return response, json.Unmarshal(body, &response)
}
