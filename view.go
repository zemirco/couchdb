package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/go-querystring/query"
)

// Execute specified view function from specified design document.
func (v *View) Get(name string, params QueryParameters) (*ViewResponse, error) {
	q, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s_view/%s?%s", v.Url, name, q.Encode())
	body, err := request("GET", url, nil, "")
	if err != nil {
		return nil, err
	}
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
	body, err := request("GET", url, data, "application/json")
	if err != nil {
		return nil, err
	}
	return newViewResponse(body)
}

func newViewResponse(body []byte) (*ViewResponse, error) {
	var response *ViewResponse
	return response, json.Unmarshal(body, &response)
}
