package main

import (
	"context"
	"time"

	"github.com/Silhouette-sophist/static_parser/graphdb"
	"github.com/Silhouette-sophist/static_parser/zap_log"
)

func main() {
	// 配置本地 Docker 中的默认 Neo4j
	cfg := graphdb.Config{
		URI:      "bolt://localhost:7687", // Docker 映射的本地地址和端口
		Username: "neo4j",                 // 默认用户名
		Password: "chen150928",            // 默认密码，如果你修改过请用新密码

		// 连接池配置使用默认值即可，也可根据需要调整
		MaxConnectionPoolSize:        10,
		ConnectionAcquisitionTimeout: 60 * time.Second,
		MaxConnectionLifetime:        30 * time.Minute,
	}

	// 初始化全局客户端
	background := context.Background()
	if err := graphdb.InitGlobalClient(background, cfg); err != nil {
		zap_log.CtxError(background, "无法连接到 Neo4j: %v", err)
	}

	// 初始化成功后即可使用客户端进行操作
	zap_log.CtxInfo(background, "Neo4j 客户端初始化成功")

	if err := graphdb.NewRepoRepository(graphdb.GetGlobalClient()).Create(background, graphdb.Repo{
		Name:        "test2",
		GitRepo:     "xxxx",
		Summary:     "summary",
		Description: "description",
	}); err != nil {
		zap_log.CtxError(background, "", err)
	}
	zap_log.CtxInfo(background, "Create success")
	_, err := graphdb.NewRepoRepository(graphdb.GetGlobalClient()).GetByName(background, "test")
	if err != nil {
		zap_log.CtxError(background, "", err)
	}
	zap_log.CtxInfo(background, "Get success %v")
}
