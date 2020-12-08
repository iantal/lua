package service

import (
	"context"
	"os"
	"path/filepath"

	ldprotos "github.com/iantal/ld/protos/ld"
	vaprotos "github.com/iantal/va/protos/va"
	"github.com/sirupsen/logrus"

	"github.com/iantal/lua/internal/files"
	"github.com/iantal/lua/internal/repository"
	"github.com/iantal/lua/internal/util"
	protos "github.com/iantal/lua/protos/lua"

	"github.com/jinzhu/gorm"
)

type Analyzer struct {
	log            *util.StandardLogger
	store          *files.Local
	filesDB        *repository.FileDB
	dependenciesDB *repository.DependenciesDB
	rmHost         string
	ld             ldprotos.UsedLanguagesClient
	va             vaprotos.VulnerabilityAnalyzerClient
}

func NewAnalyzer(log *util.StandardLogger, stor *files.Local, db *gorm.DB, rmHost string, ld ldprotos.UsedLanguagesClient, va vaprotos.VulnerabilityAnalyzerClient) *Analyzer {
	filesDB := repository.NewFileDB(log, db)
	depsDB := repository.NewDependenciesDB(log, db)

	return &Analyzer{log, stor, filesDB, depsDB, rmHost, ld, va}
}

func (a *Analyzer) Analyze(projectID, commit string, libraries []*protos.LuaLibrary) {
	projectPath := filepath.Join(a.store.FullPath(projectID), commit, "bundle")

	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		a.log.WithFields(logrus.Fields{
			"projectID": projectID,
			"commit":    commit,
		}).Info("Downloading repository from rm")
		err := a.downloadRepository(projectID, commit)
		if err != nil {
			a.log.WithFields(logrus.Fields{
				"projectID": projectID,
				"commit":    commit,
				"error":     err,
			}).Error("Could not download bundled repository")
			return
		}
	}

	bp := commit + ".bundle"
	srcPath := a.store.FullPath(filepath.Join(projectID, commit, "bundle", bp))
	destPath := a.store.FullPath(filepath.Join(projectID, commit, "unbundle"))

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		a.log.WithFields(logrus.Fields{
			"projectID": projectID,
			"commit":    commit,
		}).Info("Unbundle downloaded repository")

		err := a.store.Unbundle(srcPath, destPath)
		if err != nil {
			a.log.WithFields(logrus.Fields{
				"projectID": projectID,
				"commit":    commit,
				"error":     err,
			}).Error("Could not unbundle repository")
			return
		}
	}

	a.execPipeline(projectID, commit, libraries)
}

func (a *Analyzer) execPipeline(projectID, commit string, libraries []*protos.LuaLibrary) {
	a.log.WithFields(logrus.Fields{
		"projectID": projectID,
		"commit":    commit,
	}).Info("Executing pipeline")

	fbl := a.getFilesByLanguage(projectID, commit, libraries)

	r := a.matchUsedLibraries(a.extractDeclarations(fbl))

	a.log.WithFields(logrus.Fields{
		"projectID": projectID,
		"commit":    commit,
	}).Info("Persisting data")
	for res := range r {
		a.filesDB.AddFile(res)

		if len(res.Dependencies) > 0 {
			a.log.WithFields(logrus.Fields{
				"projectID":    projectID,
				"commit":       commit,
				"projectName":  res.Name,
				"dependencies": res.Dependencies,
			}).Info("File has dependencies")

			dependencyNames := []string{}
			for _, d := range res.Dependencies {
				dependencyNames = append(dependencyNames, d.Name)
			}

			vaReq := &vaprotos.VulnerabilityAnalyzeRequest{
				ProjectID:  projectID,
				CommitHash: commit,
				Libraries:  dependencyNames,
			}

			_, err := a.va.Analyze(context.Background(), vaReq)
			if err != nil {
				a.log.WithField("error", err).Error("Failed to send request to VA")
			}
		}
	}

}
