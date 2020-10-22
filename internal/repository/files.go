package repository

import (
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"github.com/iantal/lua/internal/domain"
	"github.com/jinzhu/gorm"
)

// FileDB defines the CRUD operations for storing projects in the db
type FileDB struct {
	log hclog.Logger
	db  *gorm.DB
}

// NewFileDB returns a FileDB object for handling CRUD operations
func NewFileDB(log hclog.Logger, db *gorm.DB) *FileDB {
	db.AutoMigrate(&domain.File{})
	return &FileDB{
		log: log,
		db:  db,
	}
}

// AddFile adds a library to the db
func (l *FileDB) AddFile(file *domain.File) {
	l.db.Create(&file)
	return
}

// GetFilesByIDAndCommit returns all files for the given id and commit
func (l *FileDB) GetFilesByIDAndCommit(id, commit string) []*domain.File {
	var files []*domain.File
	uid, err := uuid.Parse(id)
	if err != nil {
		l.log.Error("No libraries with projectId {} were found")
		return nil
	}
	l.db.Where("project_id = ? AND commit_hash = ?", uid, commit).Find(&files)
	return files
}
