package couchdb

// ViewResponse is response for querying design documents.
type ViewResponse struct {
	Offset    int   `json:"offset,omitempty"`
	Rows      []Row `json:"rows,omitempty"`
	TotalRows int   `json:"total_rows,omitempty"`
	UpdateSeq int   `json:"update_seq,omitempty"`
}

// Row is single row inside design document query response.
type Row struct {
	ID    string                 `json:"id"`
	Key   interface{}            `json:"key"`
	Value interface{}            `json:"value,omitempty"`
	Doc   map[string]interface{} `json:"doc,omitempty"`
}
