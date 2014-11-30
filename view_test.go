package couchdb

import (
	"testing"
)

type DataDocument struct {
	Document
	Type string `json:"type"`
	Foo  string `json:"foo"`
	Beep string `json:"beep"`
	Age  int    `json:"age"`
}

var c_view, _ = NewClient("http://127.0.0.1:5984/")
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
	view := DesignDocumentView{}
	view.Map =
		`
		function(doc){
			if (doc.type == 'data') emit(doc.foo)
		}
	`
	// create a bit more comple design document
	t.Log("creating complex design document...")
	complexView := DesignDocumentView{}
	complexView.Map =
		`
		function(doc) {
			if (doc.type === 'data') emit([doc.foo, doc.beep])
		}
		`

	// create design document with int key
	t.Log("creating int design document...")
	intView := DesignDocumentView{}
	intView.Map =
		`
		function(doc) {
			if (doc.type === 'data') emit([doc.foo, doc.age])
		}
		`

	views := make(map[string]DesignDocumentView)
	views["foo"] = view
	views["complex"] = complexView
	views["int"] = intView

	// views := make(map[string]interface{})
	design := &DesignDocument{
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
		Age:  10,
	}

	_, err = db_view.Post(doc1)
	if err != nil {
		t.Fatal(err)
	}

	doc2 := &DataDocument{
		Type: "data",
		Foo:  "foo2",
		Beep: "beep2",
		Age:  20,
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

func TestDesignDocumentName(t *testing.T) {
	doc := new(DesignDocument)
	err := db_view.Get(doc, "_design/test")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Name() != "test" {
		t.Error("design document Name() error")
	}
}

func TestDesignDocumentView(t *testing.T) {
	doc := new(DesignDocument)
	err := db_view.Get(doc, "_design/test")
	if err != nil {
		t.Fatal(err)
	}
	_, ok := doc.Views["foo"]
	if ok == false {
		t.Error("design document view error")
	}
}

func TestViewGetWithQueryParameters(t *testing.T) {
	view := db_view.View("test")
	params := QueryParameters{
		Key: `"foo1"`,
	}
	res, err := view.Get("foo", params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 1 {
		t.Error("view get error")
	}
}

func TestViewGetWithStartKeyEndKey(t *testing.T) {
	view := db_view.View("test")

	params := QueryParameters{
		StartKey: `["foo2","beep2"]`,
		EndKey:   `["foo2","beep2"]`,
	}
	res, err := view.Get("complex", params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 1 {
		t.Error("view get error")
	}
}

func TestViewGetWithInteger(t *testing.T) {
	view := db_view.View("test")

	params := QueryParameters{
		StartKey: `["foo2",20]`,
		EndKey:   `["foo2",20]`,
	}
	res, err := view.Get("int", params)
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
