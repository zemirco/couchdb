package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"

	"github.com/segmentio/pointer"
)

var (
	client *Client
	c      *Client
	cView  *Client
	db     Database
	dbView Database
)

func TestMain(m *testing.M) {
	u, err := url.Parse("http://127.0.0.1:5984/")
	if err != nil {
		panic(err)
	}
	client, err = NewClient(u)
	if err != nil {
		panic(err)
	}
	c, err = NewClient(u)
	if err != nil {
		panic(err)
	}
	db = c.Use("dummy")
	cView, err = NewClient(u)
	if err != nil {
		panic(err)
	}
	dbView = cView.Use("gotest")
	code := m.Run()
	// clean up
	os.Exit(code)
}

func TestInfo(t *testing.T) {
	info, err := client.Info()
	if err != nil {
		t.Fatal(err)
	}
	if info.Couchdb != "Welcome" {
		t.Errorf("expected Welcome got %s", info.Couchdb)
	}
}

func TestActiveTasks(t *testing.T) {
	res, err := client.ActiveTasks()
	if err != nil {
		t.Fatal(err)
	}
	out := make([]Task, 0)
	if !reflect.DeepEqual(out, res) {
		t.Errorf("expected %v got %v", out, res)
	}
}

func TestAll(t *testing.T) {
	res, err := client.All()
	if err != nil {
		t.Fatal(err)
	}
	if res[0] != "_replicator" {
		t.Errorf("expected 1st db to be _replicator but got %s", res[0])
	}
	if res[1] != "_users" {
		t.Errorf("expected 2nd db to be _users but got %s", res[1])
	}
}

func TestGet(t *testing.T) {
	info, err := client.Get("_users")
	if err != nil {
		t.Fatal(err)
	}
	if info.DbName != "_users" {
		t.Errorf("expected name _users got %s", info.DbName)
	}
	if info.CompactRunning {
		t.Errorf("expected compact running to be false got true")
	}
}

func TestCreate(t *testing.T) {
	status, err := client.Create("dummy")
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ok {
		t.Errorf("expected ok to be true got false")
	}
}

func TestCreateFail(t *testing.T) {
	_, err := client.Create("dummy")
	if err == nil {
		t.Fatal("creating duplicate database should return an error")
	}
	if couchdbError, ok := err.(*Error); ok {
		if couchdbError.StatusCode != http.StatusPreconditionFailed {
			t.Fatal("creating duplicate database should return an error")
		}
	}
}

func TestCreateUser(t *testing.T) {
	user := NewUser("john", "password", []string{})
	res, err := client.CreateUser(user)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Errorf("expected ok to be true got false")
	}
	if res.ID != "org.couchdb.user:john" {
		t.Errorf("expected res id org.couchdb.user:john but got %s", res.ID)
	}
}

func TestCreateSession(t *testing.T) {
	res, err := client.CreateSession("john", "password")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Errorf("expected ok to be true got false")
	}
	if res.Name != "john" {
		t.Errorf("expected res name john but got %s", res.Name)
	}
}

func TestGetSession(t *testing.T) {
	session, err := client.GetSession()
	if err != nil {
		t.Fatal(err)
	}
	if !session.Ok {
		t.Errorf("expected ok to be true got false")
	}
	if session.UserContext.Name != "john" {
		t.Errorf("expected user context name john but got %s", session.UserContext.Name)
	}
}

func TestDeleteSession(t *testing.T) {
	res, err := client.DeleteSession()
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Errorf("expected ok to be true got false")
	}
}

func TestGetUser(t *testing.T) {
	user, err := client.GetUser("john")
	if err != nil {
		t.Fatal(err)
	}
	if user.Name != "john" {
		t.Errorf("expected name john but got %s", user.Name)
	}
	if user.Type != "user" {
		t.Errorf("expected type user but got %s", user.Type)
	}
	if user.Iterations != 10 {
		t.Errorf("expected 10 iterations but got %d", user.Iterations)
	}
}

func TestDeleteUser(t *testing.T) {
	user, err := client.GetUser("john")
	if err != nil {
		t.Fatal(err)
	}
	res, err := client.DeleteUser(user)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Errorf("expected ok to be true got false")
	}
	if res.ID != "org.couchdb.user:john" {
		t.Errorf("expected res id to be org.couchdb.user:john but got %s", res.ID)
	}
}

func TestGetSessionAdmin(t *testing.T) {
	session, err := client.GetSession()
	if err != nil {
		t.Fatal(err)
	}
	if !session.Ok {
		t.Error("expected ok to be true but got false")
	}
	roles := []string{"_admin"}
	if !reflect.DeepEqual(roles, session.UserContext.Roles) {
		t.Errorf("expected roles %v but got %v", roles, session.UserContext.Roles)
	}
}

