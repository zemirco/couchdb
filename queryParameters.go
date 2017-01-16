package couchdb

// QueryParameters is struct to define url query parameters for design documents.
// http://docs.couchdb.org/en/latest/api/ddoc/views.html#db-design-design-doc-view-view-name
type QueryParameters struct {
	Conflicts       *bool   `url:"conflicts,omitempty"`
	Descending      *bool   `url:"descending,omitempty"`
	Group           *bool   `url:"group,omitempty"`
	IncludeDocs     *bool   `url:"include_docs,omitempty"`
	Attachments     *bool   `url:"attachments,omitempty"`
	AttEncodingInfo *bool   `url:"att_encoding_info,omitempty"`
	InclusiveEnd    *bool   `url:"inclusive_end,omitempty"`
	Reduce          *bool   `url:"reduce,omitempty"`
	UpdateSeq       *bool   `url:"update_seq,omitempty"`
	GroupLevel      *int    `url:"group_level,omitempty"`
	Limit           *int    `url:"limit,omitempty"`
	Skip            *int    `url:"skip,omitempty"`
	Key             *string `url:"key,omitempty"`
	EndKey          *string `url:"endkey,comma,omitempty"`
	EndKeyDocID     *string `url:"end_key_doc_id,omitempty"`
	Stale           *string `url:"stale,omitempty"`
	StartKey        *string `url:"startkey,comma,omitempty"`
	StartKeyDocID   *string `url:"startkey_docid,omitempty"`
}
