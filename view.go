package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
	"io"
)

type View struct {
	Url string
	*Database
}

// Execute specified view function from specified design document.
func (v *View) Get(name string, params QueryParameters) (*ViewResponse, error) {
	q, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	uri := fmt.Sprintf("%s_view/%s?%s", v.Url, name, q.Encode())
	body, err := v.Database.Client.request("GET", uri, nil, "")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newViewResponse(body)
}

// Execute specified view function from specified design document.
// Unlike View.Get for accessing views, View.Post supports
// the specification of explicit keys to be retrieved from the view results.
func (v *View) Post(name string, keys []string, params QueryParameters) (*ViewResponse, error) {
	// create POST body
	res, err := json.Marshal(keys)
	if err != nil {
		return nil, err
	}
	// create query string
	q, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s_view/%s?%s", v.Url, name, q.Encode())
	data := bytes.NewReader(res)
	body, err := v.Database.Client.request("GET", url, data, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newViewResponse(body)
}

func newViewResponse(body io.ReadCloser) (*ViewResponse, error) {
	response := &ViewResponse{}
	return response, json.NewDecoder(body).Decode(&response)
}
