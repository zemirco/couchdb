package couchdb

import (
	"encoding/json"
	"fmt"
)

// Get server information.
func (c *Client) Info() (*Server, error) {
	body, err := request("GET", c.Url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	server := &Server{}
	return server, json.Unmarshal(body, &server)
}

func (c *Client) Log() (string, error) {
	url := fmt.Sprintf("%s_log", c.Url)
	body, err := request("GET", url, nil, "")
	if err != nil {
		return "", err
	}
	return (string(body)), nil
}

// List of running tasks.
func (c *Client) ActiveTasks() ([]Task, error) {
	url := fmt.Sprintf("%s_active_tasks", c.Url)
	body, err := request("GET", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	tasks := []Task{}
	return tasks, json.Unmarshal(body, &tasks)
}

// Get all databases.
func (c *Client) All() ([]string, error) {
	url := fmt.Sprintf("%s_all_dbs", c.Url)
	body, err := request("GET", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	data := []string{}
	return data, json.Unmarshal(body, &data)
}

// Get database.
func (c *Client) Get(name string) (*DatabaseInfo, error) {
	url := fmt.Sprintf("%s%s", c.Url, name)
	body, err := request("GET", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	dbInfo := &DatabaseInfo{}
	return dbInfo, json.Unmarshal(body, &dbInfo)
}

// Create database.
func (c *Client) Create(name string) (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s%s", c.Url, name)
	body, err := request("PUT", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	return newDatabaseResponse(body)
}

// Delete database.
func (c *Client) Delete(name string) (*DatabaseResponse, error) {
	url := fmt.Sprintf("%s%s", c.Url, name)
	body, err := request("DELETE", url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	return newDatabaseResponse(body)
}

// Use database.
func (c *Client) Use(name string) Database {
	return Database{c.Url + name + "/"}
}