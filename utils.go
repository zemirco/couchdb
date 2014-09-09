package couchdb

import (
  "path/filepath"
  "mime"
  "io"
  "net/http"
  "io/ioutil"
  "encoding/json"
)

// Get mime type from file name.
func mimeType(name string) string {
  ext := filepath.Ext(name)
  return mime.TypeByExtension(ext)
}

// Make HTTP request.
// Treat status code other than 2xx as Error.
func request(method, url string, data io.Reader) ([]byte, error) {
  req, err := http.NewRequest(method, url, data)
  if err != nil {
    return nil, err
  }
  req.Header.Set("Content-Type", "application/json")
  client := &http.Client{}
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
    return nil, newError(res, body)
  }
  return body, nil
}

// Convert HTTP response from CouchDB into Error.
func newError(res *http.Response, body []byte) error {
  var error *Error
  err := json.Unmarshal(body, &error)
  if err != nil {
    return err
  }
  error.Method = res.Request.Method
  error.Url = res.Request.URL.String()
  error.StatusCode = res.StatusCode
  return error
}
