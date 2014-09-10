package couchdb

import (
	"testing"
)

type ViewDocument struct {
	Document
	Views map[string]interface{} `json:"views"`
}

func (doc *ViewDocument) GetDocument() *Document {
	return &doc.Document
}

type DataDocument struct {
	Document
	Type string `json:"type"`
	Foo  string `json:"foo"`
	Beep string `json:"beep"`
}

func (doc *DataDocument) GetDocument() *Document {
	return &doc.Document
}

var c_view = Client{"http://127.0.0.1:5984/"}
var db_view = c_view.Use("gotest")

func TestViewBefore(t *testing.T) {

	// create database
	t.Log("creating database...")
	_, err := c_view.Create("gotest")
	if err != nil {
		t.Fatal(err)
	}

	// create design document
	t.Log("creating design document...")
	view := make(map[string]string)
	view["map"] =
		`
    function(doc){
      if (doc.type == 'data') emit(doc.foo)
    }
  `
	views := make(map[string]interface{})
	views["foo"] = view

	design := &ViewDocument{
		Document: Document{
			Id: "_design/test",
		},
		Views: views,
	}

	_, err = db_view.Post(design)
	if err != nil {
		t.Fatal(err)
	}

	// create dummy data
	t.Log("creating dummy data...")
	doc1 := &DataDocument{
		Type: "data",
		Foo:  "foo1",
		Beep: "beep1",
	}

	_, err = db_view.Post(doc1)
	if err != nil {
		t.Fatal(err)
	}

	doc2 := &DataDocument{
		Type: "data",
		Foo:  "foo2",
		Beep: "beep2",
	}

	_, err = db_view.Post(doc2)
	if err != nil {
		t.Fatal(err)
	}
}

func TestViewGet(t *testing.T) {
	view := db_view.View("test")
	params := QueryParameters{}
	res, err := view.Get("foo", params)
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalRows != 2 || res.Offset != 0 {
		t.Error("view get error")
	}
}

func TestViewGetWithQueryParameters(t *testing.T) {
  view := db_view.View("test")
  params := QueryParameters{
    Key: "\"foo1\"",
  }
  res, err := view.Get("foo", params)
  if err != nil {
    t.Fatal(err)
  }
	if len(res.Rows) != 1 {
		t.Error("view get error")
	}
}

func TestViewAfter(t *testing.T) {
	t.Log("deleting test data for view tests...")
	_, err := c_view.Delete("gotest")
	if err != nil {
		t.Fatal(err)
	}
}
