package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	"github.com/iantal/lua/internal/files"
	"github.com/iantal/lua/internal/repository"
	protos "github.com/iantal/lua/protos/lua"
	"github.com/jinzhu/gorm"
)

type Analyzer struct {
	log    hclog.Logger
	store  *files.Local
	db     *repository.FileDB
	rmHost string
	ldHost string
}

func NewAnalyzer(log hclog.Logger, stor *files.Local, db *gorm.DB, rmHost, ldHost string) *Analyzer {
	dbs := repository.NewFileDB(log, db)
	return &Analyzer{log, stor, dbs, rmHost, ldHost}
}

func (a *Analyzer) Analyze(projectID, commit string, libraries []*protos.Library) error {
	projectPath := filepath.Join(a.store.FullPath(projectID), commit, "bundle")

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		err := a.downloadRepository(projectID, commit)
		if err != nil {
			return fmt.Errorf("Could not download bundled repository for project %s, commit %s, error %s", projectID, commit, err)
		}
	}

	bp := commit + ".bundle"
	srcPath := a.store.FullPath(filepath.Join(projectID, commit, "bundle", bp))
	destPath := a.store.FullPath(filepath.Join(projectID, commit, "unbundle"))

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		err := a.store.Unbundle(srcPath, destPath)
		if err != nil {
			return fmt.Errorf("Could not unbundle repository for project %s, commit %s, error %s", projectID, commit, err)
		}
	}

	return nil
}
