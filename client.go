package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// Client holds all info for database client
type Client struct {
	Username  string
	Password  string
	URL       string
	CookieJar *cookiejar.Jar
}

// NewClient returns new couchdb client for given url
func NewClient(url string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Client{"", "", url, jar}, nil
}

// NewAuthClient returns new couchdb client with basic authentication
func NewAuthClient(username, password, url string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Client{username, password, url, jar}, nil
}

// Info returns some information about the server
func (c *Client) Info() (*Server, error) {
	body, err := c.request(http.MethodGet, c.URL, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	server := &Server{}
	return server, json.NewDecoder(body).Decode(&server)
}

// ActiveTasks returns list of currently running tasks
func (c *Client) ActiveTasks() ([]Task, error) {
	url := fmt.Sprintf("%s_active_tasks", c.URL)
	body, err := c.request(http.MethodGet, url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	tasks := []Task{}
	return tasks, json.NewDecoder(body).Decode(&tasks)
}

// All returns list of all databases on server
func (c *Client) All() ([]string, error) {
	url := fmt.Sprintf("%s_all_dbs", c.URL)
	body, err := c.request(http.MethodGet, url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	data := []string{}
	return data, json.NewDecoder(body).Decode(&data)
}

// Get database.
func (c *Client) Get(name string) (*DatabaseInfo, error) {
	url := fmt.Sprintf("%s%s", c.URL, name)
	body, err := c.request(http.MethodGet, url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	dbInfo := &DatabaseInfo{}
	return dbInfo, json.NewDecoder(body).Decode(&dbInfo)
}

// Create database.
func (c *Client) Create(name string) (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s%s", c.URL, name)
	body, err := c.request(http.MethodPut, url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDatabaseResponse(body)
}

// Delete database.
func (c *Client) Delete(name string) (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s%s", c.URL, name)
	body, err := c.request(http.MethodDelete, url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDatabaseResponse(body)
}

// CreateUser creates a new user in _users database
func (c *Client) CreateUser(user User) (*DocumentResponse, error) {
	url := fmt.Sprintf("%s_users/%s", c.URL, user.ID)
	res, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(res)
	body, err := c.request(http.MethodPut, url, data, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDocumentResponse(body)
}

// GetUser returns user by given name
func (c *Client) GetUser(name string) (*User, error) {
	url := fmt.Sprintf("%s_users/org.couchdb.user:%s", c.URL, name)
	body, err := c.request(http.MethodGet, url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	user := &User{}
	return user, json.NewDecoder(body).Decode(&user)
}

// DeleteUser removes user from database
func (c *Client) DeleteUser(user *User) (*DocumentResponse, error) {
	db := c.Use("_users")
	return db.Delete(user)
}

// CreateSession creates a new session and logs in user
func (c *Client) CreateSession(name, password string) (*PostSessionResponse, error) {
	url := fmt.Sprintf("%s_session", c.URL)
	creds := Credentials{name, password}
	res, err := json.Marshal(creds)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(res)
	body, err := c.request(http.MethodPost, url, data, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	sessionResponse := &PostSessionResponse{}
	return sessionResponse, json.NewDecoder(body).Decode(&sessionResponse)
}

// GetSession returns session for currently logged in user
func (c *Client) GetSession() (*GetSessionResponse, error) {
	url := fmt.Sprintf("%s_session", c.URL)
	body, err := c.request(http.MethodGet, url, nil, "")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	sessionResponse := &GetSessionResponse{}
	return sessionResponse, json.NewDecoder(body).Decode(&sessionResponse)
}

// DeleteSession removes current session and logs out user
func (c *Client) DeleteSession() (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s_session", c.URL)
	body, err := c.request(http.MethodDelete, url, nil, "")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	databaseResponse := &DatabaseResponse{}
	return databaseResponse, json.NewDecoder(body).Decode(&databaseResponse)
}

// Use database.
func (c *Client) Use(name string) Database {
	return Database{
		URL:    c.URL + name + "/",
		Client: c,
	}
}

// ReplicationRequest is JSON object for post request to _replicate URL.
//
// http://docs.couchdb.org/en/1.6.1/api/server/common.html#replicate
type ReplicationRequest struct {
	Document
	Cancel       bool              `json:"cancel,omitempty"`
	Continuous   bool              `json:"continuous,omitempty"`
	CreateTarget bool              `json:"create_target,omitempty"`
	DocIDs       []string          `json:"doc_ids,omitempty"`
	Proxy        string            `json:"proxy,omitempty"`
	Source       string            `json:"source,omitempty"`
	Target       string            `json:"target,omitempty"`
	Filter       string            `json:"filter,omitempty"`
	QueryParams  map[string]string `json:"query_params,omitempty"`
}

// ReplicationResponse is JSON object for response from post request to _replicate URL.
//
// http://docs.couchdb.org/en/1.6.1/api/server/common.html#replicate
type ReplicationResponse struct {
	History              []ReplicationHistory `json:"history"`
	Ok                   bool                 `json:"ok"`
	ReplicationIDVersion float64              `json:"replication_id_version"`
	SessionID            string               `json:"session_id"`
	SourceLastSeq        float64              `json:"source_last_seq"`
}

// RFC1123 is time format used by CouchDB for history fields.
// We have to define a custom type because Go uses RFC 3339 as default JSON time format.
//
// https://golang.org/pkg/time/#Time.MarshalJSON
// http://docs.couchdb.org/en/1.6.1/api/server/common.html#replicate
type RFC1123 time.Time

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// https://golang.org/pkg/encoding/json/#Unmarshaler
func (r *RFC1123) UnmarshalJSON(data []byte) error {
	var tmp string
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	t, err := time.Parse(time.RFC1123, tmp)
	if err != nil {
		return err
	}
	*r = RFC1123(t)
	return nil
}

// ReplicationHistory is part of the ReplicationResponse JSON object.
//
// http://docs.couchdb.org/en/1.6.1/api/server/common.html#replicate
type ReplicationHistory struct {
	DocWriteFailures float64 `json:"doc_write_failures"`
	DocsRead         float64 `json:"docs_read"`
	DocsWritten      float64 `json:"docs_written"`
	EndLastSeq       float64 `json:"end_last_seq"`
	EndTime          RFC1123 `json:"end_time"`
	MissingChecked   float64 `json:"missing_checked"`
	MissingFound     float64 `json:"missing_found"`
	RecordedSeq      float64 `json:"recorded_seq"`
	SessionID        string  `json:"session_id"`
	StartLastSeq     float64 `json:"start_last_seq"`
	StartTime        RFC1123 `json:"start_time"`
}

// Timestamp is time format used by CouchDB for the _replication_state_time field.
// It simply is a unix timestamp (number of seconds since 1 Jan 1970).
// We have to define our own custom type because Go uses RFC 3339 as default JSON time format.
//
// ttp://docs.couchdb.org/en/latest/replication/replicator.html#basics
type Timestamp time.Time

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// https://golang.org/pkg/encoding/json/#Unmarshaler
func (t *Timestamp) UnmarshalJSON(data []byte) error {
	var tmp float64
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	*t = Timestamp(time.Unix(int64(tmp), 0))
	return nil
}

// Replication is a document from the _replicator database.
// ReplicationState, ReplicationStateTime, ReplicationStateReason and ReplicationID are
// automatically updated by CouchDB.
//
// http://docs.couchdb.org/en/1.6.1/replication/replicator.html#basics
type Replication struct {
	ReplicationRequest
	ReplicationState       string    `json:"_replication_state"`
	ReplicationStateTime   Timestamp `json:"_replication_state_time"`
	ReplicationStateReason string    `json:"_replication_state_reason"`
	ReplicationID          string    `json:"_replication_id"`
}

// Replicate sends POST request to the _replicate URL.
//
// http://docs.couchdb.org/en/1.6.1/api/server/common.html#replicate
func (c *Client) Replicate(req ReplicationRequest) (*ReplicationResponse, error) {
	res, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(res)
	body, err := c.request(http.MethodPost, c.URL+"_replicate", data, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	r := &ReplicationResponse{}
	return r, json.NewDecoder(body).Decode(&r)
}

// internal helper function for http requests
func (c *Client) request(method, url string, data io.Reader, contentType string) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	// basic auth
	if c.Username != "" && c.Password != "" {
		req.SetBasicAuth(c.Username, c.Password)
	}
	// add cookies
	client := &http.Client{Jar: c.CookieJar}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// handle CouchDB http errors
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, newError(res)
	}
	// save new cookies
	c.CookieJar.SetCookies(req.URL, res.Cookies())
	return res.Body, nil
}
