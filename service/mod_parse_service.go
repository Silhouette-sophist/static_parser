package service

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	vs "github.com/Silhouette-sophist/static_parser/visitor"
	"golang.org/x/mod/modfile"
)

// ModuleInfo 表示一个Go模块的信息
type ModuleInfo struct {
	Path         string        // 模块路径
	Dir          string        // 模块所在目录
	GoVersion    string        // Go版本
	Requires     []Dependency  // 直接依赖
	Replaces     []ReplaceRule // 替换规则
	Imports      []string      // 导入的包（从.go文件中提取）
	Error        error         // 解析过程中发生的错误
	PkgFuncMap   map[string][]*vs.FuncInfo
	PkgVarMap    map[string][]*vs.VarInfo
	PkgStructMap map[string][]*vs.StructInfo
}

// Dependency 表示模块的依赖
type Dependency struct {
	Path     string // 依赖路径
	Version  string // 依赖版本
	Indirect bool   // 是否为间接依赖
}

// ReplaceRule 表示模块的替换规则
type ReplaceRule struct {
	OldPath    string // 原路径
	OldVersion string // 原版本
	NewPath    string // 新路径
	NewVersion string // 新版本
}

// ParseRepo 匹配仓库信息
func ParseRepo(repoPath string) ([]*ModuleInfo, error) {
	modules, err := FindAllModules(repoPath)
	if err != nil {
		return nil, err
	}
	return modules, nil
}

// ParseModule 解析单个go.mod文件
func ParseModule(dir string) (*ModuleInfo, error) {
	start := time.Now()
	defer func() {
		fmt.Printf("ParseModule dir:%s cost: %v\n", dir, time.Since(start))
	}()
	info := &ModuleInfo{
		Dir:          dir,
		PkgFuncMap:   make(map[string][]*vs.FuncInfo),
		PkgVarMap:    make(map[string][]*vs.VarInfo),
		PkgStructMap: make(map[string][]*vs.StructInfo),
	}
	// 读取go.mod文件
	modPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(modPath)
	if err != nil {
		return info, fmt.Errorf("读取go.mod失败: %v", err)
	}
	// 解析go.mod文件
	modeFile, err := modfile.Parse(modPath, data, nil)
	if err != nil {
		return info, fmt.Errorf("解析go.mod失败: %v", err)
	}
	info.Path = modeFile.Module.Mod.Path
	if modeFile.Go != nil {
		info.GoVersion = modeFile.Go.Version
	}
	// 匹配mod文件中内容
	AppendModuleInfo(info)
	// 解析依赖
	for _, req := range modeFile.Require {
		info.Requires = append(info.Requires, Dependency{
			Path:     req.Mod.Path,
			Version:  req.Mod.Version,
			Indirect: req.Indirect,
		})
	}
	// 解析替换规则
	for _, replace := range modeFile.Replace {
		info.Replaces = append(info.Replaces, ReplaceRule{
			OldPath:    replace.Old.Path,
			OldVersion: replace.Old.Version,
			NewPath:    replace.New.Path,
			NewVersion: replace.New.Version,
		})
	}
	// 解析目录中所有.go文件的导入
	imports, err := ParseImportsFromDir(dir)
	if err != nil {
		log.Printf("警告: 解析目录 %s 中的导入失败: %v", dir, err)
	}
	info.Imports = imports
	return info, nil
}

// AppendModuleInfo 解析模块中的所有.go文件
func AppendModuleInfo(modInfo *ModuleInfo) {
	filepath.Walk(modInfo.Dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		rFilePath, err := filepath.Rel(modInfo.Dir, path)
		if err != nil {
			return err
		}
		relDir, err := DeductRelativeDir(modInfo.Dir, path)
		if err != nil {
			return err
		}
		curPkg := modInfo.Path
		if relDir != "" {
			curPkg = modInfo.Path + "/" + relDir
		}
		fileFuncVisitor, err := ParseSingleFile(curPkg, rFilePath, path)
		if err != nil {
			return err
		}
		modInfo.PkgFuncMap[curPkg] = append(modInfo.PkgFuncMap[curPkg], fileFuncVisitor.FileFuncInfos...)
		modInfo.PkgVarMap[curPkg] = append(modInfo.PkgVarMap[curPkg], fileFuncVisitor.FilePkgVars...)
		modInfo.PkgStructMap[curPkg] = append(modInfo.PkgStructMap[curPkg], fileFuncVisitor.FileStructs...)
		return nil
	})
}

// DeductRelativeDir 计算子文件相对于父目录的目录路径（排除文件名）
func DeductRelativeDir(parentDir, childPath string) (string, error) {
	// 计算相对路径
	relPath, err := filepath.Rel(parentDir, childPath)
	if err != nil {
		return "", err
	}
	// 排除文件名，只保留目录部分
	dir := filepath.Dir(relPath)
	// 如果结果是 "."，说明就在父目录下，返回空字符串
	if dir == "." {
		return "", nil
	}
	return dir, nil
}

// 从目录中的所有.go文件解析导入的包
func ParseImportsFromDir(dir string) ([]string, error) {
	var imports []string
	fset := token.NewFileSet()
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 跳过非.go文件和测试文件
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		// 解析文件
		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("解析文件 %s 失败: %v", path, err)
		}
		// 提取导入
		for _, imp := range file.Imports {
			importPath := strings.Trim(imp.Path.Value, `"`)
			imports = append(imports, importPath)
		}
		return nil
	})
	return imports, err
}

// 递归查找目录中的所有模块
func FindAllModules(rootDir string) ([]*ModuleInfo, error) {
	modules := make([]*ModuleInfo, 0)
	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// 跳过隐藏目录
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return fs.SkipDir
		}
		// 检查是否为go.mod文件
		if !d.IsDir() && d.Name() == "go.mod" {
			moduleDir := filepath.Dir(path)
			module, err := ParseModule(moduleDir)
			if err != nil {
				module.Error = err
			}
			modules = append(modules, module)
			// 跳过子目录（避免处理嵌套模块，除非需要）
			return fs.SkipDir
		}
		return nil
	})
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Path < modules[j].Path
	})
	return modules, err
}
