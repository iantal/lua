package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	ldprotos "github.com/iantal/ld/protos/ld"
	vaprotos "github.com/iantal/va/protos/va"
	"github.com/sirupsen/logrus"

	"github.com/iantal/lua/internal/domain"
	"github.com/iantal/lua/internal/files"
	"github.com/iantal/lua/internal/repository"
	"github.com/iantal/lua/internal/util"
	protos "github.com/iantal/lua/protos/lua"

	"github.com/jinzhu/gorm"
)

type void struct{}

var member void

type Analyzer struct {
	log     *util.StandardLogger
	store   *files.Local
	filesDB *repository.FileDB
	rmHost  string
	ld      ldprotos.UsedLanguagesClient
	va      vaprotos.VulnerabilityAnalyzerClient
}

func NewAnalyzer(log *util.StandardLogger, stor *files.Local, db *gorm.DB, rmHost string, ld ldprotos.UsedLanguagesClient, va vaprotos.VulnerabilityAnalyzerClient) *Analyzer {
	filesDB := repository.NewFileDB(log, db)
	return &Analyzer{log, stor, filesDB, rmHost, ld, va}
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

	dependencyNames := []string{}
	dependencySet := make(map[string]void)

	for res := range r {
		a.filesDB.AddFile(res)
		for _, d := range res.Dependencies {
			if _, exists := dependencySet[d.Name]; !exists {
				dependencySet[d.Name] = member
				dependencyNames = append(dependencyNames, d.Name)
			}
		}
	}

	a.log.WithFields(logrus.Fields{
		"projectId": projectID,
		"commit":    commit,
		"totalLibs": len(dependencyNames),
	}).Info("Number of dependencies that will be analyzed")

	scan := &domain.Scan{
		ProjectID:  projectID,
		CommitHash: commit,
		Libraries:  dependencyNames,
	}

	a.startScan(scan)
}

func (a *Analyzer) startScan(scan *domain.Scan) {
	jsonData, err := json.Marshal(scan)
	if err != nil {
		a.log.WithFields(logrus.Fields{
			"projectID": scan.ProjectID,
			"commit":    scan.CommitHash,
			"err":       err,
		}).Error("Error marshaling vulnerable libraries")
		return
	}

	resp, err := http.Post("http://odc-service:8009/api/v1/scans", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		a.log.WithFields(logrus.Fields{
			"projectID": scan.ProjectID,
			"commit":    scan.CommitHash,
			"err":       err,
		}).Error("Error sending request to odc")
		return
	}

	defer resp.Body.Close()
}
