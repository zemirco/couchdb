package couchdb

// BulkDoc describes POST /db/_bulk_docs request object.
// http://docs.couchdb.org/en/latest/api/database/bulk-api.html#post--db-_bulk_docs
type BulkDoc struct {
	AllOrNothing bool       `json:"all_or_nothing,omitempty"`
	NewEdits     bool       `json:"new_edits,omitempty"`
	Docs         []CouchDoc `json:"docs"`
}
