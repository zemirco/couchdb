package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-querystring/query"
)

// View performs actions and certain view documents
type View struct {
	URL string
	*Database
}

// Get executes specified view function from specified design document.
func (v *View) Get(name string, params QueryParameters) (*ViewResponse, error) {
	q, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s_view/%s?%s", v.URL, name, q.Encode())
	res, err := v.Database.Client.Request(http.MethodGet, uri, nil, "")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return newViewResponse(res.Body)
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
	res, err := v.Database.Client.Request(http.MethodPost, url, &b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	return newViewResponse(res.Body)
}

func newViewResponse(body io.Reader) (*ViewResponse, error) {
	response := &ViewResponse{}
	return response, json.NewDecoder(body).Decode(&response)
}
