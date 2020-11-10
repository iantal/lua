package server

import (
	"context"

	ldprotos "github.com/iantal/ld/protos/ld"
	"github.com/iantal/lua/internal/files"
	"github.com/iantal/lua/internal/service"
	"github.com/iantal/lua/internal/util"
	protos "github.com/iantal/lua/protos/lua"
	vaprotos "github.com/iantal/va/protos/va"
	"github.com/sirupsen/logrus"

	"github.com/jinzhu/gorm"
)

type LibraryUsageAnalyser struct {
	log *util.StandardLogger
	as  *service.Analyzer
}

func NewLibraryUsageAnalyser(l *util.StandardLogger, stor *files.Local, db *gorm.DB, rmHost string, ld ldprotos.UsedLanguagesClient, va vaprotos.VulnerabilityAnalyzerClient) *LibraryUsageAnalyser {
	return &LibraryUsageAnalyser{l, service.NewAnalyzer(l, stor, db, rmHost, ld, va)}
}

func (l *LibraryUsageAnalyser) Analyze(ctx context.Context, rr *protos.AnalyzeRequest) (*protos.AnalyzeResponse, error) {
	projectID := rr.GetProjectID()
	commit := rr.GetCommitHash()
	l.log.WithFields(logrus.Fields{
		"projectID": projectID,
		"commit": commit,
	}).Info("Handle request for project")

	go func() {
		l.as.Analyze(projectID, commit, rr.GetLibraries())
	}()
	return &protos.AnalyzeResponse{}, nil
}
