package couchdb

// PostSessionResponse is response from posting to session api.
type PostSessionResponse struct {
	Ok    bool
	Name  string
	Roles []string
}
