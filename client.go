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

// Create a new client.
type Client struct {
	Url       string
	CookieJar *cookiejar.Jar
}

func NewClient(url string) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &Client{url, jar}, nil
}

// Get server information.
func (c *Client) Info() (*Server, error) {
	body, err := c.request("GET", c.Url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	server := &Server{}
	return server, json.Unmarshal(body, &server)
}

func (c *Client) Log() (string, error) {
	url := fmt.Sprintf("%s_log", c.Url)
	body, err := c.request("GET", url, nil, "")
	if err != nil {
		return "", err
	}
	return (string(body)), nil
}

// List of running tasks.
func (c *Client) ActiveTasks() ([]Task, error) {
	url := fmt.Sprintf("%s_active_tasks", c.Url)
	body, err := c.request("GET", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	tasks := []Task{}
	return tasks, json.Unmarshal(body, &tasks)
}

// Get all databases.
func (c *Client) All() ([]string, error) {
	url := fmt.Sprintf("%s_all_dbs", c.Url)
	body, err := c.request("GET", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	data := []string{}
	return data, json.Unmarshal(body, &data)
}

// Get database.
func (c *Client) Get(name string) (*DatabaseInfo, error) {
	url := fmt.Sprintf("%s%s", c.Url, name)
	body, err := c.request("GET", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	dbInfo := &DatabaseInfo{}
	return dbInfo, json.Unmarshal(body, &dbInfo)
}

// Create database.
func (c *Client) Create(name string) (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s%s", c.Url, name)
	body, err := c.request("PUT", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	return newDatabaseResponse(body)
}

// Delete database.
func (c *Client) Delete(name string) (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s%s", c.Url, name)
	body, err := c.request("DELETE", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	return newDatabaseResponse(body)
}

// Create user.
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
	return newDocumentResponse(body)
}

// Get user.
func (c *Client) GetUser(name string) (*User, error) {
	url := fmt.Sprintf("%s_users/org.couchdb.user:%s", c.Url, name)
	body, err := c.request("GET", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	user := &User{}
	return user, json.Unmarshal(body, &user)
}

// Delete user.
func (c *Client) DeleteUser(user User) (*DocumentResponse, error) {
	db := c.Use("_users")
	return db.Delete(user)
}

// Create session.
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
	sessionResponse := &PostSessionResponse{}
	return sessionResponse, json.Unmarshal(body, &sessionResponse)
}

// Get session.
func (c *Client) GetSession() (*GetSessionResponse, error) {
	url := fmt.Sprintf("%s_session", c.Url)
	body, err := c.request("GET", url, nil, "")
	if err != nil {
		return nil, err
	}
	sessionResponse := &GetSessionResponse{}
	return sessionResponse, json.Unmarshal(body, &sessionResponse)
}

// Delete session
func (c *Client) DeleteSession() (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s_session", c.Url)
	body, err := c.request("DELETE", url, nil, "")
	if err != nil {
		return nil, err
	}
	databaseResponse := &DatabaseResponse{}
	return databaseResponse, json.Unmarshal(body, &databaseResponse)
}

// Use database.
func (c *Client) Use(name string) Database {
	return Database{
		Url:    c.Url + name + "/",
		Client: c,
	}
}

// internal helper function for http requests
func (c *Client) request(method, url string, data io.Reader, contentType string) ([]byte, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	// add cookies
	client := &http.Client{Jar: c.CookieJar}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	// save new cookies
	c.CookieJar.SetCookies(req.URL, res.Cookies())
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
