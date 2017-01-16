package couchdb

import "fmt"

// Error describes CouchDB error.
type Error struct {
	Method     string
	URL        string
	StatusCode int
	Type       string `json:"error"`
	Reason     string
}

func (e *Error) Error() string {
	return fmt.Sprintf(
		"CouchDB - %s %s, Status Code: %d, Error: %s, Reason: %s",
		e.Method,
		e.URL,
		e.StatusCode,
		e.Type,
		e.Reason,
	)
}
