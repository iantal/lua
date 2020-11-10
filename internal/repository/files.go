package repository

import (
	"github.com/google/uuid"
	"github.com/iantal/lua/internal/domain"
	"github.com/iantal/lua/internal/util"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// FileDB defines the CRUD operations for storing projects in the db
type FileDB struct {
	log *util.StandardLogger
	db  *gorm.DB
}

// NewFileDB returns a FileDB object for handling CRUD operations
func NewFileDB(log *util.StandardLogger, db *gorm.DB) *FileDB {
	db.AutoMigrate(&domain.File{})
	return &FileDB{
		log: log,
		db:  db,
	}
}

// AddFile adds a library to the db
func (l *FileDB) AddFile(file *domain.File) {
	l.db.Create(&file)
}

// GetFilesByIDAndCommit returns all files for the given id and commit
func (l *FileDB) GetFilesByIDAndCommit(id, commit string) []*domain.File {
	var files []*domain.File
	uid, err := uuid.Parse(id)
	if err != nil {
		l.log.WithFields(logrus.Fields{
			"projectID": id,
			"commit": commit,
		}).Error("No libraries were found")
		return nil
	}
	l.db.Where("project_id = ? AND commit_hash = ?", uid, commit).Find(&files)
	return files
}
