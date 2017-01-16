package couchdb

// Credentials has information about POST _session form parameters.
// http://docs.couchdb.org/en/latest/api/server/authn.html#cookie-authentication
type Credentials struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}
