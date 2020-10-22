package server

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/iantal/lua/internal/files"
	"github.com/iantal/lua/internal/service"
	protos "github.com/iantal/lua/protos/lua"
	"github.com/jinzhu/gorm"
)

type LibraryUsageAnalyser struct {
	log hclog.Logger
	as  *service.Analyzer
}

func NewLibraryUsageAnalyser(l hclog.Logger, stor *files.Local, db *gorm.DB, rmHost, ldHost string) *LibraryUsageAnalyser {
	return &LibraryUsageAnalyser{l, service.NewAnalyzer(l, stor, db, rmHost, ldHost)}
}

func (l *LibraryUsageAnalyser) Analyze(ctx context.Context, rr *protos.AnalyzeRequest) (*protos.AnalyzeResponse, error) {
	projectID := rr.GetProjectID()
	commit := rr.GetCommitHash()
	l.log.Info("Handle request for project", "projectID", projectID, "commit", commit)
	err := l.as.Analyze(projectID, commit, rr.GetLibraries())
	if err != nil {
		l.log.Error("Error %s", err)
	}
	return &protos.AnalyzeResponse{}, nil
}
