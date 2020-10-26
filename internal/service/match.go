package service

import (
	"strings"
	"sync"

	"github.com/iantal/lua/internal/domain"
	"github.com/iantal/lua/protos/lua"
)

func (a *Analyzer) matchUsedLibraries(data <-chan javaPipelineData) <-chan *domain.File {
	oc := make(chan *domain.File)

	go func() {
		wg := &sync.WaitGroup{}

		for file := range data {
			wg.Add(1)
			go a.matchFile(file, oc, wg)
		}

		wg.Wait()
		close(oc)
	}()

	return oc
}

func (a *Analyzer) matchFile(in javaPipelineData, output chan<- *domain.File, wg *sync.WaitGroup) {
	declarations := strings.Split(in.File.Declarations, ",")
	in.File.Dependencies = match(declarations, in.Libraries)
	output <- in.File
	wg.Done()
}

func match(declarations []string, libraries []*lua.Library) []domain.Dependency {
	usages := []domain.Dependency{}

	for _, l := range libraries {
		classes := []string{}
		for _, c := range l.Classes {
			for _, d := range declarations {
				if c == d {
					classes = append(classes, c)
				}
			}
		}
		if len(classes) > 0 {
			dep := domain.Dependency{
				Name:    l.Name,
				Classes: strings.Join(classes, ","),
			}
			usages = append(usages, dep)
		}
	}

	return usages
}