func TestDelete(t *testing.T) {
	status, err := client.Delete("dummy")
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ok {
		t.Error("expected ok to be true but got false")
	}
}

func TestDeleteFail(t *testing.T) {
	_, err := client.Delete("dummy")
	if err == nil {
		t.Fatal("should not delete non existing database")
	}
	if couchdbError, ok := err.(*Error); ok {
		if couchdbError.StatusCode != http.StatusNotFound {
			t.Fatal("should not delete non existing database")
		}
	}
}

func TestUse(t *testing.T) {
	db := client.Use("_users")
	if db.Name != "_users/" {
		t.Errorf("expected _users/ got %s", db.Name)
	}
}

type animal struct {
	Document
	Type   string `json:"type"`
	Animal string `json:"animal"`
	Owner  string `json:"owner"`
}

func TestReplication(t *testing.T) {
	name := "replication"
	name2 := "replication2"
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	// add some documents to database
	db := client.Use(name)
	for _, a := range []string{"dog", "mouse", "cat"} {
		doc := &animal{
			Type:   "animal",
			Animal: a,
		}
		if _, err := db.Post(doc); err != nil {
			t.Error(err)
		}
	}
	// replicate
	req := ReplicationRequest{
		CreateTarget: true,
		Source:       "http://localhost:5984/" + name,
		Target:       "http://localhost:5984/" + name2,
	}
	r, err := c.Replicate(req)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Ok {
		t.Error("expected ok to be true but got false instead")
	}
	// remove both databases
	for _, d := range []string{name, name2} {
		if _, err := client.Delete(d); err != nil {
			t.Fatal(err)
		}
	}
}

func TestReplicationFilter(t *testing.T) {
	dbName := "replication_filter"
	dbName2 := "replication_filter2"
	// create database
	if _, err := client.Create(dbName); err != nil {
		t.Error(err)
	}
	// add some documents to database
	db := client.Use(dbName)
	docs := []animal{
		{
			Type:   "animal",
			Animal: "dog",
			Owner:  "john",
		},
		{
			Type:   "animal",
			Animal: "cat",
			Owner:  "john",
		},
		{
			Type:   "animal",
			Animal: "horse",
			Owner:  "steve",
		},
	}
	for _, doc := range docs {
		if _, err := db.Post(&doc); err != nil {
			t.Error(err)
		}
	}
	// create view document with filter function in first database
	designDocument := &DesignDocument{
		Document: Document{
			ID: "_design/animals",
		},
		Language: "javascript",
		Filters: map[string]string{
			"byOwner": `
				function(doc, req) {
					if (doc.owner === req.query.owner) {
						return true
					}
					return false
				}
			`,
		},
	}
	if _, err := db.Post(designDocument); err != nil {
		t.Error(err)
	}
	// create replication with filter function
	req := ReplicationRequest{
		CreateTarget: true,
		Source:       "http://localhost:5984/" + dbName,
		Target:       "http://localhost:5984/" + dbName2,
		Filter:       "animals/byOwner",
		QueryParams: map[string]string{
			"owner": "john",
		},
	}
	if _, err := c.Replicate(req); err != nil {
		t.Error(err)
	}
	// check replicated database
	db = client.Use(dbName2)
	allDocs, err := db.AllDocs(nil)
	if err != nil {
		t.Error(err)
	}
	if len(allDocs.Rows) != 2 {
		t.Errorf("expected exactly two documents but got %d instead", len(allDocs.Rows))
	}
	// remove both databases
	for _, d := range []string{dbName, dbName2} {
		if _, err := client.Delete(d); err != nil {
			t.Fatal(err)
		}
	}
}

// test continuous replication to test getting replication document
// with custom time format.
func TestReplicationContinuous(t *testing.T) {
	dbName := "continuous"
	dbName2 := "continuous2"
	// create database
	if _, err := client.Create(dbName); err != nil {
		t.Error(err)
	}
	// create replication document inside _replicate database
	req := ReplicationRequest{
		Document: Document{
			ID: "awesome",
		},
		Continuous:   true,
		CreateTarget: true,
		Source:       "http://localhost:5984/" + dbName,
		Target:       "http://localhost:5984/" + dbName2,
	}
	if _, err := c.Replicate(req); err != nil {
		t.Error(err)
	}
	tasks, err := c.ActiveTasks()
	if err != nil {
		t.Error(err)
	}
	if tasks[0].Type != "replication" {
		t.Errorf("expected type replication but got %s instead", tasks[0].Type)
	}
	// remove both databases
	for _, d := range []string{dbName, dbName2} {
		if _, err := client.Delete(d); err != nil {
			t.Fatal(err)
		}
	}
}

