package visitor

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type PkgStaticInfo struct {
	Pkg             string
	FileFuncInfoMap map[string][]*FuncInfo
	FilePkgVarMap   map[string][]*VarInfo
	FileStructMap   map[string][]*StructInfo
}

type BaseAstInfo struct {
	Pkg       string
	RFilePath string
	Name      string
	Content   string
}

type BaseAstPosition struct {
	RFilePath string
	OffSet    int
	Line      int
	Column    int
}

type FileFuncVisitor struct {
	BaseAstInfo
	FileSet       *token.FileSet
	File          *ast.File
	FileBytes     []byte
	FileFuncInfos []*FuncInfo
	FilePkgVars   []*VarInfo
	FileStructs   []*StructInfo
	ImportPkgMap  map[string]string
}

type FuncInfo struct {
	BaseAstInfo
	Receiver      *VarInfo
	Params        []*VarInfo
	Results       []*VarInfo
	StartPosition *BaseAstPosition
	EndPosition   *BaseAstPosition
	ChildCounts   int
	RelatedPkgVar map[string]struct{}
}

type VarInfo struct {
	BaseAstInfo
	Type     string
	Value    string
	BaseType string
}

type StructInfo struct {
	BaseAstInfo
	Fields        []*VarInfo
	StartPosition *BaseAstPosition
	EndPosition   *BaseAstPosition
}

func (f *FileFuncVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.GenDecl:
		// 1.导入声明
		if n.Tok == token.IMPORT {
			for _, spec := range n.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					path := strings.Trim(importSpec.Path.Value, `"`)
					var name string
					if importSpec.Name != nil {
						name = importSpec.Name.Name
					} else {
						lastSplit := strings.LastIndex(path, "/")
						if lastSplit > 0 {
							name = path[lastSplit+1:]
						} else {
							name = path
						}
					}
					f.ImportPkgMap[name] = path
				}
			}
			// 2.包常量和变量声明
		} else if n.Tok == token.CONST || n.Tok == token.VAR {
			for _, spec := range n.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valueSpec.Names {
						varInfo := &VarInfo{
							BaseAstInfo: BaseAstInfo{
								Name:      name.Name,
								RFilePath: f.RFilePath,
								Pkg:       f.Pkg,
								Content:   string(f.FileBytes[valueSpec.Pos()-1 : valueSpec.End()-1]),
							},
							Type: f.parseExprTypeInfo(valueSpec.Type),
						}
						f.FilePkgVars = append(f.FilePkgVars, varInfo)
					}
				}
			}
		} else if n.Tok == token.TYPE {
			startPosition := f.FileSet.Position(n.Pos())
			endPosition := f.FileSet.Position(n.End())
			for _, spec := range n.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					structInfo := &StructInfo{
						BaseAstInfo: BaseAstInfo{
							Name:      typeSpec.Name.Name,
							RFilePath: f.RFilePath,
							Pkg:       f.Pkg,
							Content:   string(f.FileBytes[n.Pos()-1 : n.End()-1]),
						},
						StartPosition: &BaseAstPosition{
							RFilePath: f.RFilePath,
							OffSet:    startPosition.Offset,
							Line:      startPosition.Line,
							Column:    startPosition.Column,
						},
						EndPosition: &BaseAstPosition{
							RFilePath: f.RFilePath,
							OffSet:    endPosition.Offset,
							Line:      endPosition.Line,
							Column:    endPosition.Column,
						},
					}
					if structType, ok := typeSpec.Type.(*ast.StructType); ok {
						f.handleFileList(structType.Fields.List, func(varInfo *VarInfo) {
							structInfo.Fields = append(structInfo.Fields, varInfo)
						})
					}
					f.FileStructs = append(f.FileStructs, structInfo)
				}
			}
		}
	case *ast.FuncDecl:
		funcInfo := f.parseNameFuncInfo(n)
		f.FileFuncInfos = append(f.FileFuncInfos, funcInfo)
		// 采集内部所有匿名函数
		ast.Inspect(n.Body, func(node ast.Node) bool {
			if node == nil {
				return true
			}
			switch nd := node.(type) {
			case *ast.FuncLit:
				childFuncInfo := f.parseAnonymousFuncInfo(nd, funcInfo)
				f.FileFuncInfos = append(f.FileFuncInfos, childFuncInfo)
			}
			return true
		})
		for _, stmt := range n.Body.List {
			funcCallMap := make(map[string]struct{})
			ast.Inspect(stmt, func(node ast.Node) bool {
				if node == nil {
					return true
				}
				switch nd := node.(type) {
				case *ast.CallExpr:
					funcTypeInfo := f.parseExprTypeInfo(nd.Fun)
					funcCallMap[funcTypeInfo] = struct{}{}
				case *ast.Ident:
					funcTypeInfo := f.parseExprTypeInfo(nd)
					if _, ok := funcCallMap[funcTypeInfo]; !ok {
						funcInfo.RelatedPkgVar[funcTypeInfo] = struct{}{}
					}
				case *ast.SelectorExpr:
					funcTypeInfo := f.parseExprTypeInfo(nd)
					if _, ok := funcCallMap[funcTypeInfo]; !ok {
						funcInfo.RelatedPkgVar[funcTypeInfo] = struct{}{}
					}
				}
				return true
			})
		}
	}
	return f
}

