package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
)

// Client holds all info for database client
type Client struct {
	Username  string
	Password  string
	Url       string
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
	body, err := c.request("GET", c.Url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	server := &Server{}
	return server, json.NewDecoder(body).Decode(&server)
}

// Log returns logs from database
func (c *Client) Log() (string, error) {
	url := fmt.Sprintf("%s_log", c.Url)
	body, err := c.request("GET", url, nil, "")
	if err != nil {
		return "", err
	}
	defer body.Close()
	log, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}
	return (string(log)), nil
}

// ActiveTasks returns list of currently running tasks
func (c *Client) ActiveTasks() ([]Task, error) {
	url := fmt.Sprintf("%s_active_tasks", c.Url)
	body, err := c.request("GET", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	tasks := []Task{}
	return tasks, json.NewDecoder(body).Decode(&tasks)
}

// All returns list of all databases on server
func (c *Client) All() ([]string, error) {
	url := fmt.Sprintf("%s_all_dbs", c.Url)
	body, err := c.request("GET", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	data := []string{}
	return data, json.NewDecoder(body).Decode(&data)
}

// Get database.
func (c *Client) Get(name string) (*DatabaseInfo, error) {
	url := fmt.Sprintf("%s%s", c.Url, name)
	body, err := c.request("GET", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	dbInfo := &DatabaseInfo{}
	return dbInfo, json.NewDecoder(body).Decode(&dbInfo)
}

// Create database.
func (c *Client) Create(name string) (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s%s", c.Url, name)
	body, err := c.request("PUT", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDatabaseResponse(body)
}

// Delete database.
func (c *Client) Delete(name string) (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s%s", c.Url, name)
	body, err := c.request("DELETE", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDatabaseResponse(body)
}

// CreateUser creates a new user in _users database
func (c *Client) CreateUser(user User) (*DocumentResponse, error) {
	url := fmt.Sprintf("%s_users/%s", c.Url, user.Id)
	res, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(res)
	body, err := c.request("PUT", url, data, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDocumentResponse(body)
}

// GetUser returns user by given name
func (c *Client) GetUser(name string) (*User, error) {
	url := fmt.Sprintf("%s_users/org.couchdb.user:%s", c.Url, name)
	body, err := c.request("GET", url, nil, "application/json")
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
	url := fmt.Sprintf("%s_session", c.Url)
	creds := Credentials{name, password}
	res, err := json.Marshal(creds)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(res)
	body, err := c.request("POST", url, data, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	sessionResponse := &PostSessionResponse{}
	return sessionResponse, json.NewDecoder(body).Decode(&sessionResponse)
}

// GetSession returns session for currently logged in user
func (c *Client) GetSession() (*GetSessionResponse, error) {
	url := fmt.Sprintf("%s_session", c.Url)
	body, err := c.request("GET", url, nil, "")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	sessionResponse := &GetSessionResponse{}
	return sessionResponse, json.NewDecoder(body).Decode(&sessionResponse)
}

// DeleteSession removes current session and logs out user
func (c *Client) DeleteSession() (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s_session", c.Url)
	body, err := c.request("DELETE", url, nil, "")
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
		Url:    c.Url + name + "/",
		Client: c,
	}
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
