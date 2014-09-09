package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
)

type Database struct {
	Url string
}

// Head request.
func (db *Database) Head(id string) (*http.Response, error) {
	return http.Head(db.Url + id)
}

// Get document.
func (db *Database) Get(doc CouchDoc, id string) error {
	body, err := request("GET", db.Url+id, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, doc)
}

// Put document.
func (db *Database) Put(doc CouchDoc) (*DocumentResponse, error) {
	res, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	document := doc.GetDocument()
	body, err := request("PUT", db.Url+document.Id, bytes.NewReader(res))
	if err != nil {
		return nil, err
	}
	return newDocumentResponse(body)
}

// Post document.
func (db *Database) Post(doc CouchDoc) (*DocumentResponse, error) {
	res, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	body, err := request("POST", db.Url, bytes.NewReader(res))
	if err != nil {
		return nil, err
	}
	return newDocumentResponse(body)
}

// Delete document.
func (db *Database) Delete(doc CouchDoc) (*DocumentResponse, error) {
	document := doc.GetDocument()
	body, err := request("DELETE", db.Url+document.Id+"?rev="+document.Rev, nil)
	if err != nil {
		return nil, err
	}
	return newDocumentResponse(body)
}

// Put attachment.
func (db *Database) PutAttachment(doc CouchDoc, path string) (*DocumentResponse, error) {

	// target url
	document := doc.GetDocument()
	url := db.Url + document.Id

	// get file from disk
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// create new writer
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	// create first "application/json" document part
	err = writeJSON(document, writer, file)
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
	req, err := http.NewRequest("PUT", url, &buffer)
	if err != nil {
		return nil, err
	}
	contentType := fmt.Sprintf("multipart/related; boundary=%q", writer.Boundary())
	req.Header.Set("Content-Type", contentType)

	// do http request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return newDocumentResponse(body)
}
