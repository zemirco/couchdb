package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/go-querystring/query"
)

// ViewService is an interface for dealing with a view inside a CouchDB database.
type ViewService interface {
	Get(name string, params QueryParameters) (*ViewResponse, error)
	Post(name string, keys []string, params QueryParameters) (*ViewResponse, error)
}

// View performs actions and certain view documents
type View struct {
	URL    string
	Client *Client
}

// Get executes specified view function from specified design document.
func (v *View) Get(name string, params QueryParameters) (*ViewResponse, error) {
	q, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s_view/%s?%s", v.URL, name, q.Encode())
	res, err := v.Client.Request(http.MethodGet, uri, nil, "")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response ViewResponse
	return &response, json.NewDecoder(res.Body).Decode(&response)
}

// Post executes specified view function from specified design document.
// Unlike View.Get for accessing views, View.Post supports
// the specification of explicit keys to be retrieved from the view results.
func (v *View) Post(name string, keys []string, params QueryParameters) (*ViewResponse, error) {
	content := struct {
		Keys []string `json:"keys"`
	}{
		Keys: keys,
	}
	// create POST body
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(content); err != nil {
		return nil, err
	}
	// create query string
	q, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s_view/%s?%s", v.URL, name, q.Encode())
	res, err := v.Client.Request(http.MethodPost, url, &b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response ViewResponse
	return &response, json.NewDecoder(res.Body).Decode(&response)
}
