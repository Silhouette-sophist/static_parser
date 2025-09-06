package graphdb

type AstIndex struct {
	Repo      string
	Version   string
	Package   string
	Directory string
	UniqueId  string // 唯一索引ast元素id
}

type AstModule struct {
	AstIndex
	Module      string
	AstPackages []AstPackage
}

type AstPackage struct {
	AstIndex
	AstFiles []AstFile
}

type AstFile struct {
	AstIndex
	RelPath      string
	Name         string
	AstStructs   []AstStruct
	AstInterface []AstInterface
	AstFuncInfos []AstFuncInfo
	AstVariables []AstVariable
}

type AstStruct struct {
	AstIndex
	Name    string
	Content string
}

type AstInterface struct {
	AstIndex
	Name    string
	Content string
}

type AstVariable struct {
	AstIndex
	Name    string
	Content string
}

type AstFuncInfo struct {
	AstIndex
	Name     string
	Content  string
	Params   []AstVariable
	Results  []AstVariable
	Receiver AstVariable
}
