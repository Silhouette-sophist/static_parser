package service

import (
	"context"
	"testing"
)

func TestLoadPackages(t *testing.T) {
	ctx := context.Background()
	repoPath := "/Users/silhouette/codeworks/static_parser"
	pkgPath := "github.com/Silhouette-sophist/static_parser/service"
	// LoadPackages(ctx, &LoadConfig{
	// 	RepoPath: repoPath,
	// 	PkgPath:  pkgPath,
	// 	LoadEnum: LoadAllPkg,
	// })
	LoadPackages(ctx, &LoadConfig{
		RepoPath: repoPath,
		PkgPath:  pkgPath,
		LoadEnum: LoadSpecificPkgWithChild,
	})
}

func TestLoadAllPackages(t *testing.T) {
	ctx := context.Background()
	repoPath := "/Users/silhouette/codeworks/static_parser"
	pkgPath := "github.com/Silhouette-sophist/static_parser/service"
	LoadPackages(ctx, &LoadConfig{
		RepoPath: repoPath,
		PkgPath:  pkgPath,
		LoadEnum: LoadAllPkg,
	})
}