func TestRequest(t *testing.T) {
	name := "test_request"
	// create database
	if _, err := client.Create(name); err != nil {
		t.Fatal(err)
	}
	// add some documents to database
	db := client.Use(name)
	animals := []string{"dog", "mouse", "cat"}
	docs := make([]CouchDoc, len(animals))
	for i, a := range animals {
		doc := &animal{
			Type:   "animal",
			Animal: a,
		}
		docs[i] = doc
	}
	if _, err := db.Bulk(docs); err != nil {
		t.Fatal(err)
	}
	// get all documents
	includeDocs := true
	q := QueryParameters{
		IncludeDocs: &includeDocs,
	}
	data, err := db.AllDocs(&q)
	if err != nil {
		t.Fatal(err)
	}
	// change single document
	doc := data.Rows[0].Doc
	// make post request to database
	doc["owner"] = "zemirco"
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(doc); err != nil {
		t.Fatal(err)
	}
	u := fmt.Sprintf("%s/%s", name, doc["_id"])
	if _, err := client.Request(http.MethodPut, u, &b, "application/json"); err != nil {
		t.Fatal(err)
	}
	// remove database
	if _, err := client.Delete(name); err != nil {
		t.Fatal(err)
	}
}

// database tests
type DummyDocument struct {
	Document
	Foo  string `json:"foo"`
	Beep string `json:"beep"`
}

func TestBefore(t *testing.T) {
	if _, err := client.Create("dummy"); err != nil {
		panic(err)
	}
}

func TestDocumentPost(t *testing.T) {
	doc := &DummyDocument{
		Document: Document{
			ID: "testid",
		},
		Foo:  "bar",
		Beep: "bopp",
	}
	if doc.Rev != "" {
		t.Errorf("expected new document to have empty revision but got %s", doc.Rev)
	}
	res, err := db.Post(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Error("expected ok to be true but got false instead")
	}
}

func TestDocumentHead(t *testing.T) {
	head, err := db.Head("testid")
	if err != nil {
		t.Fatal(err)
	}
	if head.StatusCode != 200 {
		t.Errorf("expected status code to be 200 but got %d", head.StatusCode)
	}
}