func (f *FileFuncVisitor) parseAnonymousFuncInfo(funcLit *ast.FuncLit, parentFuncInfo *FuncInfo) *FuncInfo {
	parentFuncInfo.ChildCounts++
	startPosition := f.FileSet.Position(funcLit.Pos())
	endPosition := f.FileSet.Position(funcLit.End())
	funcInfo := &FuncInfo{
		BaseAstInfo: BaseAstInfo{
			Name:      fmt.Sprintf("%s$%d", parentFuncInfo.Name, parentFuncInfo.ChildCounts),
			RFilePath: parentFuncInfo.RFilePath,
			Pkg:       parentFuncInfo.Pkg,
			Content:   string(f.FileBytes[startPosition.Offset:endPosition.Offset]),
		},
		StartPosition: &BaseAstPosition{
			RFilePath: f.RFilePath,
			OffSet:    startPosition.Offset,
			Line:      startPosition.Line,
			Column:    startPosition.Column,
		},
		EndPosition: &BaseAstPosition{
			RFilePath: f.RFilePath,
			OffSet:    endPosition.Offset,
			Line:      endPosition.Line,
			Column:    endPosition.Column,
		},
	}
	if funcLit.Type.Params != nil {
		f.handleFileList(funcLit.Type.Params.List, func(varInfo *VarInfo) {
			funcInfo.Params = append(funcInfo.Params, varInfo)
		})
	}
	if funcLit.Type.Results != nil {
		f.handleFileList(funcLit.Type.Results.List, func(varInfo *VarInfo) {
			funcInfo.Results = append(funcInfo.Results, varInfo)
		})
	}
	return funcInfo
}

func (f *FileFuncVisitor) parseNameFuncInfo(funcDecl *ast.FuncDecl) *FuncInfo {
	startPosition := f.FileSet.Position(funcDecl.Pos())
	endPosition := f.FileSet.Position(funcDecl.End())
	funcInfo := &FuncInfo{
		BaseAstInfo: BaseAstInfo{
			Name:      funcDecl.Name.Name,
			RFilePath: f.RFilePath,
			Pkg:       f.Pkg,
			Content:   string(f.FileBytes[startPosition.Offset:endPosition.Offset]),
		},
		StartPosition: &BaseAstPosition{
			RFilePath: f.RFilePath,
			OffSet:    startPosition.Offset,
			Line:      startPosition.Line,
			Column:    startPosition.Column,
		},
		EndPosition: &BaseAstPosition{
			RFilePath: f.RFilePath,
			OffSet:    endPosition.Offset,
			Line:      endPosition.Line,
			Column:    endPosition.Column,
		},
		RelatedPkgVar: map[string]struct{}{},
	}
	if funcDecl.Recv != nil {
		f.handleFileList(funcDecl.Recv.List, func(varInfo *VarInfo) {
			funcInfo.Receiver = varInfo
		})
	}
	if funcDecl.Type.Params != nil {
		f.handleFileList(funcDecl.Type.Params.List, func(varInfo *VarInfo) {
			funcInfo.Params = append(funcInfo.Params, varInfo)
		})
	}
	if funcDecl.Type.Results != nil {
		f.handleFileList(funcDecl.Type.Results.List, func(varInfo *VarInfo) {
			funcInfo.Results = append(funcInfo.Results, varInfo)
		})
	}
	return funcInfo
}

