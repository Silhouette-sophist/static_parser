package service

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	vs "github.com/Silhouette-sophist/static_parser/visitor"
)

const (
	FuncType = "func"
)

var fvi = FileVisitInfo{}

type FileVisitInfo struct {
	FilePath  string
	FuncInfos []*vs.FuncInfo
}

// ParseSingleFile 解析单个文件中的函数信息
func ParseSingleFile(curPkg, rFilePath, filePath string) (*vs.FileFuncVisitor, error) {
	fileSet := token.NewFileSet()
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	file, err := parser.ParseFile(fileSet, filePath, fileBytes, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	fileName := filepath.Base(filePath)
	fileFuncVisitor := &vs.FileFuncVisitor{
		BaseAstInfo: vs.BaseAstInfo{
			RFilePath: rFilePath,
			Pkg:       curPkg,
			Name:      fileName,
			Content:   string(fileBytes),
		},
		FileSet:      fileSet,
		File:         file,
		FileBytes:    fileBytes,
		ImportPkgMap: make(map[string]string),
	}
	ast.Walk(fileFuncVisitor, file)
	sort.Slice(fileFuncVisitor.FileFuncInfos, func(i, j int) bool {
		return fileFuncVisitor.FileFuncInfos[i].StartPosition.OffSet < fileFuncVisitor.FileFuncInfos[j].StartPosition.OffSet
	})
	return fileFuncVisitor, nil
}

// ParseSingleDir 匹配单个包的数据
func ParseSingleDir(dirPath string) error {
	fileSet := token.NewFileSet()
	pkgs, err := parser.ParseDir(fileSet, dirPath, func(fi os.FileInfo) bool {
		return strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
	}, parser.ParseComments)
	if err != nil {
		return err
	}
	for pkgId, pkg := range pkgs {
		for fileId, file := range pkg.Files {
			//TODO 虽然可以对file做ast.Walk但是没法比较好的做module匹配&传递包名信息
			for declId, decl := range file.Decls {
				fmt.Printf("pkgId: %v, fileId: %v, declId: %v, decl: %v\n", pkgId, fileId, declId, decl)
			}
		}
	}
	return nil
}

// ParseAllDirs 匹配所有目录中的函数
func ParseAllDirs(rootDir string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// 跳过非目录
		if !info.IsDir() {
			return nil
		}
		// 检查目录中是否有Go文件
		files, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		hasGoFiles := false
		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".go" {
				hasGoFiles = true
				break
			}
		}
		if hasGoFiles {
			// 解析当前目录作为一个包
			if err := ParseSingleDir(path); err != nil {
				fmt.Printf("解析包 %s 失败: %v\n", path, err)
			}
		}
		return nil
	})
}
