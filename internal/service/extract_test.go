package service

import (
	"bufio"
	"os"
	"testing"

	"github.com/hashicorp/go-hclog"
)

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

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

func TestShouldReturnExtractedImports(t *testing.T) {
	a := &Analyzer{}
	in := "../../testdata/java_imports.in"
	out := "../../testdata/java_imports.out"

	expected, err := readLines(out)
	if err != nil {
		t.Errorf("Could not read from %s", out)
	}

	actual := a.extractJavaImports(in)

	if !equal(actual, expected) {
		t.Errorf("expected %s, actual %s", expected, actual)
	}
}

func TestInexistingFile(t *testing.T) {
	a := &Analyzer{
		log: hclog.Default(),
	}

	in := "a.in"
	actual := a.extractJavaImports(in)

	if !equal(actual, nil) {
		t.Errorf("expected %s, actual %s", "nil", actual)
	}
}

func TestReadUtfCharacters(t *testing.T) {
	a := &Analyzer{}
	in := "../../testdata/java_imports_2.in"

	actual := a.extractJavaImports(in)

	if !equal(actual, []string{}) {
		t.Errorf("expected %s, actual %s", []string{}, actual)
	}
}
