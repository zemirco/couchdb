package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/segmentio/pointer"
)

var client *Client

const (
	// The folllowing characters are valid in databse names: /_$()-+
	// However there is a known bug in couchdb with + (skipping + for testing)
	// https://issues.apache.org/jira/browse/COUCHDB-1580
	validSpecialCharacter = "/_$()-/"
	invalidCharacters     = "."
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
	name, err := RandDBName(5)
	if err != nil {
		t.Fatal(err)
	}
	status, err := client.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ok {
		t.Errorf("expected ok to be true got false")
	}
	if _, err := client.Delete(name); err != nil {
		t.Fatal(err)
	}
}

func TestCreateWithSpecialCharacters(t *testing.T) {
	name, err := RandDBName(5)
	if err != nil {
		t.Fatal(err)
	}
	name = name + validSpecialCharacter + name
	status, err := client.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ok {
		t.Errorf("expected ok to be true got false")
	}
	if _, err := client.Delete(name); err != nil {
		t.Fatal(err)
	}
}

func TestCreateFailWithInvalidCharacters(t *testing.T) {
	name, err := RandDBName(5)
	if err != nil {
		t.Fatal(err)
	}
	name = name + invalidCharacters + name
	_, err = client.Create(name)
	if err == nil {
		t.Errorf("Database created with invalid characters")
	}

}

func TestCreateFail(t *testing.T) {
	name, err := RandDBName(5)
	if err != nil {
		t.Fatal(err)
	}
	// create db first time
	status, err := client.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ok {
		t.Errorf("expected ok to be true got false")
	}
	// try to create db again
	if _, err := client.Create(name); err == nil {
		t.Fatal("creating duplicate database should return an error")
	}
	if couchdbError, ok := err.(*Error); ok {
		if couchdbError.StatusCode != http.StatusPreconditionFailed {
			t.Fatal("creating duplicate database should return an error")
		}
	}
	if _, err := client.Delete(name); err != nil {
		t.Fatal(err)
	}
}

func TestCreateUser(t *testing.T) {
	name, err := RandDBName(5)
	if err != nil {
		t.Fatal(err)
	}
	user := NewUser(name, "password", []string{})
	res, err := client.CreateUser(user)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Errorf("expected ok to be true got false")
	}
	if res.ID != "org.couchdb.user:"+name {
		t.Errorf("expected res id org.couchdb.user:%s but got %s", name, res.ID)
	}
	u, err := client.GetUser(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteUser(u); err != nil {
		t.Fatal(err)
	}
}

func TestCreateSession(t *testing.T) {
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create user first
	user := NewUser(name, "password", []string{})
	if _, err := client.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	// now create session
	res, err := client.CreateSession(name, "password")
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Errorf("expected ok to be true got false")
	}
	if res.Name != name {
		t.Errorf("expected res name %s but got %s", name, res.Name)
	}
	// remove user after test
	u, err := client.GetUser(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteUser(u); err != nil {
		t.Fatal(err)
	}
}

