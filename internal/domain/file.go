package domain

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type File struct {
	gorm.Model   `json:"-"`
	ProjectID    uuid.UUID `gorm:"type:uuid;primary_key;" json:"projectId"`
	CommitHash   string    `gorm:"primary_key" json:"commit,omitempty"`
	Name         string    `json:"name,omitempty"`
	Extension    string    `json:"extension,omitempty"`
	Declarations string    `json:"declarations,omitempty"` // comma separated values
	Usages       string    `json:"usages,omitempty"`       // comma separated values
}

// NewFile creates a File
func NewFile(projectID uuid.UUID, commitHash, name, extension, declarations, usages string) *File {
	return &File{
		ProjectID:    projectID,
		CommitHash:   commitHash,
		Name:         name,
		Extension:    extension,
		Declarations: declarations,
		Usages:       usages,
	}
}
