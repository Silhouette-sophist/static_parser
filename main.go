package main

import (
	"fmt"

	"github.com/Silhouette-sophist/static_parser/service"
)

func main() {
	// 需要先导入 service 包
	funcInfos, err := service.ParseFileFunc("./visitor/file_func_visitor.go")
	if err != nil {
		panic(err)
	}
	for _, funcInfo := range funcInfos {
		fmt.Println(funcInfo)
	}
}
