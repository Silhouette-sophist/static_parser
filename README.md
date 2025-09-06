# static_parser
golang通用ast解析器工具

## 工程骨架信息采集

### ast
https://pkg.go.dev/go/ast

### treesitter
https://github.com/tree-sitter/tree-sitter

### abcoder
https://github.com/cloudwego/abcoder

## 调用关系
### ssa采集
https://pkg.go.dev/golang.org/x/tools/go/ssa

### treesitter
https://github.com/tree-sitter/go-tree-sitter

### abcoder
https://github.com/cloudwego/abcoder
实际上是 ast+types

## 评估体系
### 准确性
- 点信息
```text
ssa ~ ast > treesitter
```
- 边信息
```text
ssa > treesitter > ast
```

### 实效性
- 点信息
```text
ast > treesitter >> ssa
```
- 边信息
```text
ast > treesitter >> ssa
```

### 前置依赖
- ast 满足语法要求
- treesitter 满足语法要求
- ssa 满足语法要求+可编译
- abcoder 满足语法要求+可编译