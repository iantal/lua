package repository

import (
	"github.com/iantal/lua/internal/domain"
	"github.com/iantal/lua/internal/util"
	"github.com/jinzhu/gorm"
)

// DependenciesDB defines the CRUD operations for storing projects in the db
type DependenciesDB struct {
	log *util.StandardLogger
	db  *gorm.DB
}

// NewDependenciesDB returns a FileDB object for handling CRUD operations
func NewDependenciesDB(log *util.StandardLogger, db *gorm.DB) *DependenciesDB {
	db.AutoMigrate(&domain.File{})
	return &DependenciesDB{
		log: log,
		db:  db,
	}
}

// AddDependency adds a library to the db
func (l *DependenciesDB) AddDependency(file *domain.File) {
	l.db.Create(&file)
	return
}
