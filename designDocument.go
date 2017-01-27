package couchdb

import "strings"

const langJavaScript = "javascript"

// DesignDocument is a special type of CouchDB document that contains application code.
// http://docs.couchdb.org/en/latest/json-structure.html#design-document
type DesignDocument struct {
	Document
	Language string                        `json:"language,omitempty"`
	Views    map[string]DesignDocumentView `json:"views,omitempty"`
	Filters  map[string]string             `json:"filters,omitempty"`
}

// Name returns design document name without the "_design/" prefix
func (dd DesignDocument) Name() string {
	return strings.TrimPrefix(dd.ID, "_design/")
}

// DesignDocumentView contains map/reduce functions.
type DesignDocumentView struct {
	Map    string `json:"map,omitempty"`
	Reduce string `json:"reduce,omitempty"`
}
