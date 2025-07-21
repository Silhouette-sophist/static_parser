package service

import (
	"context"
	"fmt"

	"golang.org/x/tools/go/packages"
)

type LoadEnum int

const (
	LoadCurrentRepo LoadEnum = iota
	LoadSpecificPkg
	LoadSpecificPkgWithChild
	LoadAllPkg
)

type LoadConfig struct {
	RepoPath string
	PkgPath  string
	LoadEnum LoadEnum
}

func LoadPackages(ctx context.Context, loadConfig *LoadConfig) {
	// 配置加载选项
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
			packages.NeedImports | packages.NeedDeps | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo,
		Tests: false,               // 包含测试包
		Dir:   loadConfig.RepoPath, // 当前目录作为基准
	}
	// 加载包
	loadPatterns := make([]string, 0)
	if loadConfig.LoadEnum == LoadCurrentRepo {
		loadPatterns = append(loadPatterns, "./...")
	} else if loadConfig.LoadEnum == LoadAllPkg {
		loadPatterns = append(loadPatterns, "all")
	} else if loadConfig.LoadEnum == LoadSpecificPkg {
		loadPatterns = append(loadPatterns, loadConfig.PkgPath)
	} else if loadConfig.LoadEnum == LoadSpecificPkgWithChild {
		loadPatterns = append(loadPatterns, loadConfig.PkgPath)
		loadPatterns = append(loadPatterns, loadConfig.PkgPath+"/...")
	}
	pkgs, err := packages.Load(cfg, loadPatterns...)
	if err != nil {
		fmt.Printf("加载包失败: %v\n", err)
	}
	// 检查加载过程中是否有错误
	if packages.PrintErrors(pkgs) > 0 {
		fmt.Printf("加载包过程中存在错误\n")
	}
	// 打印加载的包信息
	fmt.Printf("成功加载 %d 个包\n", len(pkgs))
	for _, pkg := range pkgs {
		fmt.Printf("\n包名: %s\n", pkg.Name)
		fmt.Printf("导入路径: %s\n", pkg.PkgPath)
		fmt.Printf("Go源文件数量: %d\n", len(pkg.GoFiles))
		fmt.Printf("依赖数量: %d\n", len(pkg.Imports))
	}
}
