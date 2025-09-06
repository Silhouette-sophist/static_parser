package graphdb

// Repo 节点结构体
type Repo struct {
	Name        string
	GitRepo     string
	Description string
	Summary     string
	UniqueId    string
}

// Branch 分支结构体
type Branch struct {
	Name        string
	Description string
	Summary     string
	UniqueId    string
}

// Commit commit结构体
type Commit struct {
	Name        string
	Description string
	Summary     string
	UniqueId    string
}

// Tag tag结构体
type Tag struct {
	Name        string
	Description string
	Summary     string
	UniqueId    string
}
