package couchdb

// Task describes currently running task.
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
	UpdatedOn    int `json:"updated_on"`
}
