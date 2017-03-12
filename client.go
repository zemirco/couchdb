package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// Client holds all info for database client
type Client struct {
	Username  string
	Password  string
	BaseURL   *url.URL
	CookieJar *cookiejar.Jar
}

// NewClient returns new couchdb client for given url
func NewClient(u *url.URL) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	c := &Client{
		Username:  "",
		Password:  "",
		BaseURL:   u,
		CookieJar: jar,
	}
	return c, nil
}

// NewAuthClient returns new couchdb client with basic authentication
func NewAuthClient(username, password string, u *url.URL) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Client{
		Username:  username,
		Password:  password,
		BaseURL:   u,
		CookieJar: jar,
	}, nil
}

// Info returns some information about the server
func (c *Client) Info() (*Server, error) {
	u := ""
	res, err := c.Request(http.MethodGet, u, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	server := &Server{}
	return server, json.NewDecoder(res.Body).Decode(&server)
}

// ActiveTasks returns list of currently running tasks
func (c *Client) ActiveTasks() ([]Task, error) {
	u := "_active_tasks"
	res, err := c.Request(http.MethodGet, u, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	tasks := []Task{}
	return tasks, json.NewDecoder(res.Body).Decode(&tasks)
}

// All returns list of all databases on server
func (c *Client) All() ([]string, error) {
	u := "_all_dbs"
	res, err := c.Request(http.MethodGet, u, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data := []string{}
	return data, json.NewDecoder(res.Body).Decode(&data)
}

// Get database.
func (c *Client) Get(name string) (*DatabaseInfo, error) {
	u := url.PathEscape(name)
	res, err := c.Request(http.MethodGet, u, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	dbInfo := &DatabaseInfo{}
	return dbInfo, json.NewDecoder(res.Body).Decode(&dbInfo)
}

// Create database.
func (c *Client) Create(name string) (*DatabaseResponse, error) {
	u := url.PathEscape(name)
	res, err := c.Request(http.MethodPut, u, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response DatabaseResponse
	return &response, json.NewDecoder(res.Body).Decode(&response)
}

// Delete database.
func (c *Client) Delete(name string) (*DatabaseResponse, error) {
	u := url.PathEscape(name)
	res, err := c.Request(http.MethodDelete, u, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response DatabaseResponse
	return &response, json.NewDecoder(res.Body).Decode(&response)
}

// CreateUser creates a new user in _users database
func (c *Client) CreateUser(user User) (*DocumentResponse, error) {
	u := fmt.Sprintf("_users/%s", user.ID)
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(user); err != nil {
		return nil, err
	}
	res, err := c.Request(http.MethodPut, u, &b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response DocumentResponse
	return &response, json.NewDecoder(res.Body).Decode(&response)
}

// GetUser returns user by given name
func (c *Client) GetUser(name string) (*User, error) {
	u := fmt.Sprintf("_users/org.couchdb.user:%s", name)
	res, err := c.Request(http.MethodGet, u, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	user := &User{}
	return user, json.NewDecoder(res.Body).Decode(&user)
}

// DeleteUser removes user from database
func (c *Client) DeleteUser(user *User) (*DocumentResponse, error) {
	db := c.Use("_users")
	return db.Delete(user)
}

// CreateSession creates a new session and logs in user
func (c *Client) CreateSession(name, password string) (*PostSessionResponse, error) {
	u := "_session"
	creds := Credentials{name, password}
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(creds); err != nil {
		return nil, err
	}
	res, err := c.Request(http.MethodPost, u, &b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	sessionResponse := &PostSessionResponse{}
	return sessionResponse, json.NewDecoder(res.Body).Decode(&sessionResponse)
}

// GetSession returns session for currently logged in user
func (c *Client) GetSession() (*GetSessionResponse, error) {
	u := "_session"
	res, err := c.Request(http.MethodGet, u, nil, "")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	sessionResponse := &GetSessionResponse{}
	return sessionResponse, json.NewDecoder(res.Body).Decode(&sessionResponse)
}

// DeleteSession removes current session and logs out user
func (c *Client) DeleteSession() (*DatabaseResponse, error) {
	u := "_session"
	res, err := c.Request(http.MethodDelete, u, nil, "")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	databaseResponse := &DatabaseResponse{}
	return databaseResponse, json.NewDecoder(res.Body).Decode(&databaseResponse)
}

// Use database.
func (c *Client) Use(name string) DatabaseService {
	return &Database{
		Name:   name,
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
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(req); err != nil {
		return nil, err
	}
	res, err := c.Request(http.MethodPost, "_replicate", &b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	r := &ReplicationResponse{}
	return r, json.NewDecoder(res.Body).Decode(&r)
}

// Request creates new http request and does it.
func (c *Client) Request(method, uri string, data io.Reader, contentType string) (*http.Response, error) {
	rel, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	u := c.BaseURL.ResolveReference(rel)
	req, err := http.NewRequest(method, u.String(), data)
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
	return res, nil
}

const (
	fileNameMap    = "map.js"
	fileNameReduce = "reduce.js"
)

// Parse takes a location and parses all design documents with corresponding views.
// The folder structure must look like this.
//
//   design
//   |-- player
//   |   |-- byAge
//   |   |   |-- map.js
//   |   |   `-- reduce.js
//   |   `-- byName
//   |       `-- map.js
//   `-- user
//       |-- byEmail
//       |   |-- map.js
//       |   `-- reduce.js
//       `-- byUsername
//           `-- map.js
func (c *Client) Parse(dirname string) ([]DesignDocument, error) {
	docs := []DesignDocument{}
	// get all directories inside location which will become separate design documents
	dirs, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, err
	}
	for _, dir := range dirs {
		designDocumentName := dir.Name()
		ff, err := ioutil.ReadDir(filepath.Join(dirname, designDocumentName))
		if err != nil {
			return nil, err
		}
		d := DesignDocument{
			Document: Document{
				ID: fmt.Sprintf("_design/%s", designDocumentName),
			},
			Language: langJavaScript,
			Views:    map[string]DesignDocumentView{},
		}
		for _, j := range ff {
			viewName := j.Name()
			// create new view inside design document
			view := DesignDocumentView{}
			// get map function
			pathMap := filepath.Join(dirname, designDocumentName, viewName, fileNameMap)
			bMap, err := ioutil.ReadFile(pathMap)
			if err != nil {
				return nil, err
			}
			view.Map = string(bMap)
			// get reduce function only if it exists
			pathReduce := filepath.Join(dirname, designDocumentName, viewName, fileNameReduce)
			if _, err := os.Stat(pathReduce); err != nil {
				// ignore error that file does not exist but return other errors
				if !os.IsNotExist(err) {
					return nil, err
				}
			} else {
				bReduce, err := ioutil.ReadFile(pathReduce)
				if err != nil {
					return nil, err
				}
				view.Reduce = string(bReduce)
			}
			d.Views[viewName] = view
		}
		docs = append(docs, d)
	}
	return docs, nil
}
