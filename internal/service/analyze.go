package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	ldprotos "github.com/iantal/ld/protos/ld"
	"github.com/iantal/lua/internal/files"
	"github.com/iantal/lua/internal/repository"
	protos "github.com/iantal/lua/protos/lua"

	"github.com/jinzhu/gorm"
)

type Analyzer struct {
	log            hclog.Logger
	store          *files.Local
	filesDB        *repository.FileDB
	dependenciesDB *repository.DependenciesDB
	rmHost         string
	ld             ldprotos.UsedLanguagesClient
}

func NewAnalyzer(log hclog.Logger, stor *files.Local, db *gorm.DB, rmHost string, ld ldprotos.UsedLanguagesClient) *Analyzer {
	filesDB := repository.NewFileDB(log, db)
	depsDB := repository.NewDependenciesDB(log, db)

	return &Analyzer{log, stor, filesDB, depsDB, rmHost, ld}
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

	a.execPipeline(projectID, commit, libraries)

	return nil
}

func (a *Analyzer) execPipeline(projectID, commit string, libraries []*protos.Library) {

	r := a.matchUsedLibraries(
		a.extractDeclarations(
			a.getFilesByLanguage(projectID, commit, libraries)))

	a.log.Info("Persisting data", "project", projectID, "commit", commit)
	for res := range r {
		a.filesDB.AddFile(res)
	}

}
