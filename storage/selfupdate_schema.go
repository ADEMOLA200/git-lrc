package storage

type PendingUpdateState struct {
	Version          string `json:"version"`
	StagedBinaryPath string `json:"staged_binary_path"`
	DownloadedAt     string `json:"downloaded_at"`
	IntegrityHash    string `json:"integrity_hash,omitempty"`
}

type UpdateLockMetadata struct {
	PID       int    `json:"pid"`
	UID       string `json:"uid,omitempty"`
	Username  string `json:"username,omitempty"`
	Command   string `json:"command"`
	Version   string `json:"version"`
	StartedAt string `json:"started_at"`
}
