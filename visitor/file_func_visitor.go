package visitor

import (
	"go/ast"
	"go/token"
	"strings"
)

type BaseAstInfo struct {
	Pkg       string
	RFilePath string
	Name      string
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
		if n.Tok == token.IMPORT {
			for _, spec := range n.Specs {
				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					path := importSpec.Path.Value
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
		}
	case *ast.FuncDecl:
		startPosition := f.FileSet.Position(n.Pos())
		endPosition := f.FileSet.Position(n.End())
		funcInfo := &FuncInfo{
			BaseAstInfo: BaseAstInfo{
				Name:      n.Name.Name,
				RFilePath: f.RFilePath,
				Pkg:       f.Pkg,
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
		f.FileFuncInfos = append(f.FileFuncInfos, funcInfo)
		if n.Recv != nil {
			f.handleFileList(n.Recv.List, func(varInfo *VarInfo) {
				funcInfo.Receiver = varInfo
			})
		}
		if n.Type.Params != nil {
			f.handleFileList(n.Type.Params.List, func(varInfo *VarInfo) {
				funcInfo.Params = append(funcInfo.Params, varInfo)
			})
		}
		if n.Type.Results != nil {
			f.handleFileList(n.Type.Results.List, func(varInfo *VarInfo) {
				funcInfo.Results = append(funcInfo.Results, varInfo)
			})
		}
	}
	return f
}

func (f *FileFuncVisitor) handleFileList(list []*ast.Field, handleFunc func(varInfo *VarInfo)) {
	for _, field := range list {
		typeInfo := f.parseExprTypeInfo(field.Type)
		baseTypeInfo := f.parseExprBaseType(field.Type)
		f.handleCompleteTypeInfo(baseTypeInfo, func(complteTypeInfo string) {
			baseTypeInfo = complteTypeInfo
		})
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
