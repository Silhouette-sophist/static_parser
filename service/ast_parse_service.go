package service

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"

	"github.com/Silhouette-sophist/static_parser/visitor"
)

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
		},
		FileSet:   fileSet,
		File:      file,
		FileBytes: fileBytes,
	}
	ast.Walk(fileFuncVisitor, file)
	return fileFuncVisitor.FileFuncInfos, nil
}
