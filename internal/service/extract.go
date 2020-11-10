package service

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

func (a *Analyzer) extractDeclarations(data []javaPipelineData) <-chan javaPipelineData {
	oc := make(chan javaPipelineData)

	go func() {
		wg := &sync.WaitGroup{}

		for _, file := range data {
			wg.Add(1)
			go a.parseFile(file, oc, wg)
		}

		wg.Wait()
		close(oc)
	}()

	return oc
}

func (a *Analyzer) parseFile(in javaPipelineData, output chan<- javaPipelineData, wg *sync.WaitGroup) {
	file := in.File

	filePath := filepath.Join(file.ProjectID.String(), file.CommitHash, "unbundle", file.CommitHash, in.File.Name)
	fullPath := a.store.FullPath(filePath)
	in.File.Declarations = strings.Join(a.extractJavaImports(fullPath), ",")

	output <- in
	wg.Done()
}

func (a *Analyzer) extractJavaImports(fpath string) []string {
	var result []string
	r := regexp.MustCompile("^(\\s*)?import\\s+(static\\s+)?(?P<import>.*);")

	file, err := os.Open(fpath)
	if err != nil {
		a.log.Error("Cannot open file", "file", fpath, "error", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		e := extractImport(line, "import", r)
		if e != "" {
			result = append(result, e)
		}
	}

	if err := scanner.Err(); err != nil {
		a.log.WithFields(logrus.Fields{
			"file":  fpath,
			"error": err,
		}).Error("Cannot read file")
		return []string{}
	}

	return result
}

func extractImport(line, regexName string, expression *regexp.Regexp) string {

	result := make(map[string]string)
	match := expression.FindStringSubmatch(line)
	if match == nil {
		return ""
	}
	for i, name := range expression.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}
	return result[regexName]
}
