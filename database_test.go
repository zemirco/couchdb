package couchdb

import "testing"

type DummyDocument struct {
	Document
	Foo  string `json:"foo"`
	Beep string `json:"beep"`
}

// init client and test database
var c, _ = NewClient("http://127.0.0.1:5984/")
var db = c.Use("dummy")

func TestBefore(t *testing.T) {
	_, err := client.Create("dummy")
	if err != nil {
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
	if res.ID != "testid" || res.Ok == false {
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
	if res.ID != "testid" || res.Ok == false {
		t.Error("delete document response error")
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
	if res.ID != "testid" || res.Ok == false {
		t.Error("put attachment error")
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
	if res.ID != "testid" || res.Ok == false {
		t.Error("put document response error")
	}
}

func TestDocumentBulkDocs(t *testing.T) {
	// first dummy document
	doc1 := DummyDocument{
		Foo:  "foo1",
		Beep: "beep1",
	}
	// second dummy document
	doc2 := DummyDocument{
		Foo:  "foo2",
		Beep: "beep2",
	}
	// slice of dummy document
	docs := []DummyDocument{doc1, doc2}

	res, err := db.Bulk(docs)
	if err != nil {
		t.Fatal(err)
	}
	if res[0].Ok != true || res[1].Ok != true {
		t.Error("bulk docs error")
	}
}

func TestAllDocs(t *testing.T) {
	res, err := db.AllDocs()
	if err != nil {
		t.Fatal(err)
	}
	if res.TotalRows != 3 {
		t.Errorf("expected total rows equals 3 but got %v", res.TotalRows)
	}
	if len(res.Rows) != 3 {
		t.Errorf("expected length rows equals 3 but got %v", len(res.Rows))
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
	t.Logf("%#v", postResponse)
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
	t.Logf("%#v", purgeResponse)
	if purgeResponse.PurgeSeq != 1 {
		t.Errorf("expected purge seq to be 1 but got %d instead", purgeResponse.PurgeSeq)
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
	if res.Ok != true {
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
	t.Log("deleting dummy database")
	_, err := client.Delete("dummy")
	if err != nil {
		t.Fatal(err)
	}
}