func (f *FileFuncVisitor) handleFileList(list []*ast.Field, handleFunc func(varInfo *VarInfo)) {
	for _, field := range list {
		baseTypeInfo := f.parseExprBaseType(field.Type)
		rawBaseTypeInfo := baseTypeInfo
		f.handleCompleteTypeInfo(baseTypeInfo, func(complteTypeInfo string) {
			baseTypeInfo = complteTypeInfo
		})
		typeInfo := f.parseExprTypeInfo(field.Type)
		typeInfo = strings.Replace(typeInfo, rawBaseTypeInfo, baseTypeInfo, -1)
		if len(field.Names) > 0 {
			for _, name := range field.Names {
				handleFunc(&VarInfo{
					BaseAstInfo: BaseAstInfo{
						Name:      name.Name,
						RFilePath: f.RFilePath,
						Pkg:       f.Pkg,
					},
					Type:     typeInfo,
					BaseType: baseTypeInfo,
				})
			}
		} else {
			handleFunc(&VarInfo{
				BaseAstInfo: BaseAstInfo{
					Name:      "_",
					RFilePath: f.RFilePath,
					Pkg:       f.Pkg,
				},
				Type:     typeInfo,
				BaseType: baseTypeInfo,
			})
		}
	}
}

func (f *FileFuncVisitor) handleCompleteTypeInfo(typeInfo string, handleFunc func(complteTypeInfo string)) {
	if typeInfo == "" {
		return
	}
	lastSplit := strings.LastIndex(typeInfo, ".")
	if lastSplit > 0 {
		basePkg := typeInfo[:lastSplit]
		if basePkgPath, ok := f.ImportPkgMap[basePkg]; ok {
			typeInfo = basePkgPath + typeInfo[lastSplit:]
		}
	}
	handleFunc(typeInfo)
}

func (f *FileFuncVisitor) parseExprTypeInfo(expr ast.Expr) string {
	switch n := expr.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.SelectorExpr:
		return f.parseExprTypeInfo(n.X) + "." + n.Sel.Name
	case *ast.StarExpr:
		return "*" + f.parseExprTypeInfo(n.X)
	case *ast.ArrayType:
		return "[]" + f.parseExprTypeInfo(n.Elt)
	case *ast.MapType:
		return "map[" + f.parseExprTypeInfo(n.Key) + "]" + f.parseExprTypeInfo(n.Value)
	case *ast.FuncType:
		return string(f.FileBytes[n.Pos()-1 : n.End()])
	default:
		return ""
	}
}

func (f *FileFuncVisitor) parseExprBaseType(expr ast.Expr) string {
	switch n := expr.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.StarExpr:
		return f.parseExprBaseType(n.X)
	case *ast.SelectorExpr:
		return f.parseExprBaseType(n.X) + "." + n.Sel.Name
	case *ast.ArrayType:
		return f.parseExprBaseType(n.Elt)
	case *ast.MapType:
		return f.parseExprBaseType(n.Value)
	case *ast.FuncType:
		return string(f.FileBytes[n.Pos()-1 : n.End()])
	default:
		return ""
	}
}
