package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/google/go-querystring/query"
)

// DatabaseService is an interface for dealing with a single CouchDB database.
type DatabaseService interface {
	AllDocs(params *QueryParameters) (*ViewResponse, error)
	Head(id string) (*http.Response, error)
	Get(doc CouchDoc, id string) error
	Put(doc CouchDoc) (*DocumentResponse, error)
	Post(doc CouchDoc) (*DocumentResponse, error)
	Delete(doc CouchDoc) (*DocumentResponse, error)
	PutAttachment(doc CouchDoc, path string) (*DocumentResponse, error)
	Bulk(docs []CouchDoc) ([]DocumentResponse, error)
	Purge(req map[string][]string) (*PurgeResponse, error)
	GetSecurity() (*SecurityDocument, error)
	PutSecurity(secDoc SecurityDocument) (*DatabaseResponse, error)
	View(name string) ViewService
}

// Database performs actions on certain database
type Database struct {
	Client *Client
	Name   string
}

// AllDocs returns all documents in selected database.
// http://docs.couchdb.org/en/latest/api/database/bulk-api.html
func (db *Database) AllDocs(params *QueryParameters) (*ViewResponse, error) {
	q, err := query.Values(params)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s_all_docs?%s", db.Name, q.Encode())
	res, err := db.Client.Request(http.MethodGet, u, nil, "")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response ViewResponse
	return &response, json.NewDecoder(res.Body).Decode(&response)
}

// Head request.
func (db *Database) Head(id string) (*http.Response, error) {
	u := fmt.Sprintf("%s%s", db.Name, id)
	body, err := db.Client.Request(http.MethodHead, u, nil, "")
	if err != nil {
		return nil, err
	}
	return body, nil
}

// Get document.
func (db *Database) Get(doc CouchDoc, id string) error {
	u := fmt.Sprintf("%s%s", db.Name, id)
	res, err := db.Client.Request(http.MethodGet, u, nil, "application/json")
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return json.NewDecoder(res.Body).Decode(doc)
}

// Put document.
func (db *Database) Put(doc CouchDoc) (*DocumentResponse, error) {
	u := fmt.Sprintf("%s%s", db.Name, doc.GetID())
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(doc); err != nil {
		return nil, err
	}
	res, err := db.Client.Request(http.MethodPut, u, &b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response DocumentResponse
	return &response, json.NewDecoder(res.Body).Decode(&response)
}

// Post document.
func (db *Database) Post(doc CouchDoc) (*DocumentResponse, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(doc); err != nil {
		return nil, err
	}
	res, err := db.Client.Request(http.MethodPost, db.Name, &b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response DocumentResponse
	return &response, json.NewDecoder(res.Body).Decode(&response)
}

// Delete document.
func (db *Database) Delete(doc CouchDoc) (*DocumentResponse, error) {
	u := fmt.Sprintf("%s%s?rev=%s", db.Name, doc.GetID(), doc.GetRev())
	res, err := db.Client.Request(http.MethodDelete, u, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response DocumentResponse
	return &response, json.NewDecoder(res.Body).Decode(&response)
}

// PutAttachment adds attachment to document
func (db *Database) PutAttachment(doc CouchDoc, path string) (*DocumentResponse, error) {

	// target url
	u := fmt.Sprintf("%s%s", db.Name, doc.GetID())

	// get file from disk
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// create new writer
	buffer := bytes.Buffer{}
	writer := multipart.NewWriter(&buffer)

	// create first "application/json" document part
	document := Document{
		ID:  doc.GetID(),
		Rev: doc.GetRev(),
	}
	err = writeJSON(&document, writer, file)
	if err != nil {
		return nil, err
	}

	// write actual file content to multipart message
	err = writeMultipart(writer, file)
	if err != nil {
		return nil, err
	}

	// finish multipart message and write trailing boundary
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	// create http request
	contentType := fmt.Sprintf("multipart/related; boundary=%q", writer.Boundary())
	res, err := db.Client.Request(http.MethodPut, u, &buffer, contentType)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var response DocumentResponse
	return &response, json.NewDecoder(res.Body).Decode(&response)
}

// Bulk allows to create and update multiple documents
// at the same time within a single request. The basic operation is similar to
// creating or updating a single document, except that you batch
// the document structure and information.
func (db *Database) Bulk(docs []CouchDoc) ([]DocumentResponse, error) {
	bulk := BulkDoc{
		Docs: docs,
	}
	u := fmt.Sprintf("%s_bulk_docs", db.Name)
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(bulk); err != nil {
		return nil, err
	}
	res, err := db.Client.Request(http.MethodPost, u, &b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	response := []DocumentResponse{}
	return response, json.NewDecoder(res.Body).Decode(&response)

}

// View returns view for given name.
func (db *Database) View(name string) ViewService {
	u := fmt.Sprintf("%s_design/%s/", db.Name, name)
	return &View{
		URL:    u,
		Client: db.Client,
	}
}

// PurgeResponse is response from POST request to the _purge URL.
type PurgeResponse struct {
	PurgeSeq float64 `json:"purge_seq"`
	Purged   map[string][]string
}

// Purge permanently removes the references to deleted documents from the database.
//
// http://docs.couchdb.org/en/1.6.1/api/database/misc.html
func (db *Database) Purge(req map[string][]string) (*PurgeResponse, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(req); err != nil {
		return nil, err
	}
	res, err := db.Client.Request(http.MethodPost, db.Name+"_purge", &b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	response := &PurgeResponse{}
	return response, json.NewDecoder(res.Body).Decode(&response)
}

// Element is single element inside Admins/Members in security document.
type Element struct {
	Names []string `json:"names"`
	Roles []string `json:"roles"`
}

// SecurityDocument describes document _security document.
type SecurityDocument struct {
	Admins  Element `json:"admins"`
	Members Element `json:"members"`
}

// GetSecurity returns security document.
// http://docs.couchdb.org/en/latest/api/database/security.html
func (db *Database) GetSecurity() (*SecurityDocument, error) {
	u := fmt.Sprintf("%s_security", db.Name)
	res, err := db.Client.Request(http.MethodGet, u, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var secDoc SecurityDocument
	return &secDoc, json.NewDecoder(res.Body).Decode(&secDoc)
}

// PutSecurity sets the security object for the given database.
// http://docs.couchdb.org/en/latest/api/database/security.html#put--db-_security
func (db *Database) PutSecurity(secDoc SecurityDocument) (*DatabaseResponse, error) {
	u := fmt.Sprintf("%s_security", db.Name)
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(secDoc); err != nil {
		return nil, err
	}
	res, err := db.Client.Request(http.MethodPut, u, &b, "application/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	r := new(DatabaseResponse)
	return r, json.NewDecoder(res.Body).Decode(r)
}
