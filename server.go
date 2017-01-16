package couchdb

// Server gives access to the welcome string and version information.
// http://docs.couchdb.org/en/latest/intro/api.html#server
type Server struct {
	Couchdb string
	UUID    string
	Vendor  struct {
		Version string
		Name    string
	}
	Version string
}
