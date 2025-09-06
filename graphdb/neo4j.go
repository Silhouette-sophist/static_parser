package graphdb

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Config 存储Neo4j连接配置
type Config struct {
	URI      string // 连接地址，如 "bolt://localhost:7687"
	Username string // 用户名
	Password string // 密码
	Database string // 数据库名，默认 "graphdb"

	// 连接池配置
	MaxConnectionPoolSize        int           // 最大连接池大小
	ConnectionAcquisitionTimeout time.Duration // 获取连接超时时间
	MaxConnectionLifetime        time.Duration // 连接最大存活时间
}

// Client 封装Neo4j驱动和会话管理
type Client struct {
	driver neo4j.DriverWithContext
	config Config
}

// NewClient 创建新的Neo4j客户端
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	// 设置默认配置
	if cfg.Database == "" {
		cfg.Database = "neo4j"
	}
	if cfg.MaxConnectionPoolSize <= 0 {
		cfg.MaxConnectionPoolSize = 10
	}
	if cfg.ConnectionAcquisitionTimeout <= 0 {
		cfg.ConnectionAcquisitionTimeout = 60 * time.Second
	}
	if cfg.MaxConnectionLifetime <= 0 {
		cfg.MaxConnectionLifetime = 30 * time.Minute
	}

	// 创建认证令牌
	authToken := neo4j.BasicAuth(cfg.Username, cfg.Password, "")

	// 配置驱动
	configurers := []func(*neo4j.Config){
		func(config *neo4j.Config) {
			config.MaxConnectionPoolSize = cfg.MaxConnectionPoolSize
			config.ConnectionAcquisitionTimeout = cfg.ConnectionAcquisitionTimeout
			config.MaxConnectionLifetime = cfg.MaxConnectionLifetime
		},
	}

	// 创建驱动实例
	driver, err := neo4j.NewDriverWithContext(
		cfg.URI,
		authToken,
		configurers...,
	)
	if err != nil {
		return nil, fmt.Errorf("创建驱动失败: %w", err)
	}

	// 验证连接
	if err := driver.VerifyConnectivity(ctx); err != nil {
		_ = driver.Close(ctx)
		return nil, fmt.Errorf("验证连接失败: %w", err)
	}

	return &Client{
		driver: driver,
		config: cfg,
	}, nil
}

// NewSession 创建新会话
func (c *Client) NewSession(ctx context.Context, accessMode neo4j.AccessMode) neo4j.SessionWithContext {
	return c.driver.NewSession(ctx, neo4j.SessionConfig{
		DatabaseName: c.config.Database,
		AccessMode:   accessMode,
	})
}

// Close 关闭驱动
func (c *Client) Close(ctx context.Context) error {
	if c.driver != nil {
		return c.driver.Close(ctx)
	}
	return nil
}

// 全局Client实例
var (
	globalClient *Client
	clientOnce   sync.Once
)

// InitGlobalClient 初始化全局Client
func InitGlobalClient(ctx context.Context, cfg Config) error {
	var err error
	clientOnce.Do(func() {
		globalClient, err = NewClient(ctx, cfg)
	})
	return err
}

// GetGlobalClient 获取全局Client
func GetGlobalClient() *Client {
	if globalClient == nil {
		panic("全局Neo4j客户端未初始化，请先调用InitGlobalClient")
	}
	return globalClient
}

// RepoRepository Repo节点操作封装
type RepoRepository struct {
	client *Client
}

// NewRepoRepository 创建Repo仓库实例
func NewRepoRepository(client *Client) *RepoRepository {
	return &RepoRepository{client: client}
}

// Create 创建Repo节点
func (r *RepoRepository) Create(ctx context.Context, repo Repo) error {
	session := r.client.NewSession(ctx, neo4j.AccessModeWrite)
	defer session.Close(ctx)

	query := `
		CREATE (n:Repo {
			name: $name,
			git_repo: $git_repo,
			description: $description,
			summary: $summary
		})
	`
	params := map[string]interface{}{
		"name":        repo.Name,
		"git_repo":    repo.GitRepo,
		"description": repo.Description,
		"summary":     repo.Summary,
	}

	_, err := session.Run(ctx, query, params)
	return err
}

