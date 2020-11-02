package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-hclog"
	ldprotos "github.com/iantal/ld/protos/ld"
	mcdprotos "github.com/iantal/mcd/protos/mcd"

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
	mcd            mcdprotos.DownloaderClient
}

func NewAnalyzer(log hclog.Logger, stor *files.Local, db *gorm.DB, rmHost string, ld ldprotos.UsedLanguagesClient, mcd mcdprotos.DownloaderClient) *Analyzer {
	filesDB := repository.NewFileDB(log, db)
	depsDB := repository.NewDependenciesDB(log, db)

	return &Analyzer{log, stor, filesDB, depsDB, rmHost, ld, mcd}
}

func (a *Analyzer) Analyze(projectID, commit string, libraries []*protos.LuaLibrary) error {
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

func (a *Analyzer) execPipeline(projectID, commit string, libraries []*protos.LuaLibrary) {

	r := a.matchUsedLibraries(
		a.extractDeclarations(
			a.getFilesByLanguage(projectID, commit, libraries)))

	a.log.Info("Persisting data", "project", projectID, "commit", commit)
	for res := range r {
		a.filesDB.AddFile(res)

		if len(res.Dependencies) > 0 {
			fmt.Printf("File %s", res.Name)
			// TODO: send to VA
		}
	}

}