func TestGetSession(t *testing.T) {
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create user first
	user := NewUser(name, "password", []string{})
	if _, err := client.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	// now create session
	if _, err := client.CreateSession(name, "password"); err != nil {
		t.Fatal(err)
	}
	// get session
	session, err := client.GetSession()
	if err != nil {
		t.Fatal(err)
	}
	if !session.Ok {
		t.Errorf("expected ok to be true got false")
	}
	if session.UserContext.Name != name {
		t.Errorf("expected user context name %s but got %s", name, session.UserContext.Name)
	}
	// remove user after test
	u, err := client.GetUser(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteUser(u); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteSession(t *testing.T) {
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create user first
	user := NewUser(name, "password", []string{})
	if _, err := client.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	// create session
	if _, err := client.CreateSession(name, "password"); err != nil {
		t.Fatal(err)
	}
	// now delete session
	res, err := client.DeleteSession()
	if err != nil {
		t.Fatal(err)
	}
	if !res.Ok {
		t.Errorf("expected ok to be true got false")
	}
	// remove user after test
	u, err := client.GetUser(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.DeleteUser(u); err != nil {
		t.Fatal(err)
	}
}

func TestGetUser(t *testing.T) {
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create user first
	user := NewUser(name, "password", []string{})
	if _, err := client.CreateUser(user); err != nil {
		t.Fatal(err)
	}
	// get user
	u, err := client.GetUser(name)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != name {
		t.Errorf("expected name %s but got %s", name, user.Name)
	}
	if u.Type != "user" {
		t.Errorf("expected type user but got %s", user.Type)
	}
	if u.Iterations != 10 {
		t.Errorf("expected 10 iterations but got %d", user.Iterations)
	}
	// delete user after test
	if _, err := client.DeleteUser(u); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteUser(t *testing.T) {
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create user first
	u := NewUser(name, "password", []string{})
	if _, err := client.CreateUser(u); err != nil {
		t.Fatal(err)
	}
	user, err := client.GetUser(name)
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
	if res.ID != "org.couchdb.user:"+name {
		t.Errorf("expected res id to be org.couchdb.user:%s but got %s", name, res.ID)
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
	name, err := RandDBName(5)
	if err != nil {
		t.Fatal(err)
	}
	// create db first time
	if _, err := client.Create(name); err != nil {
		t.Fatal(err)
	}
	// delete database
	status, err := client.Delete(name)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Ok {
		t.Error("expected ok to be true but got false")
	}
}

func TestDeleteFail(t *testing.T) {
	name, err := RandDBName(5)
	if err != nil {
		t.Fatal(err)
	}
	// create db first
	if _, err := client.Create(name); err != nil {
		t.Fatal(err)
	}
	// delete db first time
	if _, err := client.Delete(name); err != nil {
		t.Fatal(err)
	}
	// delete db second time
	if _, err := client.Delete("dummy"); err == nil {
		t.Fatal("should not delete non existing database")
	}
	if couchdbError, ok := err.(*Error); ok {
		if couchdbError.StatusCode != http.StatusNotFound {
			t.Fatal("should not delete non existing database")
		}
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
	r, err := client.Replicate(req)
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
		Language: langJavaScript,
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
	if _, err := client.Replicate(req); err != nil {
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
	if _, err := client.Replicate(req); err != nil {
		t.Error(err)
	}
	tasks, err := client.ActiveTasks()
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

func TestDocumentPost(t *testing.T) {
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
	// use database
	doc := &DummyDocument{
		Document: Document{
			ID: "test" + validSpecialCharacter + "id",
		},
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
	// remove database
	if _, err := client.Delete(name); err != nil {
		t.Fatal(err)
	}
}

func TestDocumentHead(t *testing.T) {
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
	// create document
	doc := &DummyDocument{
		Document: Document{
			ID: "testid",
		},
	}
	if _, err := db.Post(doc); err != nil {
		t.Fatal(err)
	}
	// check head
	head, err := db.Head("testid")
	if err != nil {
		t.Fatal(err)
	}
	if head.StatusCode != 200 {
		t.Errorf("expected status code to be 200 but got %d", head.StatusCode)
	}
}

func TestDocumentGet(t *testing.T) {
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
	// create document
	doc := &DummyDocument{
		Document: Document{
			ID: "testid",
		},
		Foo:  "bar",
		Beep: "bopp",
	}
	if _, err := db.Post(doc); err != nil {
		t.Fatal(err)
	}
	// now test getting the document
	d := new(DummyDocument)
	if err := db.Get(d, "testid"); err != nil {
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
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
	// create document
	doc := &DummyDocument{
		Document: Document{
			ID: "testid",
		},
		Foo:  "bar",
		Beep: "bopp",
	}
	if _, err := db.Post(doc); err != nil {
		t.Fatal(err)
	}
	// get document
	d := new(DummyDocument)
	if err := db.Get(d, "testid"); err != nil {
		t.Fatal(err)
	}
	// change document
	d.Foo = "baz"
	res, err := db.Put(d)
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
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
	// create document
	doc := &DummyDocument{
		Document: Document{
			ID: "testid",
		},
		Foo:  "bar",
		Beep: "bopp",
	}
	if _, err := db.Post(doc); err != nil {
		t.Fatal(err)
	}
	// get document
	d := new(DummyDocument)
	if err := db.Get(d, "testid"); err != nil {
		t.Fatal(err)
	}
	// delete document
	res, err := db.Delete(d)
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
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
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
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
	doc := &DummyDocument{
		Document: Document{
			ID: "testid",
		},
		Foo:  "bar",
		Beep: "bopp",
	}
	if _, err := db.PutAttachment(doc, "./test/dog.jpg"); err != nil {
		t.Fatal(err)
	}
	// get existing document
	d := &DummyDocument{}
	if err := db.Get(d, "testid"); err != nil {
		t.Fatal(err)
	}
	// update document with attachment
	d.Foo = "awesome"
	res, err := db.Put(d)
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
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
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
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
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
	doc3 := &DummyDocument{
		Foo:  "foo3",
		Beep: "beep3",
	}
	// slice of dummy document
	docs := []CouchDoc{doc1, doc2, doc3}
	if _, err := db.Bulk(docs); err != nil {
		t.Error(err)
	}
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

func TestView(t *testing.T) {
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
	// create database
	design := &DesignDocument{
		Document: Document{
			ID: "_design/test",
		},
		Language: langJavaScript,
		Views: map[string]DesignDocumentView{
			"foo": {
				Map: `
					function(doc) {
						if (doc.type === 'data') {
							emit(doc.foo);
						}
					}
				`,
			},
			"int": {
				Map: `
					function(doc) {
						if (doc.type === 'data') {
							emit([doc.foo, doc.age]);
						}
					}
				`,
			},
			"complex": {
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
	if _, err := db.Post(design); err != nil {
		t.Fatal(err)
	}
	// create design document for person
	designPerson := DesignDocument{
		Document: Document{
			ID: "_design/person",
		},
		Language: langJavaScript,
		Views: map[string]DesignDocumentView{
			"ageByGender": {
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
	if _, err := db.Post(&designPerson); err != nil {
		t.Fatal(err)
	}
	// create dummy data
	doc1 := &DataDocument{
		Type: "data",
		Foo:  "foo1",
		Beep: "beep1",
		Age:  10,
	}
	if _, err := db.Post(doc1); err != nil {
		t.Fatal(err)
	}
	doc2 := &DataDocument{
		Type: "data",
		Foo:  "foo2",
		Beep: "beep2",
		Age:  20,
	}
	if _, err := db.Post(doc2); err != nil {
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
	if _, err := db.Bulk(people); err != nil {
		t.Fatal(err)
	}

	// run sub test
	t.Run("get", func(t *testing.T) {
		view := db.View("test")
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
	})

	t.Run("design document name", func(t *testing.T) {
		doc := new(DesignDocument)
		err := db.Get(doc, "_design/test")
		if err != nil {
			t.Fatal(err)
		}
		if doc.Name() != "test" {
			t.Errorf("expected name to be test but got %s", doc.Name())
		}
	})

	t.Run("design document view", func(t *testing.T) {
		doc := new(DesignDocument)
		err := db.Get(doc, "_design/test")
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := doc.Views["foo"]; !ok {
			t.Error("expected foo mapping function to exists but it does not")
		}
	})

	t.Run("get with query parameters", func(t *testing.T) {
		view := db.View("test")
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
	})

	t.Run("get with start and end key", func(t *testing.T) {
		view := db.View("test")
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
	})

	t.Run("get with integer", func(t *testing.T) {
		view := db.View("test")
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
	})

	t.Run("get with reduce", func(t *testing.T) {
		view := db.View("person")
		params := QueryParameters{}
		res, err := view.Get("ageByGender", params)
		if err != nil {
			t.Fatal(err)
		}
		ageTotalSum := res.Rows[0].Value.(float64)
		if ageTotalSum != 372 {
			t.Fatalf("expected age 372 but got %v", ageTotalSum)
		}
	})

	t.Run("get with reduce and group", func(t *testing.T) {
		view := db.View("person")
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
	})

	t.Run("get without reduce", func(t *testing.T) {
		view := db.View("person")
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
	})

	t.Run("post", func(t *testing.T) {
		view := db.View("person")
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
	})

	// remove database
	if _, err := client.Delete(name); err != nil {
		t.Error(err)
	}
}

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

func TestDiff(t *testing.T) {
	tests := []struct {
		desc      string
		cache     []DesignDocument
		database  []DesignDocument
		additions int
		changes   int
		deletions int
	}{
		{
			desc: "database is empty",
			cache: []DesignDocument{
				{
					Document: Document{
						ID: "_design/player",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
				{
					Document: Document{
						ID: "_design/user",
					},
					Views: map[string]DesignDocumentView{
						"byToken": {
							Map: "function() {}",
						},
					},
				},
			},
			database:  []DesignDocument{},
			additions: 2,
			changes:   0,
			deletions: 0,
		},
		{
			desc:  "cache is empty",
			cache: []DesignDocument{},
			database: []DesignDocument{
				{
					Document: Document{
						ID:  "_design/player",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
				{
					Document: Document{
						ID:  "_design/user",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"byToken": {
							Map: "function() {}",
						},
					},
				},
			},
			additions: 0,
			changes:   0,
			deletions: 2,
		},
		{
			desc: "cache and database are equal",
			cache: []DesignDocument{
				{
					Document: Document{
						ID: "_design/player",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
			},
			database: []DesignDocument{
				{
					Document: Document{
						ID:  "_design/player",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
			},
			additions: 0,
			changes:   0,
			deletions: 0,
		},
		{
			desc: "database out of date",
			cache: []DesignDocument{
				{
					Document: Document{
						ID: "_design/player",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function(doc) {}",
						},
					},
				},
			},
			database: []DesignDocument{
				{
					Document: Document{
						ID:  "_design/player",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
			},
			additions: 0,
			changes:   1,
			deletions: 0,
		},
		{
			desc: "database has too many entries",
			cache: []DesignDocument{
				{
					Document: Document{
						ID: "_design/player",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
			},
			database: []DesignDocument{
				{
					Document: Document{
						ID:  "_design/player",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
				{
					Document: Document{
						ID:  "_design/user",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"byToken": {
							Map: "function() {}",
						},
					},
				},
			},
			additions: 0,
			changes:   0,
			deletions: 1,
		},
		{
			desc: "database is missing a design document",
			cache: []DesignDocument{
				{
					Document: Document{
						ID: "_design/player",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
				{
					Document: Document{
						ID: "_design/user",
					},
					Views: map[string]DesignDocumentView{
						"byToken": {
							Map: "function() {}",
						},
					},
				},
			},
			database: []DesignDocument{
				{
					Document: Document{
						ID:  "_design/player",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
			},
			additions: 1,
			changes:   0,
			deletions: 0,
		},
		{
			desc: "database is missing a design document and has an old version of anther one",
			cache: []DesignDocument{
				{
					Document: Document{
						ID: "_design/player",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function(doc) {}",
						},
					},
				},
				{
					Document: Document{
						ID: "_design/user",
					},
					Views: map[string]DesignDocumentView{
						"byToken": {
							Map: "function() {}",
						},
					},
				},
			},
			database: []DesignDocument{
				{
					Document: Document{
						ID:  "_design/player",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
			},
			additions: 1,
			changes:   1,
			deletions: 0,
		},
		{
			desc: "database has too many design documents and an old version of another one",
			cache: []DesignDocument{
				{
					Document: Document{
						ID: "_design/player",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function(doc) {}",
						},
					},
				},
			},
			database: []DesignDocument{
				{
					Document: Document{
						ID:  "_design/player",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"byName": {
							Map: "function() {}",
						},
					},
				},
				{
					Document: Document{
						ID:  "_design/user",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"byToken": {
							Map: "function() {}",
						},
					},
				},
			},
			additions: 0,
			changes:   1,
			deletions: 1,
		},
		{
			desc:  "database has internal design document which should not be deleted",
			cache: []DesignDocument{},
			database: []DesignDocument{
				{
					Document: Document{
						ID:  "_design/_auth",
						Rev: "abc",
					},
					Views: map[string]DesignDocumentView{
						"validate_doc_update": {
							Map: "function() {}",
						},
					},
				},
			},
			additions: 0,
			changes:   0,
			deletions: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			difference := diff(test.cache, test.database)
			if len(difference.additions) != test.additions {
				t.Errorf(
					"exp %d additions but got %d",
					test.additions,
					len(difference.additions),
				)
			}
			if len(difference.changes) != test.changes {
				t.Errorf(
					"exp %d changes but got %d",
					test.changes,
					len(difference.changes),
				)
			}
			if len(difference.deletions) != test.deletions {
				t.Errorf(
					"exp %d deletions but got %d",
					test.deletions,
					len(difference.deletions),
				)
			}
		})
	}
}

// remove all white space and line breaks from string
func clean(s string) string {
	return strings.Replace(strings.Replace(s, " ", "", -1), "\n", "", -1)
}

func TestParse(t *testing.T) {
	docs, err := client.Parse(filepath.Join("example", "design"))
	if err != nil {
		t.Error(err)
	}
	if len(docs) != 2 {
		t.Errorf("expected 2 design documents but got %d", len(docs))
	}
	// check first design document
	player := docs[0]
	if player.Name() != "player" {
		t.Errorf("expected 1st design document to be player but got %s", docs[0].Name())
	}
	byAge, ok := player.Views["byAge"]
	if !ok {
		t.Error("cannot find byAge view in first design document")
	}
	byAgeMap := "function(doc){if(doc.type==='player'){emit(doc.age)}}"
	if byAgeMap != clean(byAge.Map) {
		t.Errorf("expected byAge map %s but got %s", byAgeMap, clean(byAge.Map))
	}
	byAgeReduce := "function(keys,values,rereduce){returnsum(values)}"
	if byAgeReduce != clean(byAge.Reduce) {
		t.Errorf("expected byAge reduce %s but got %s", byAgeReduce, clean(byAge.Reduce))
	}
	byName, ok := player.Views["byName"]
	if !ok {
		t.Error("cannot find byName view in first design document")
	}
	byNameMap := "function(doc){if(doc.type==='player'){emit(doc.name)}}"
	if byNameMap != clean(byName.Map) {
		t.Errorf("expected byName map %s but got %s", byNameMap, clean(byName.Map))
	}
	if byName.Reduce != "" {
		t.Errorf("expected byName reduce to be empty but got %s", byName.Reduce)
	}
	// check second design document
	user := docs[1]
	if user.Name() != "user" {
		t.Errorf("expected 2nd design document to be user but got %s", user.Name())
	}
	byEmail, ok := user.Views["byEmail"]
	if !ok {
		t.Error("cannot find byEmail view in second design document")
	}
	byEmailMap := "function(doc){if(doc.type==='user'){emit(doc.email)}}"
	if byEmailMap != clean(byEmail.Map) {
		t.Errorf("expected byEmail map %s but got %s", byEmailMap, clean(byEmail.Map))
	}
	byEmailReduce :=
		"function(keys,values,rereduce){if(rereduce){returnsum(values)}else{returnvalues.length}}"
	if byEmailReduce != clean(byEmail.Reduce) {
		t.Errorf("expected byEmail reduce %s but got %s", byEmailReduce, clean(byEmail.Reduce))
	}
	byUsername, ok := user.Views["byUsername"]
	if !ok {
		t.Error("cannot find byUsername view in second design document")
	}
	byUsernameMap := "function(doc){if(doc.type==='user'){emit(doc.username)}}"
	if byUsernameMap != clean(byUsername.Map) {
		t.Errorf("expected byUsername map %s but got %s", byUsernameMap, clean(byUsername.Map))
	}
	if byUsername.Reduce != "" {
		t.Errorf("expected byUsername reduce to be empty but got %s", byUsername.Reduce)
	}
}

func TestSeed(t *testing.T) {
	// create random database
	name, err := RandDBName(10)
	if err != nil {
		t.Error(err)
	}
	// create database
	if _, err := client.Create(name); err != nil {
		t.Error(err)
	}
	db := client.Use(name)
	docs, err := client.Parse(filepath.Join("example", "design"))
	if err != nil {
		t.Error(err)
	}
	// add some design documents
	// ---
	// player
	// |- byAge
	// |- byName
	// user
	// |- byEmail
	// |- byUsername
	if err := db.Seed(docs); err != nil {
		t.Error(err)
	}
	// simulate player design document has changed on disk
	changedPlayer := DesignDocument{
		Document: Document{
			ID: "_design/player",
		},
		Language: langJavaScript,
		Views: map[string]DesignDocumentView{
			"byName": {
				Map: "function() {}",
			},
		},
	}
	// add new design document
	car := DesignDocument{
		Document: Document{
			ID: "_design/car",
		},
		Language: langJavaScript,
		Views: map[string]DesignDocumentView{
			"byColor": {
				Map: "function() {/* byColor map */}",
			},
			"byBrand": {
				Map:    "function() {/* byBrand map */}",
				Reduce: "function() {/* byBrand reduce */}",
			},
		},
	}
	// seed should
	// - remove "_design/user" document
	// - update "_design/player" document
	// - add "_design/car" document
	if err := db.Seed([]DesignDocument{changedPlayer, car}); err != nil {
		t.Error(err)
	}
	// make sure "_design/player" has been updated
	var playerDesignDoc DesignDocument
	if err := db.Get(&playerDesignDoc, "_design/player"); err != nil {
		t.Error(err)
	}
	if !strings.HasPrefix(playerDesignDoc.Rev, "2-") {
		t.Errorf("exp. revision to start with 2- but got %s", playerDesignDoc.Rev)
	}
	byNameMap := playerDesignDoc.Views["byName"].Map
	if byNameMap != "function() {}" {
		t.Errorf("exp. byName map function to be empty but got %s", byNameMap)
	}
	// make sure "_design/user" has been deleted
	var userDesignDoc DesignDocument
	if err := db.Get(&userDesignDoc, "_design/user"); err != nil {
		if cerr, ok := err.(*Error); ok {
			if cerr.StatusCode != http.StatusNotFound {
				t.Error(err)
			}
		} else {
			t.Error(err)
		}
	}
	// make sure "_design/car" exists
	var carDesignDoc DesignDocument
	if err := db.Get(&carDesignDoc, "_design/car"); err != nil {
		t.Error(err)
	}
	byColorMap := carDesignDoc.Views["byColor"].Map
	if byColorMap != "function() {/* byColor map */}" {
		t.Errorf("expected byColor map function but got %s", byColorMap)
	}
	byBrand := carDesignDoc.Views["byBrand"]
	if byBrand.Map != "function() {/* byBrand map */}" {
		t.Errorf("expected byBrand map function but got %s", byBrand.Map)
	}
	if byBrand.Reduce != "function() {/* byBrand reduce */}" {
		t.Errorf("expected byBrand reduce function but got %s", byBrand.Reduce)
	}
	if _, err := client.Delete(name); err != nil {
		t.Error(err)
	}
}
