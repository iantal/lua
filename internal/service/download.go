package service

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

func (a *Analyzer) downloadRepository(projectID, commit string) error {
	resp, err := http.DefaultClient.Get("http://" + a.rmHost + "/api/v1/projects/" + projectID + "/" + commit + "/download")
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected error code 200 got %d", resp.StatusCode)
	}

	a.log.WithFields(logrus.Fields{
		"projectID": projectID,
		"commit":    commit,
		"file":      resp.Header.Get("Content-Disposition"),
	}).Info("Content-Dispozition")

	a.save(projectID, commit, resp.Body)
	resp.Body.Close()

	return nil
}

func (a *Analyzer) save(projectID, commit string, r io.ReadCloser) {
	a.log.WithFields(logrus.Fields{
		"projectID": projectID,
		"commit":    commit,
	}).Info("Save project to storage")

	bp := commit + ".bundle"
	fp := filepath.Join(projectID, commit, "bundle", bp)
	err := a.store.Save(fp, r)

	if err != nil {
		a.log.WithFields(logrus.Fields{
			"projectID": projectID,
			"commit":    commit,
			"error":     err,
		}).Error("Unable to save file")
	}
}
