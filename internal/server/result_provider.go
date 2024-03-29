package server

import (
	"context"

	"github.com/iantal/lua/internal/domain"
	"github.com/iantal/lua/internal/repository"
	"github.com/iantal/lua/internal/util"
	protos "github.com/iantal/lua/protos/luaresult"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

type ResultProvider struct {
	log     *util.StandardLogger
	filesDB *repository.FileDB
}

func NewResultProvider(log *util.StandardLogger, db *gorm.DB) *ResultProvider {
	filesDB := repository.NewFileDB(log, db)
	return &ResultProvider{log, filesDB}
}

func (r *ResultProvider) ProvideVulnerableLibrariesData(ctx context.Context, rr *protos.ResultRequest) (*protos.ResultResponse, error) {
	projectID := rr.GetProjectID()
	commit := rr.GetCommitHash()
	libs := rr.GetLibraries()

	r.log.WithFields(logrus.Fields{
		"projectID": projectID,
		"commit":    commit,
		"libraries": libs,
	}).Info("Providing results for project")

	vulnerableLibs := []*protos.VulnerableLibrary{}

	files := r.filesDB.GetFilesByIDAndCommit(projectID, commit)
	for _, lib := range libs {
		vulnerableLibs = append(vulnerableLibs, r.mapLibrary(projectID, commit, lib, files))
	}

	response := &protos.ResultResponse{VulnerableLibraries: vulnerableLibs}
	return response, nil
}

func (r *ResultProvider) mapLibrary(projectID, commit, lib string, files []*domain.File) *protos.VulnerableLibrary {

	affectedFiles := r.filesDB.GetAffectedProjectFiles(projectID, commit, lib)
	vl := &protos.VulnerableLibrary{
		ProjectID:            projectID,
		CommitHash:           commit,
		Name:                 lib,
		AffectedProjectFiles: affectedFiles,
	}

	return vl
}
