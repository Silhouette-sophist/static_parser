package service

import (
	"path/filepath"
	"testing"
)

func TestParseRepo(t *testing.T) {
	curRepoPath, err := filepath.Abs("./..")
	if err != nil {
		t.Fatal(err)
	}
	modules, err := ParseRepo(curRepoPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, module := range modules {
		t.Log(module)
	}
}
