package couchdb

import (
	"fmt"
	"testing"

	"github.com/segmentio/pointer"
)

type DataDocument struct {
	Document
	Type string `json:"type"`
	Foo  string `json:"foo"`
	Beep string `json:"beep"`
	Age  int    `json:"age"`
}

type Person struct {
	Document
	Type   string  `json:"type"`
	Name   string  `json:"name"`
	Age    float64 `json:"age"`
	Gender string  `json:"gender"`
}

var cView, _ = NewClient("http://127.0.0.1:5984/")
var dbView = cView.Use("gotest")

func TestViewBefore(t *testing.T) {
	// create database
	if _, err := cView.Create("gotest"); err != nil {
		t.Fatal(err)
	}
	design := &DesignDocument{
		Document: Document{
			ID: "_design/test",
		},
		Language: "javascript",
		Views: map[string]DesignDocumentView{
			"foo": DesignDocumentView{
				Map: `
					function(doc) {
						if (doc.type === 'data') {
							emit(doc.foo);
						}
					}
				`,
			},
			"int": DesignDocumentView{
				Map: `
					function(doc) {
						if (doc.type === 'data') {
							emit([doc.foo, doc.age]);
						}
					}
				`,
			},
			"complex": DesignDocumentView{
				Map: `
					function(doc) {
						if (doc.type === 'data') {
							emit([doc.foo, doc.beep]);
						}
					}
				`,
			},
		},
	}
	if _, err := dbView.Post(design); err != nil {
		t.Fatal(err)
	}
	// create design document for person
	designPerson := DesignDocument{
		Document: Document{
			ID: "_design/person",
		},
		Language: "javascript",
		Views: map[string]DesignDocumentView{
			"ageByGender": DesignDocumentView{
				Map: `
					function(doc) {
						if (doc.type === 'person') {
							emit(doc.gender, doc.age);
						}
					}
				`,
				Reduce: `
					function(keys, values, rereduce) {
						return sum(values);
					}
				`,
			},
		},
	}
	if _, err := dbView.Post(&designPerson); err != nil {
		t.Fatal(err)
	}
	// create dummy data
	doc1 := &DataDocument{
		Type: "data",
		Foo:  "foo1",
		Beep: "beep1",
		Age:  10,
	}
	if _, err := dbView.Post(doc1); err != nil {
		t.Fatal(err)
	}
	doc2 := &DataDocument{
		Type: "data",
		Foo:  "foo2",
		Beep: "beep2",
		Age:  20,
	}
	if _, err := dbView.Post(doc2); err != nil {
		t.Fatal(err)
	}
	// create multiple persons
	data := []struct {
		Name   string
		Age    float64
		Gender string
	}{
		{"John", 45, "male"},
		{"Frank", 40, "male"},
		{"Steve", 60, "male"},
		{"Max", 26, "male"},
		{"Marc", 36, "male"},
		{"Nick", 18, "male"},
		{"Jessica", 49, "female"},
		{"Lily", 20, "female"},
		{"Sophia", 66, "female"},
		{"Chloe", 12, "female"},
	}
	people := make([]Person, len(data))
	for index, d := range data {
		people[index] = Person{
			Type:   "person",
			Name:   d.Name,
			Age:    d.Age,
			Gender: d.Gender,
		}
	}
	// bulk save people to database
	if _, err := dbView.Bulk(people); err != nil {
		t.Fatal(err)
	}
}

func TestViewGet(t *testing.T) {
	view := dbView.View("test")
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
	err := dbView.Get(doc, "_design/test")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Name() != "test" {
		t.Error("design document Name() error")
	}
}

func TestDesignDocumentView(t *testing.T) {
	doc := new(DesignDocument)
	err := dbView.Get(doc, "_design/test")
	if err != nil {
		t.Fatal(err)
	}
	_, ok := doc.Views["foo"]
	if !ok {
		t.Error("design document view error")
	}
}

func TestViewGetWithQueryParameters(t *testing.T) {
	view := dbView.View("test")
	params := QueryParameters{
		Key: pointer.String(fmt.Sprintf("%q", "foo1")),
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
	view := dbView.View("test")
	params := QueryParameters{
		StartKey: pointer.String(fmt.Sprintf("[%q,%q]", "foo2", "beep2")),
		EndKey:   pointer.String(fmt.Sprintf("[%q,%q]", "foo2", "beep2")),
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
	view := dbView.View("test")
	params := QueryParameters{
		StartKey: pointer.String(fmt.Sprintf("[%q,%d]", "foo2", 20)),
		EndKey:   pointer.String(fmt.Sprintf("[%q,%d]", "foo2", 20)),
	}
	res, err := view.Get("int", params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 1 {
		t.Error("view get error")
	}
}

func TestViewGetWithReduce(t *testing.T) {
	view := dbView.View("person")
	params := QueryParameters{}
	res, err := view.Get("ageByGender", params)
	if err != nil {
		t.Fatal(err)
	}
	ageTotalSum := res.Rows[0].Value.(float64)
	if ageTotalSum != 372 {
		t.Fatalf("expected age 372 but got %v", ageTotalSum)
	}
}

func TestViewGetWithReduceAndGroup(t *testing.T) {
	view := dbView.View("person")
	params := QueryParameters{
		Key:        pointer.String(fmt.Sprintf("%q", "female")),
		GroupLevel: pointer.Int(1),
	}
	res, err := view.Get("ageByGender", params)
	if err != nil {
		t.Fatal(err)
	}
	ageTotalFemale := res.Rows[0].Value.(float64)
	if ageTotalFemale != 147 {
		t.Fatalf("expected age 147 but got %v", ageTotalFemale)
	}
}

func TestViewGetWithoutReduce(t *testing.T) {
	view := dbView.View("person")
	params := QueryParameters{
		Key:    pointer.String(fmt.Sprintf("%q", "male")),
		Reduce: pointer.Bool(false),
	}
	res, err := view.Get("ageByGender", params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 6 {
		t.Fatalf("expected 6 rows but got %d instead", len(res.Rows))
	}
}

func TestViewPost(t *testing.T) {
	view := dbView.View("person")
	params := QueryParameters{
		Reduce: pointer.Bool(false),
	}
	res, err := view.Post("ageByGender", []string{"male"}, params)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rows) != 6 {
		t.Fatalf("expected 6 rows but got %d instead", len(res.Rows))
	}
}

func TestViewAfter(t *testing.T) {
	if _, err := cView.Delete("gotest"); err != nil {
		t.Fatal(err)
	}
}
