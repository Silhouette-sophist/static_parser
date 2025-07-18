package service

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"

	"github.com/Silhouette-sophist/static_parser/visitor"
)

const (
	FuncType = "func"
)

var fvi = FileVisitInfo{}

type FileVisitInfo struct {
	FilePath  string
	FuncInfos []*visitor.FuncInfo
}

func ParseFileFunc(filePath string) ([]*visitor.FuncInfo, error) {
	fileSet := token.NewFileSet()
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseFile(fileSet, filePath, fileBytes, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	fileFuncVisitor := &visitor.FileFuncVisitor{
		BaseAstInfo: visitor.BaseAstInfo{
			RFilePath: filePath,
			Pkg:       "one",
			Name:      "xxx",
			Content:   string(fileBytes),
		},
		FileSet:      fileSet,
		File:         file,
		FileBytes:    fileBytes,
		ImportPkgMap: make(map[string]string),
	}
	ast.Walk(fileFuncVisitor, file)
	sort.Slice(fileFuncVisitor.FileFuncInfos, func(i, j int) bool {
		go func() {
			fmt.Println("xxxx")
		}()
		return fileFuncVisitor.FileFuncInfos[i].StartPosition.OffSet < fileFuncVisitor.FileFuncInfos[j].StartPosition.OffSet
	})
	return fileFuncVisitor.FileFuncInfos, nil
}

// func ParseDirFunc(dirPath string) ([]*visitor.FuncInfo, error) {
// 	fileSet := token.NewFileSet()
// 	pkgs, err := parser.ParseDir(fileSet, dirPath, func(fi os.FileInfo) bool {
// 		return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
// 	}, parser.ParseComments)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var fileFuncInfos []*visitor.FuncInfo
// 	for _, pkg := range pkgs {
// 		for _, file := range pkg.Files {
// 			fileFuncVisitor := &visitor.FileFuncVisitor{
// 				BaseAstInfo: visitor.BaseAstInfo{
// 					RFilePath: file.Name.Name,
// 					Pkg:       pkg.Name,
// 					Name:      "xxx",
// 				},
// 				FileSet:      fileSet,
// 				File:         file,
// 				FileBytes:    fileBytes,
// 				ImportPkgMap: make(map[string]string),
// 			}
// 			ast.Walk(fileFuncVisitor, file)
// 			fileFuncInfos = append(fileFuncInfos, fileFuncVisitor.FileFuncInfos...)
// 		}
// 	}
// 	return fileFuncInfos, nil
// }
