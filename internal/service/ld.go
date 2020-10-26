package service

import (
	"context"
	"sync"

	"github.com/google/uuid"
	ldprotos "github.com/iantal/ld/protos/ld"
	"github.com/iantal/lua/internal/domain"
	protos "github.com/iantal/lua/protos/lua"
)

type javaPipelineData struct {
	File      *domain.File
	Libraries []*protos.Library
}

func (a *Analyzer) getFilesByLanguage(projectID, commit string, libraries []*protos.Library) <-chan javaPipelineData {
	a.log.Info("Requesting ld for languages and files", "project", projectID, "commit", commit)
	c := make(chan javaPipelineData)

	go func() {

		wg := &sync.WaitGroup{}

		r := &ldprotos.BreakdownRequest{
			ProjectID:  projectID,
			CommitHash: commit,
		}
		resp, err := a.ld.Breakdown(context.Background(), r)
		if err != nil {
			a.log.Error("Could not get languages and files for", "projectID", projectID, "commit", commit)
		} else {
			for _, language := range resp.Breakdown {
				wg.Add(1)
				go filterLanguage(projectID, commit, language.Name, libraries, language.Files, c, wg)
			}
		}
		wg.Wait()
		close(c)
	}()

	return c

}

func filterLanguage(projectID, commit, language string, libraries []*protos.Library, files []string, output chan<- javaPipelineData, wg *sync.WaitGroup) {
	if language == "Java" {
		for _, file := range files {
			f := &domain.File{
				ProjectID:  uuid.MustParse(projectID),
				CommitHash: commit,
				Name:       file,
				Language:  language,
			}
			output <- javaPipelineData{f, libraries}
		}
	}
	wg.Done()
}
