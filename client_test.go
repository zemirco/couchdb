package couchdb

import (
	"net/http"
	"reflect"
	"testing"
)

var client, _ = NewClient("http://127.0.0.1:5984/")

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
}

func TestCreateFail(t *testing.T) {
	_, err := client.Create("dummy")
	if err == nil {
		t.Fatal("should not create duplicate database")
	}
	if couchdbError, ok := err.(*Error); ok {
		if couchdbError.StatusCode != http.StatusPreconditionFailed {
			t.Fatal("should not create duplicate database")
		}
	}
}

func TestCreateUser(t *testing.T) {
	user := NewUser("john", "password", []string{})
	res, err := client.CreateUser(user)
	if err != nil {
		t.Fatal(err)
	}
	if res.Ok == false || res.ID != "org.couchdb.user:john" {
		t.Error("create user error")
	}
}

func TestCreateSession(t *testing.T) {
	res, err := client.CreateSession("john", "password")
	if err != nil {
		t.Fatal(err)
	}
	if res.Ok == false || res.Name != "john" {
		t.Error("create session error")
	}
}

func TestGetSession(t *testing.T) {
	session, err := client.GetSession()
	if err != nil {
		t.Fatal(err)
	}
	if session.Ok == false || session.UserContext.Name != "john" {
		t.Error("get session error")
	}
}

func TestDeleteSession(t *testing.T) {
	res, err := client.DeleteSession()
	if err != nil {
		t.Fatal(err)
	}
	if res.Ok == false {
		t.Error("delete session error")
	}
}

func TestGetUser(t *testing.T) {
	user, err := client.GetUser("john")
	if err != nil {
		t.Fatal(err)
	}
	if user.Name != "john" || user.Type != "user" || user.Iterations != 10 {
		t.Error("get user error")
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
	if res.Ok == false || res.ID != "org.couchdb.user:john" {
		t.Error("delete user error")
	}
}

func TestGetSessionAdmin(t *testing.T) {
	session, err := client.GetSession()
	if err != nil {
		t.Fatal(err)
	}
	if session.Ok == false {
		t.Error("session response is false")
	}
	roles := []string{"_admin"}
	if reflect.DeepEqual(roles, session.UserContext.Roles) == false {
		t.Error("session roles are wrong")
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
	if db.URL != "http://127.0.0.1:5984/_users/" {
		t.Error("use error")
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
	res, err := client.Create(name)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", res)
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
	t.Logf("%#v", r)
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
	allDocs, err := db.AllDocs()
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
	res, err := c.Replicate(req)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", res)
	tasks, err := c.ActiveTasks()
	if err != nil {
		t.Error(err)
	}
	t.Logf("%#v", tasks)
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
