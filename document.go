package couchdb

// CouchDoc describes interface for every couchdb document.
type CouchDoc interface {
	GetID() string
	GetRev() string
}

// Document is base struct which should be embedded by any other couchdb document.
type Document struct {
	ID          string                `json:"_id,omitempty"`
	Rev         string                `json:"_rev,omitempty"`
	Attachments map[string]Attachment `json:"_attachments,omitempty"`
}

// Attachment describes attachments of a document.
// http://docs.couchdb.org/en/stable/api/document/common.html#attachments
// By using attachments you are also able to upload a document in multipart/related format.
// http://docs.couchdb.org/en/latest/api/document/common.html#creating-multiple-attachments
type Attachment struct {
	ContentType   string  `json:"content_type,omitempty"`
	Data          string  `json:"data,omitempty"`
	Digest        string  `json:"digest,omitempty"`
	EncodedLength float64 `json:"encoded_length,omitempty"`
	Encoding      string  `json:"encoding,omitempty"`
	Length        int64   `json:"length,omitempty"`
	RevPos        float64 `json:"revpos,omitempty"`
	Stub          bool    `json:"stub,omitempty"`
	Follows       bool    `json:"follows,omitempty"`
}

// GetID returns document id
func (d *Document) GetID() string {
	return d.ID
}

// GetRev returns document revision
func (d *Document) GetRev() string {
	return d.Rev
}
