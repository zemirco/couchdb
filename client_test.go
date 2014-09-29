package couchdb

import (
	"reflect"
	"regexp"
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

func TestLog(t *testing.T) {
	log, err := client.Log()
	if err != nil {
		t.Fatal(err)
	}
	valid := regexp.MustCompile("[info]")
	if valid.MatchString(log) == false {
		t.Error("invalid log")
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
	res, err := client.Create("dummy")
	if err != nil {
		t.Fatal(err)
	}
	if res.Ok == true {
		t.Fatal("should not create duplicate database")
	}
}

func TestCreateUser(t *testing.T) {
	user := NewUser("john", "password", []string{})
	res, err := client.CreateUser(user)
	if err != nil {
		t.Fatal(err)
	}
	if res.Ok == false || res.Id != "org.couchdb.user:john" {
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
	res, err := client.DeleteUser(*user)
	if err != nil {
		t.Fatal(err)
	}
	if res.Ok == false || res.Id != "org.couchdb.user:john" {
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
	res, err := client.Delete("dummy")
	if err != nil {
		t.Fatal(err)
	}
	if res.Ok == true {
		t.Fatal("should not delete non existing database")
	}
}

func TestUse(t *testing.T) {
	db := client.Use("_users")
	if db.Url != "http://127.0.0.1:5984/_users/" {
		t.Error("use error")
	}
}