// GetByName 按名称精确查询仓库（适配 Record.Get 新签名）
func (r *RepoRepository) GetByName(ctx context.Context, name string) (*Repo, error) {
	session := r.client.NewSession(ctx, neo4j.AccessModeRead)
	defer session.Close(ctx)

	query := `
		MATCH (n:Repo {name: $name})
		RETURN n.name AS name, 
		       n.git_repo AS git_repo, 
		       n.description AS description, 
		       n.summary AS summary
	`

	result, err := session.Run(ctx, query, map[string]interface{}{"name": name})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	if result.Next(ctx) {
		record := result.Record()
		repo, err := parseRepoRecord(record)
		if err != nil {
			return nil, fmt.Errorf("解析结果失败: %w", err)
		}
		return repo, nil
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("结果处理错误: %w", err)
	}

	return nil, fmt.Errorf("仓库 %s 不存在", name)
}

// Search 按条件模糊查询仓库（适配新签名）
func (r *RepoRepository) Search(ctx context.Context, keyword string, limit, offset int) ([]Repo, error) {
	session := r.client.NewSession(ctx, neo4j.AccessModeRead)
	defer session.Close(ctx)

	query := `
		MATCH (n:Repo)
		WHERE n.name CONTAINS $keyword OR n.description CONTAINS $keyword
		RETURN n.name AS name, 
		       n.git_repo AS git_repo, 
		       n.description AS description, 
		       n.summary AS summary
		ORDER BY n.name
		SKIP $offset
		LIMIT $limit
	`

	result, err := session.Run(ctx, query, map[string]interface{}{
		"keyword": keyword,
		"limit":   limit,
		"offset":  offset,
	})
	if err != nil {
		return nil, fmt.Errorf("搜索失败: %w", err)
	}

	var repos []Repo
	for result.Next(ctx) {
		repo, err := parseRepoRecord(result.Record())
		if err != nil {
			return nil, fmt.Errorf("解析结果失败: %w", err)
		}
		repos = append(repos, *repo)
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("结果处理错误: %w", err)
	}

	return repos, nil
}

// ListByGitRepo 按 Git 仓库地址查询（适配新签名）
func (r *RepoRepository) ListByGitRepo(ctx context.Context, gitRepo string) ([]Repo, error) {
	session := r.client.NewSession(ctx, neo4j.AccessModeRead)
	defer session.Close(ctx)

	query := `
		MATCH (n:Repo)
		WHERE n.git_repo CONTAINS $gitRepo
		RETURN n.name AS name, 
		       n.git_repo AS git_repo, 
		       n.description AS description, 
		       n.summary AS summary
	`

	result, err := session.Run(ctx, query, map[string]interface{}{"gitRepo": gitRepo})
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	var repos []Repo
	for result.Next(ctx) {
		repo, err := parseRepoRecord(result.Record())
		if err != nil {
			return nil, fmt.Errorf("解析结果失败: %w", err)
		}
		repos = append(repos, *repo)
	}

	if err := result.Err(); err != nil {
		return nil, fmt.Errorf("结果处理错误: %w", err)
	}

	return repos, nil
}

// parseRepoRecord 统一解析 Record 为 Repo 结构体（处理新签名）
// 封装字段获取逻辑，避免重复代码
func parseRepoRecord(record *neo4j.Record) (*Repo, error) {
	// 从记录中获取字段，处理 (any, bool) 返回值
	nameVal, ok := record.Get("name")
	if !ok {
		return nil, errors.New("缺少 name 字段")
	}
	name, ok := nameVal.(string)
	if !ok {
		return nil, errors.New("name 字段类型不是 string")
	}

	gitRepoVal, ok := record.Get("git_repo")
	if !ok {
		return nil, errors.New("缺少 git_repo 字段")
	}
	gitRepo, ok := gitRepoVal.(string)
	if !ok {
		return nil, errors.New("git_repo 字段类型不是 string")
	}

	descriptionVal, ok := record.Get("description")
	if !ok {
		return nil, errors.New("缺少 description 字段")
	}
	description, ok := descriptionVal.(string)
	if !ok {
		return nil, errors.New("description 字段类型不是 string")
	}

	summaryVal, ok := record.Get("summary")
	if !ok {
		return nil, errors.New("缺少 summary 字段")
	}
	summary, ok := summaryVal.(string)
	if !ok {
		return nil, errors.New("summary 字段类型不是 string")
	}

	return &Repo{
		Name:        name,
		GitRepo:     gitRepo,
		Description: description,
		Summary:     summary,
	}, nil
}
