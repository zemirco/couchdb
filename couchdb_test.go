package couchdb

import (
  "testing"
  "reflect"
)

var client = Client{"http://127.0.0.1:5984/"}

func TestInfo(t *testing.T) {
  info, err := client.Info()
  if err != nil {
    t.Fatal(err)
  }
  if info.Couchdb != "Welcome" {
    t.Error("Couchdb error")
  }
}

func TestActiveTasks(t *testing.T) {
  res, err := client.ActiveTasks()
  if err != nil {
    t.Fatal(err)
  }
  out := make([]Task, 0)
  if reflect.DeepEqual(out, res) == false {
    t.Error("active tasks should be an empty array")
  }
}

func TestAll(t *testing.T) {
  res, err := client.All()
  if err != nil {
    t.Fatal(err)
  }
  if res[0] != "_replicator" || res[1] != "_users" {
    t.Error("slice error")
  }
}

func TestGet(t *testing.T) {
  info, err := client.Get("_users")
  if err != nil {
    t.Fatal(err)
  }
  if info.DbName != "_users" {
    t.Error("DbName error")
  }
  if info.CompactRunning != false {
    t.Error("CompactRunning error")
  }
}

func TestCreate(t *testing.T) {
  status, err := client.Create("dummy")
  if err != nil {
    t.Fatal(err)
  }
  if status.Ok != true {
    t.Error("status error")
  }
  if status.Error != "" || status.Reason != "" {
    t.Error("error and reason error")
  }
}

func TestCreateFail(t *testing.T) {
  _, err := client.Create("dummy")
  if err == nil {
    t.Fatal("should not create duplicate database")
  }
}

func TestDelete(t *testing.T) {
  status, err := client.Delete("dummy")
  if err != nil {
    t.Fatal(err)
  }
  if status.Ok != true {
    t.Error("status error")
  }
  if status.Error != "" || status.Reason != "" {
    t.Error("error and reason error")
  }
}

func TestDeleteFail(t *testing.T) {
  _, err := client.Delete("dummy")
  if err == nil {
    t.Fatal("should not delete non existing database")
  }
}

func TestUse(t *testing.T) {
  out := Database{
    Url: "http://127.0.0.1:5984/_users/",
  }
  db := client.Use("_users")
  if reflect.DeepEqual(out, db) == false {
    t.Error("use error")
  }
}
