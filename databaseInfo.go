package couchdb

// DatabaseInfo has info about the specified database.
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
