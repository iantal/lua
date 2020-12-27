package server

import (
	"testing"

	"github.com/google/uuid"
	"github.com/iantal/lua/internal/domain"
)

func equal(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestFilterAffectedFiles(t *testing.T) {
	projectID := uuid.New()
	commit := "123"

	files := []*domain.File{
		domain.NewFile(projectID, commit, "file1", ".java", "a,b,c", []domain.Dependency{
			{
				Name: "lib1",
			},
			{
				Name: "lib2",
			},
			{
				Name: "lib3",
			},
			{
				Name: "lib4",
			},
		}),
		domain.NewFile(projectID, commit, "file2", ".java", "a,b,c", []domain.Dependency{
			{
				Name: "lib1",
			},
			{
				Name: "lib4",
			},
			{
				Name: "lib7",
			},
		}),
		domain.NewFile(projectID, commit, "file3", ".java", "a,b,c", []domain.Dependency{
			{
				Name: "lib5",
			},
			{
				Name: "lib6",
			},
		}),
	}

	affectedProjects := filterAffectedFiles("lib1", files)
	expected := []string{"file1", "file2"}
	if !equal(affectedProjects, expected) {
		t.Errorf("expected %s, actual %s", expected, affectedProjects)
	}
}
