package couchdb

// User is special CouchDB document format.
// http://docs.couchdb.org/en/latest/intro/security.html#users-documents
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

// NewUser returns new user instance.
func NewUser(name, password string, roles []string) User {
	user := User{
		Document: Document{
			ID: "org.couchdb.user:" + name,
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
