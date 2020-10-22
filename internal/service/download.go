package service

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
)

func (a *Analyzer) downloadRepository(projectID, commit string) error {
	resp, err := http.DefaultClient.Get("http://" + a.rmHost + "/api/v1/projects/" + projectID + "/" + commit + "/download")
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Expected error code 200 got %d", resp.StatusCode)
	}

	a.log.Info("Content-Dispozition", "file", resp.Header.Get("Content-Disposition"))

	a.save(projectID, commit, resp.Body)
	resp.Body.Close()

	return nil
}

func (a *Analyzer) save(projectID, commit string, r io.ReadCloser) error {
	a.log.Info("Save project - storage", "projectID", projectID)

	bp := commit + ".bundle"
	fp := filepath.Join(projectID, commit, "bundle", bp)
	err := a.store.Save(fp, r)

	if err != nil {
		a.log.Error("Unable to save file", "error", err)
		return fmt.Errorf("Unable to save file %s", err)
	}

	return nil
}
