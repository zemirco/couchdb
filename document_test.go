package couchdb

import (
  "testing"
  "log"
)

type DummyDocument struct {
  Document
  Foo string `json:"foo"`
  Beep string `json:"beep"`
}

func (doc *DummyDocument) GetId() string {
  return doc.Document.Id
}

func (doc *DummyDocument) GetRev() string {
  return doc.Document.Rev
}

// init client and test database
var c = Client{"http://127.0.0.1:5984/"}
var db = c.use("dummy")

func TestBefore(t *testing.T) {
  log.Print("creating dummy database")
  _, err := client.create("dummy")
  if err != nil {
    log.Fatal(err)
  }
}

func TestDocumentPost(t *testing.T) {
  doc := &DummyDocument{
    Document: Document{
      Id: "testid",
    },
    Foo: "bar",
    Beep: "bopp",
  }
  if doc.Rev != "" {
    t.Error("new document should not have a revision")
  }
  res, err := db.Post(doc)
  if err != nil {
    t.Fatal(err)
  }
  if res.Ok == false {
    t.Error("post document error")
  }
}

func TestDocumentHead(t *testing.T) {
  head, err := db.Head("testid")
  if err != nil {
    t.Fatal(err)
  }
  if head.StatusCode != 200 {
    t.Error("document head error")
  }
}

func TestDocumentGet(t *testing.T) {
  doc := new(DummyDocument)
  err := db.Get(doc, "testid")
  if err != nil {
    t.Fatal(err)
  }
  if doc.Foo != "bar" || doc.Beep != "bopp" {
    t.Error("document fields error")
  }
}

func TestDocumentPut(t *testing.T) {
  // get document
  doc := new(DummyDocument)
  err := db.Get(doc, "testid")
  if err != nil {
    t.Fatal(err)
  }
  // change document
  doc.Foo = "baz"
  res, err := db.Put(doc)
  if err != nil {
    t.Fatal(err)
  }
  if res.Id != "testid" || res.Ok == false {
    t.Error("put document response error")
  }
}

func TestDocumentDelete(t *testing.T) {
  // get document
  doc := new(DummyDocument)
  err := db.Get(doc, "testid")
  if err != nil {
    t.Fatal(err)
  }
  // delete document
  res, err := db.Delete(doc)
  if err != nil {
    t.Fatal(err)
  }
  if res.Id != "testid" || res.Ok == false {
    t.Error("delete document response error")
  }
}

func TestAfter(t *testing.T) {
  log.Print("deleting dummy database")
  _, err := client.delete("dummy")
  if err != nil {
    log.Fatal(err)
  }
}
