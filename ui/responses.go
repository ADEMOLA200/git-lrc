package ui

// SessionStatusResponse is returned by the connector manager session-status endpoint.
type SessionStatusResponse struct {
	Authenticated  bool   `json:"authenticated"`
	SessionExpired bool   `json:"session_expired"`
	MissingConfig  bool   `json:"missing_config"`
	DisplayName    string `json:"display_name,omitempty"`
	FirstName      string `json:"first_name,omitempty"`
	LastName       string `json:"last_name,omitempty"`
	AvatarURL      string `json:"avatar_url,omitempty"`
	UserEmail      string `json:"user_email,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	OrgID          string `json:"org_id,omitempty"`
	OrgName        string `json:"org_name,omitempty"`
	APIURL         string `json:"api_url"`
	Message        string `json:"message,omitempty"`
}
