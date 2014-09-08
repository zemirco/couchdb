package couchdb

import (
  "fmt"
)

type Client struct {
  Url string
}

// http://docs.couchdb.org/en/latest/intro/api.html#server
type Server struct {
  Couchdb string
  Uuid string
  Vendor struct {
    Version string
    Name string
  }
  Version string
}

// http://docs.couchdb.org/en/latest/api/database/common.html#get--db
type DatabaseInfo struct {
  DbName string `json:"db_name"`
  DocCount int `json:"doc_count"`
  DocDelCount int `json:"doc_del_count"`
  UpdateSeq int `json:"update_seq"`
  PurgeSeq int `json:"purge_seq"`
  CompactRunning bool `json:"compact_running"`
  DiskSize int `json:"disk_size"`
  DataSize int `json:"data_size"`
  InstanceStartTime string `json:"instance_start_time"`
  DiskFormatVersion int `json:"disk_format_version"`
  CommittedUpdateSeq int `json:"committed_update_seq"`
}

type DbResponse struct {
  Ok bool
  Error string
  Reason string
}

type Error struct {
  Method string
  Url string
  StatusCode int
  Type string `json:"error"`
  Reason string
}

func (e *Error) Error() string {
  return fmt.Sprintf("CouchDB - %s %s, Status Code: %d, Error: %s, Reason: %s", e.Method, e.Url, e.StatusCode, e.Type, e.Reason)
}

type Document struct {
  Id string `json:"_id,omitempty"`
  Rev string `json:"_rev,omitempty"`
}

type CouchDoc interface {
  GetId() string
  GetRev() string
}

type DocumentResponse struct {
  Ok bool
  Id string
  Rev string
}
