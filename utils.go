package couchdb

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"

	"github.com/zemirco/uid"
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

// RandDBName returns random CouchDB database name.
// See the docs for database name rules.
// http://docs.couchdb.org/en/2.0.0/api/database/common.html#put--db
func RandDBName(length int) (string, error) {
	// fastest string concatenation
	var buffer bytes.Buffer
	// generate first character, must be a letter
	first := uid.NewBytes(1, "abcdefghijklmnopqrstuvwxyz")
	if _, err := buffer.WriteString(first); err != nil {
		return "", err
	}
	// generate last characters
	last := uid.NewBytes(length-1, "abcdefghijklmnopqrstuvwxyz0123456789")
	if _, err := buffer.WriteString(last); err != nil {
		return "", err
	}
	return buffer.String(), nil
}
