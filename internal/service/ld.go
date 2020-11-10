package service

import (
	"context"

	"github.com/google/uuid"
	ldprotos "github.com/iantal/ld/protos/ld"
	"github.com/iantal/lua/internal/domain"
	protos "github.com/iantal/lua/protos/lua"
	"github.com/sirupsen/logrus"
)

type javaPipelineData struct {
	File      *domain.File
	Libraries []*protos.LuaLibrary
}

func (a *Analyzer) getFilesByLanguage(projectID, commit string, libraries []*protos.LuaLibrary) []javaPipelineData {
	a.log.WithFields(logrus.Fields{
		"projectID": projectID,
		"commit":    commit,
	}).Info("Requesting ld for languages and files")
	result := []javaPipelineData{}
	r := &ldprotos.BreakdownRequest{
		ProjectID:  projectID,
		CommitHash: commit,
	}
	resp, err := a.ld.Breakdown(context.Background(), r)
	if err != nil {
		a.log.WithFields(logrus.Fields{
			"projectID": projectID,
			"commit":    commit,
			"error":     err,
		}).Error("Could not get languages and files")
	} else {
		for _, language := range resp.Breakdown {
			if language.Name == "Java" {
				result = filterLanguage(projectID, commit, language.Name, libraries, language.Files)
			}
		}
	}
	return result

}

func filterLanguage(projectID, commit, language string, libraries []*protos.LuaLibrary, files []string) []javaPipelineData {
	result := []javaPipelineData{}

	for _, file := range files {
		f := &domain.File{
			ProjectID:  uuid.MustParse(projectID),
			CommitHash: commit,
			Name:       file,
			Language:   language,
		}
		result = append(result, javaPipelineData{f, libraries})
	}

	return result
}
