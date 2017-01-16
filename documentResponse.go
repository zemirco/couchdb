package couchdb

// DocumentResponse is response for multipart/related file upload.
type DocumentResponse struct {
	Ok  bool
	ID  string
	Rev string
}
