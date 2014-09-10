package couchdb

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

// Get mime type from file name.
func mimeType(name string) string {
	ext := filepath.Ext(name)
	return mime.TypeByExtension(ext)
}

// Make HTTP request.
// Treat status code other than 2xx as Error.
func request(method, url string, data io.Reader, contentType string) ([]byte, error) {
	req, err := http.NewRequest(method, url, data)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
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
	// handle CouchDB http errors
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, newError(res, body)
	}
	return body, nil
}

// Convert HTTP response from CouchDB into Error.
func newError(res *http.Response, body []byte) error {
	var error *Error
	err := json.Unmarshal(body, &error)
	if err != nil {
		return err
	}
	error.Method = res.Request.Method
	error.Url = res.Request.URL.String()
	error.StatusCode = res.StatusCode
	return error
}

// Create new CouchDB response for any document method.
func newDocumentResponse(body []byte) (*DocumentResponse, error) {
	var response *DocumentResponse
	return response, json.Unmarshal(body, &response)
}

// Create new CouchDB response for any database method.
func newDatabaseResponse(body []byte) (*DatabaseResponse, error) {
	var response *DatabaseResponse
	return response, json.Unmarshal(body, &response)
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
func writeMultipart(writer *multipart.Writer, file *os.File) error {
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

// Quote all values because CouchDB needs those double quotes in query params.
func quote(values url.Values) url.Values {
	for key, value := range values {
		values.Set(key, strconv.Quote(value[0]))
	}
	return values
}