func TestDocumentGet(t *testing.T) {
	doc := new(DummyDocument)
	err := db.Get(doc, "testid")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Foo != "bar" {
		t.Errorf("expected foo to be bar but got %s", doc.Foo)
	}
	if doc.Beep != "bopp" {
		t.Errorf("expected beep to be bopp but got %s", doc.Beep)
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
	if !res.Ok {
		t.Error("expected ok to be true but got false")
	}
	if res.ID != "testid" {
		t.Errorf("expected id testid but got %s", res.ID)
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
	if !res.Ok {
		t.Error("expected ok to be true but got false")
	}
	if res.ID != "testid" {
		t.Errorf("expected id testid but got %s", res.ID)
	}
}

func TestDocumentPutAttachment(t *testing.T) {
	doc := &DummyDocument{
		Document: Document{
			ID: "testid",
		},
		Foo:  "bar",
		Beep: "bopp",
	}
	res, err := db.PutAttachment(doc, "./test/dog.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Error("expected ok to be true but got false")
	}
	if res.ID != "testid" {
		t.Errorf("expected id testid but got %s", res.ID)
	}
}

// Test added because updating an existing document that had an attachment caused an error.
// After adding more fields to Attachment struct it now works.
func TestUpdateDocumentWithAttachment(t *testing.T) {
	// get existing document
	doc := &DummyDocument{}
	err := db.Get(doc, "testid")
	if err != nil {
		t.Fatal(err)
	}
	// update document with attachment
	doc.Foo = "awesome"
	res, err := db.Put(doc)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Error("expected ok to be true but got false")
	}
	if res.ID != "testid" {
		t.Errorf("expected id testid but got %s", res.ID)
	}
}

func TestDocumentBulkDocs(t *testing.T) {
	// first dummy document
	doc1 := &DummyDocument{
		Foo:  "foo1",
		Beep: "beep1",
	}
	// second dummy document
	doc2 := &DummyDocument{
		Foo:  "foo2",
		Beep: "beep2",
	}
	// slice of dummy document
	docs := []CouchDoc{doc1, doc2}

	res, err := db.Bulk(docs)
	if err != nil {
		t.Fatal(err)
	}
	if !res[0].Ok {
		t.Errorf("expected first ok to be true but got false")
	}
	if !res[1].Ok {
		t.Errorf("expected second ok to be true but got false")
	}
}

func TestAllDocs(t *testing.T) {
	res, err := db.AllDocs(nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalRows != 3 {
		t.Errorf("expected total rows equals 3 but got %d", res.TotalRows)
	}
	if len(res.Rows) != 3 {
		t.Errorf("expected length rows equals 3 but got %d", len(res.Rows))
	}
}

func TestPurge(t *testing.T) {
	dbName := "purge"
	// create database
	if _, err := client.Create(dbName); err != nil {
		t.Error(err)
	}
	db := client.Use(dbName)
	// create documents
	doc := &DummyDocument{
		Foo:  "bar",
		Beep: "bopp",
	}
	postResponse, err := db.Post(doc)
	if err != nil {
		t.Error(err)
	}
	// purge
	req := map[string][]string{
		postResponse.ID: {
			postResponse.Rev,
		},
	}
	purgeResponse, err := db.Purge(req)
	if err != nil {
		t.Error(err)
	}
	if purgeResponse.PurgeSeq != 1 {
		t.Errorf("expected purge seq to be 1 but got %v instead", purgeResponse.PurgeSeq)
	}
	revisions, ok := purgeResponse.Purged[postResponse.ID]
	if !ok {
		t.Error("expected to find entry at post response ID but could not find any")
	}
	if revisions[0] != postResponse.Rev {
		t.Error("expected purged revision to be the same as posted document revision")
	}
	// remove database
	if _, err := client.Delete(dbName); err != nil {
		t.Error(err)
	}
}

func TestSecurity(t *testing.T) {
	dbName := "sec"
	// create database
	if _, err := client.Create(dbName); err != nil {
		t.Error(err)
	}
	db := client.Use(dbName)
	// test putting security document first
	secDoc := SecurityDocument{
		Admins: Element{
			Names: []string{
				"admin1",
			},
			Roles: []string{
				"",
			},
		},
		Members: Element{
			Names: []string{
				"member1",
			},
			Roles: []string{
				"",
			},
		},
	}
	res, err := db.PutSecurity(secDoc)
	if err != nil {
		t.Error(err)
	}
	if !res.Ok {
		t.Error("expected true but got false")
	}
	// test getting security document
	doc, err := db.GetSecurity()
	if err != nil {
		t.Error(err)
	}
	if doc.Admins.Names[0] != "admin1" {
		t.Errorf("expected name admin1 but got %s instead", doc.Admins.Names[0])
	}
	if doc.Members.Names[0] != "member1" {
		t.Errorf("expected name member1 but got %s instead", doc.Members.Names[0])
	}
	// remove database
	if _, err := client.Delete(dbName); err != nil {
		t.Error(err)
	}
}

func TestAfter(t *testing.T) {
	if _, err := client.Delete("dummy"); err != nil {
		t.Fatal(err)
	}
}

// end database tests

// view tests
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
	people := make([]CouchDoc, len(data))
	for index, d := range data {
		people[index] = &Person{
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
	if res.TotalRows != 2 {
		t.Errorf("expected total rows to be 2 but got %d", res.TotalRows)
	}
	if res.Offset != 0 {
		t.Errorf("expected offset to be 0 but got %d", res.Offset)
	}
}

func TestDesignDocumentName(t *testing.T) {
	doc := new(DesignDocument)
	err := dbView.Get(doc, "_design/test")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Name() != "test" {
		t.Errorf("expected name to be test but got %s", doc.Name())
	}
}

func TestDesignDocumentView(t *testing.T) {
	doc := new(DesignDocument)
	err := dbView.Get(doc, "_design/test")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := doc.Views["foo"]; !ok {
		t.Error("expected foo mapping function to exists but it does not")
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
		t.Errorf("expected only one row but got %d", len(res.Rows))
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
		t.Errorf("expected only one row but got %d", len(res.Rows))
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
		t.Errorf("expected only one row but got %d", len(res.Rows))
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

// test utils

// mimeType()
var mimeTypeTests = []struct {
	in  string
	out string
}{
	{"image.jpg", "image/jpeg"},
	{"presentation.pdf", "application/pdf"},
	{"file.text", "text/plain; charset=utf-8"},
	{"archive.zip", "application/zip"},
	{"movie.avi", "video/x-msvideo"},
}

func TestMimeType(t *testing.T) {
	for _, tt := range mimeTypeTests {
		actual := mimeType(tt.in)
		if actual != tt.out {
			t.Errorf("mimeType(%s): expected %s, actual %s", tt.in, tt.out, actual)
		}
	}
}
