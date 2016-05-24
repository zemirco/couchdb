package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"reflect"
)

// Database performs actions on certain database
type Database struct {
	*Client
	URL string
}

// AllDocs returns all documents in selected database.
func (db *Database) AllDocs() (*ViewResponse, error) {
	url := fmt.Sprintf("%s_all_docs", db.URL)
	body, err := db.Client.request(http.MethodGet, url, nil, "")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newViewResponse(body)
}

// Head request.
func (db *Database) Head(id string) (*http.Response, error) {
	return http.Head(db.URL + id)
}

// Get document.
func (db *Database) Get(doc CouchDoc, id string) error {
	url := fmt.Sprintf("%s%s", db.URL, id)
	body, err := db.Client.request(http.MethodGet, url, nil, "application/json")
	if err != nil {
		return err
	}
	defer body.Close()
	return json.NewDecoder(body).Decode(doc)
}

// Put document.
func (db *Database) Put(doc CouchDoc) (*DocumentResponse, error) {
	url := fmt.Sprintf("%s%s", db.URL, doc.GetID())
	res, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(res)
	body, err := db.Client.request(http.MethodPut, url, data, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDocumentResponse(body)
}

// Post document.
func (db *Database) Post(doc CouchDoc) (*DocumentResponse, error) {
	res, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(res)
	body, err := db.Client.request(http.MethodPost, db.URL, data, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDocumentResponse(body)
}

// Delete document.
func (db *Database) Delete(doc CouchDoc) (*DocumentResponse, error) {
	url := fmt.Sprintf("%s%s?rev=%s", db.URL, doc.GetID(), doc.GetRev())
	body, err := db.Client.request(http.MethodDelete, url, nil, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDocumentResponse(body)
}

// Purge document.
func (db *Database) Purge(doc CouchDoc) (*DocumentResponse, error) {
	url := fmt.Sprintf("%s%s", db.URL, "_purge")
	res, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	data := bytes.NewReader(res)
	body, err := db.Client.request(http.MethodPost, url, data, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDocumentResponse(body)
}

// PutAttachment adds attachment to document
func (db *Database) PutAttachment(doc CouchDoc, path string) (*DocumentResponse, error) {

	// target url
	url := fmt.Sprintf("%s%s", db.URL, doc.GetID())

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
	body, err := db.Client.request(http.MethodPut, url, &buffer, contentType)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	return newDocumentResponse(body)
}

// Bulk allows to create and update multiple documents
// at the same time within a single request. The basic operation is similar to
// creating or updating a single document, except that you batch
// the document structure and information.
func (db *Database) Bulk(docs interface{}) ([]DocumentResponse, error) {
	// convert to []interface{}
	val := reflect.ValueOf(docs)
	documents := make([]interface{}, val.Len())
	for i := 0; i < val.Len(); i++ {
		documents[i] = val.Index(i).Interface()
	}
	// create bulk docs
	bulk := BulkDoc{
		Docs: documents,
	}
	res, err := json.Marshal(bulk)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s_bulk_docs", db.URL)
	data := bytes.NewReader(res)
	body, err := db.Client.request(http.MethodPost, url, data, "application/json")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	response := []DocumentResponse{}
	return response, json.NewDecoder(body).Decode(&response)

}

// View returns view for given name.
func (db *Database) View(name string) View {
	url := fmt.Sprintf("%s_design/%s/", db.URL, name)
	return View{
		URL:      url,
		Database: db,
	}
}
