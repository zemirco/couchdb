package couchdb

// GetSessionResponse returns complete information about authenticated user.
// http://docs.couchdb.org/en/latest/api/server/authn.html#get--_session
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
