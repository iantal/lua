package server

import (
	"context"

	"github.com/hashicorp/go-hclog"
	ldprotos "github.com/iantal/ld/protos/ld"
	"github.com/iantal/lua/internal/files"
	"github.com/iantal/lua/internal/service"
	protos "github.com/iantal/lua/protos/lua"

	"github.com/jinzhu/gorm"
)

type LibraryUsageAnalyser struct {
	log hclog.Logger
	as  *service.Analyzer
}

func NewLibraryUsageAnalyser(l hclog.Logger, stor *files.Local, db *gorm.DB, rmHost string, ld ldprotos.UsedLanguagesClient) *LibraryUsageAnalyser {
	return &LibraryUsageAnalyser{l, service.NewAnalyzer(l, stor, db, rmHost, ld)}
}

func (l *LibraryUsageAnalyser) Analyze(ctx context.Context, rr *protos.AnalyzeRequest) (*protos.AnalyzeResponse, error) {
	projectID := rr.GetProjectID()
	commit := rr.GetCommitHash()
	l.log.Info("Handle request for project", "projectID", projectID, "commit", commit)

	go func() {
		err := l.as.Analyze(projectID, commit, rr.GetLibraries())
		if err != nil {
			l.log.Error("Error %s", err)
		}
	}()
	return &protos.AnalyzeResponse{}, nil
}
