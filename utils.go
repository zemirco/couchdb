package couchdb

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
)

// Get mime type from file name.
func mimeType(name string) string {
	ext := filepath.Ext(name)
	return mime.TypeByExtension(ext)
}

// Convert HTTP response from CouchDB into Error.
func newError(res *http.Response) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	error := &Error{}
	err = json.Unmarshal(body, &error)
	if err != nil {
		return err
	}
	error.Method = res.Request.Method
	error.URL = res.Request.URL.String()
	error.StatusCode = res.StatusCode
	return error
}

// Create new CouchDB response for any document method.
func newDocumentResponse(body io.Reader) (*DocumentResponse, error) {
	response := &DocumentResponse{}
	return response, json.NewDecoder(body).Decode(&response)
}

// Create new CouchDB response for any database method.
func newDatabaseResponse(body io.Reader) (*DatabaseResponse, error) {
	response := &DatabaseResponse{}
	return response, json.NewDecoder(body).Decode(&response)
}

// Write JSON to multipart/related.
func writeJSON(document *Document, writer *multipart.Writer, file *os.File) error {
	partHeaders := textproto.MIMEHeader{}
	partHeaders.Set("Content-Type", "application/json")
	part, err := writer.CreatePart(partHeaders)
	if err != nil {
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	path := file.Name()

	// make empty map
	document.Attachments = make(map[string]Attachment)
	attachment := Attachment{
		Follows:     true,
		ContentType: mimeType(path),
		Length:      stat.Size(),
	}
	// add attachment to map
	filename := filepath.Base(path)
	document.Attachments[filename] = attachment

	bytes, err := json.Marshal(document)
	if err != nil {
		return err
	}

	_, err = part.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

// Write actual file content to multipart/related.
func writeMultipart(writer *multipart.Writer, file io.Reader) error {
	part, err := writer.CreatePart(textproto.MIMEHeader{})
	if err != nil {
		return err
	}

	// copy file content into multipart message
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}

	return nil
}
