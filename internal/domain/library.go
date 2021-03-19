package domain

type Scan struct {
	ProjectID  string   `json:"projectId"`
	CommitHash string   `json:"commitHash"`
	Libraries  []string `json:"libraries"`
}
