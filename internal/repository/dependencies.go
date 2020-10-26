package repository

import (
	"github.com/hashicorp/go-hclog"
	"github.com/iantal/lua/internal/domain"
	"github.com/jinzhu/gorm"
)

// DependenciesDB defines the CRUD operations for storing projects in the db
type DependenciesDB struct {
	log hclog.Logger
	db  *gorm.DB
}

// NewDependenciesDB returns a FileDB object for handling CRUD operations
func NewDependenciesDB(log hclog.Logger, db *gorm.DB) *DependenciesDB {
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
