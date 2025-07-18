package main

import (
	"fmt"

	"github.com/Silhouette-sophist/static_parser/service"
)

func main() {
	// 需要先导入 service 包
	modules, err := service.ParseRepo("/Users/silhouette/codeworks/static_parser")
	if err != nil {
		panic(err)
	}
	for _, module := range modules {
		fmt.Println(module.Path)
	}
}
