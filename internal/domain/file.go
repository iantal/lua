package domain

import (
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
)

type File struct {
	gorm.Model   `json:"-"`
	ProjectID    uuid.UUID    `gorm:"type:uuid;primary_key;" json:"projectId"`
	CommitHash   string       `gorm:"primary_key" json:"commit,omitempty"`
	Name         string       `json:"name,omitempty"`
	Language     string       `json:"extension,omitempty"`
	Declarations string       `json:"declarations,omitempty"` // comma separated values
	Dependencies []Dependency `gorm:"many2many:file_dependencies;"`
}

type Dependency struct {
	gorm.Model `json:"-"`
	Name       string
	Classes    string // comma separated values
}

// NewFile creates a File
func NewFile(projectID uuid.UUID, commitHash, name, extension, declarations string, dependencies []Dependency) *File {
	return &File{
		ProjectID:    projectID,
		CommitHash:   commitHash,
		Name:         name,
		Language:     extension,
		Declarations: declarations,
		Dependencies: dependencies,
	}
}
