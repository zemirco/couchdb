package couchdb

import (
  "encoding/json"
)

// Get server information.
func (c *Client) Info() (*Server, error) {
  body, err := request("GET", c.Url, nil)
  if err != nil {
    return nil, err
  }
  var server *Server
  err = json.Unmarshal(body, &server)
  if err != nil {
    return nil, err
  }
  return server, nil
}

// List of running tasks.
func (c *Client) ActiveTasks() ([]Task, error) {
  body, err := request("GET", c.Url + "_active_tasks", nil)
  if err != nil {
    return nil, err
  }
  var tasks []Task
  return tasks, json.Unmarshal(body, &tasks)
}

// Get all databases.
func (c *Client) All() ([]string, error) {
  body, err := request("GET", c.Url + "_all_dbs", nil)
  if err != nil {
    return nil, err
  }
  var data []string
  return data, json.Unmarshal(body, &data)
}

// Get database.
func (c *Client) Get(name string) (*DatabaseInfo, error) {
  body, err := request("GET", c.Url + name, nil)
  if err != nil {
    return nil, err
  }
  var dbInfo *DatabaseInfo
  return dbInfo, json.Unmarshal(body, &dbInfo)
}

// Create database.
func (c *Client) Create(name string) (*DatabaseResponse, error) {
  body, err := request("PUT", c.Url + name, nil)
  if err != nil {
    return nil, err
  }
  return newDatabaseResponse(body)
}

// Delete database.
func (c *Client) Delete(name string) (*DatabaseResponse, error) {
  body, err := request("DELETE", c.Url + name, nil)
  if err != nil {
    return nil, err
  }
  return newDatabaseResponse(body)
}

// Use database.
func (c *Client) Use(name string) (Database) {
  return Database{c.Url + name + "/"}
}
