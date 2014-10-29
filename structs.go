package couchdb

import (
	"fmt"
)

// http://docs.couchdb.org/en/latest/intro/api.html#server
type Server struct {
	Couchdb string
	Uuid    string
	Vendor  struct {
		Version string
		Name    string
	}
	Version string
}

// http://docs.couchdb.org/en/latest/api/database/common.html#get--db
type DatabaseInfo struct {
	DbName             string `json:"db_name"`
	DocCount           int    `json:"doc_count"`
	DocDelCount        int    `json:"doc_del_count"`
	UpdateSeq          int    `json:"update_seq"`
	PurgeSeq           int    `json:"purge_seq"`
	CompactRunning     bool   `json:"compact_running"`
	DiskSize           int    `json:"disk_size"`
	DataSize           int    `json:"data_size"`
	InstanceStartTime  string `json:"instance_start_time"`
	DiskFormatVersion  int    `json:"disk_format_version"`
	CommittedUpdateSeq int    `json:"committed_update_seq"`
}

type DatabaseResponse struct {
	Ok bool
}

type Error struct {
	Method     string
	Url        string
	StatusCode int
	Type       string `json:"error"`
	Reason     string
}

func (e *Error) Error() string {
	return fmt.Sprintf("CouchDB - %s %s, Status Code: %d, Error: %s, Reason: %s", e.Method, e.Url, e.StatusCode, e.Type, e.Reason)
}

type Document struct {
	Id          string                `json:"_id,omitempty"`
	Rev         string                `json:"_rev,omitempty"`
	Attachments map[string]Attachment `json:"_attachments,omitempty"`
}

// http://docs.couchdb.org/en/latest/api/document/common.html#creating-multiple-attachments
type Attachment struct {
	Follows     bool   `json:"follows"`
	ContentType string `json:"content_type"`
	Length      int64  `json:"length"`
}

type CouchDoc interface {
	GetDocument() *Document
}

type DocumentResponse struct {
	Ok  bool
	Id  string
	Rev string
}

// http://docs.couchdb.org/en/latest/api/server/common.html#active-tasks
type Task struct {
	ChangesDone  int `json:"changes_done"`
	Database     string
	Pid          string
	Progress     int
	StartedOn    int `json:"started_on"`
	Status       string
	Task         string
	TotalChanges int `json:"total_changes"`
	Type         string
	UpdatedOn    string `json:"updated_on"`
}

type QueryParameters struct {
	Conflicts       bool   `url:"conflicts,omitempty"`
	Descending      bool   `url:"descending,omitempty"`
	EndKey          string `url:"endkey,comma,omitempty"`
	EndKeyDocId     string `url:"end_key_doc_id,omitempty"`
	Group           bool   `url:"group,omitempty"`
	GroupLevel      int    `url:"group_level,omitempty"`
	IncludeDocs     bool   `url:"include_docs,omitempty"`
	Attachments     bool   `url:"attachments,omitempty"`
	AttEncodingInfo bool   `url:"att_encoding_info,omitempty"`
	InclusiveEnd    bool   `url:"inclusive_end,omitempty"`
	Key             string `url:"key,omitempty"`
	Limit           int    `url:"limit,omitempty"`
	Reduce          bool   `url:"reduce,omitempty"`
	Skip            int    `url:"skip,omitempty"`
	Stale           string `url:"stale,omitempty"`
	StartKey        string `url:"startkey,comma,omitempty"`
	StartKeyDocId   string `url:"startkey_docid,omitempty"`
	UpdateSeq       bool   `url:"update_seq,omitempty"`
}

type ViewResponse struct {
	Offset    int   `json:"offset,omitempty"`
	Rows      []Row `json:"rows,omitempty"`
	TotalRows int   `json:"total_rows,omitempty"`
	UpdateSeq int   `json:"update_seq,omitempty"`
}

type Row struct {
	Id    string                 `json:"id"`
	Key   interface{}            `json:"key"`
	Value interface{}            `json:"value,omitempty"`
	Doc   map[string]interface{} `json:"doc,omitempty"`
}

// http://docs.couchdb.org/en/latest/api/database/bulk-api.html#post--db-_bulk_docs
type BulkDoc struct {
	AllOrNothing bool          `json:"all_or_nothing,omitempty"`
	Docs         []interface{} `json:"docs"`
	NewEdits     bool          `json:"new_edits,omitempty"`
}

// http://docs.couchdb.org/en/latest/api/server/authn.html#cookie-authentication
type Credentials struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type PostSessionResponse struct {
	Ok    bool
	Name  string
	Roles []string
}

type User struct {
	Document
	DerivedKey     string   `json:"derived_key,omitempty"`
	Name           string   `json:"name,omitempty"`
	Roles          []string `json:"roles"`
	Password       string   `json:"password,omitempty"`     // plain text password when creating the user
	PasswordSha    string   `json:"password_sha,omitempty"` // hashed password when requesting user information
	PasswordScheme string   `json:"password_scheme,omitempty"`
	Salt           string   `json:"salt,omitempty"`
	Type           string   `json:"type,omitempty"`
	Iterations     int      `json:"iterations,omitempty"`
}

func (user User) GetDocument() *Document {
	return &user.Document
}

func NewUser(name, password string, roles []string) User {
	user := User{
		Document: Document{
			Id: "org.couchdb.user:" + name,
		},
		DerivedKey:     "",
		Name:           name,
		Roles:          roles,
		Password:       password,
		PasswordSha:    "",
		PasswordScheme: "",
		Salt:           "",
		Type:           "user",
	}
	return user
}

type GetSessionResponse struct {
	Info struct {
		Authenticated          string   `json:"authenticated"`
		AuthenticationDb       string   `json:"authentication_db"`
		AuthenticationHandlers []string `json:"authentication_handlers"`
	} `json:"info"`
	Ok          bool `json:"ok"`
	UserContext struct {
		Db    string   `json:"db"`
		Name  string   `json:"name"`
		Roles []string `json:"roles"`
	} `json:"userCtx"`
}
