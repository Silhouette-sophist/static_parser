package treesitter

import (
	"context"
	"fmt"
	"os"

	"github.com/Silhouette-sophist/static_parser/zap_log"
	. "github.com/tree-sitter/go-tree-sitter"
	tree_sitter_c "github.com/tree-sitter/tree-sitter-c/bindings/go"
	tree_sitter_cpp "github.com/tree-sitter/tree-sitter-cpp/bindings/go"
	tree_sitter_embedded_template "github.com/tree-sitter/tree-sitter-embedded-template/bindings/go"
	tree_sitter_go "github.com/tree-sitter/tree-sitter-go/bindings/go"
	tree_sitter_html "github.com/tree-sitter/tree-sitter-html/bindings/go"
	tree_sitter_java "github.com/tree-sitter/tree-sitter-java/bindings/go"
	tree_sitter_javascript "github.com/tree-sitter/tree-sitter-javascript/bindings/go"
	tree_sitter_json "github.com/tree-sitter/tree-sitter-json/bindings/go"
	tree_sitter_php "github.com/tree-sitter/tree-sitter-php/bindings/go"
	tree_sitter_python "github.com/tree-sitter/tree-sitter-python/bindings/go"
	tree_sitter_ruby "github.com/tree-sitter/tree-sitter-ruby/bindings/go"
	tree_sitter_rust "github.com/tree-sitter/tree-sitter-rust/bindings/go"
)

func getLanguage(name string) *Language {
	switch name {
	case "c":
		return NewLanguage(tree_sitter_c.Language())
	case "cpp":
		return NewLanguage(tree_sitter_cpp.Language())
	case "embedded-template":
		return NewLanguage(tree_sitter_embedded_template.Language())
	case "go":
		return NewLanguage(tree_sitter_go.Language())
	case "html":
		return NewLanguage(tree_sitter_html.Language())
	case "java":
		return NewLanguage(tree_sitter_java.Language())
	case "javascript":
		return NewLanguage(tree_sitter_javascript.Language())
	case "json":
		return NewLanguage(tree_sitter_json.Language())
	case "php":
		return NewLanguage(tree_sitter_php.LanguagePHP())
	case "python":
		return NewLanguage(tree_sitter_python.Language())
	case "ruby":
		return NewLanguage(tree_sitter_ruby.Language())
	case "rust":
		return NewLanguage(tree_sitter_rust.Language())
	default:
		return nil
	}
}

/*
### 核心概念总结
- **解析器（Parser）**：绑定语言语法，负责将代码转为语法树。
- **语法树（Tree）**：源代码的结构化表示，每个节点对应代码中的一个语法单元（如函数、变量、关键字）。
- **节点（Node）**：语法树的基本单位，包含类型（`Type()`）、内容（`Content()`）、子节点（`ChildCount()`/`Child(i)`）等信息。
- **遍历（Walk）**：通过回调函数遍历树中所有节点，筛选目标类型节点并提取信息。

通过这四步，可实现对任意语言代码的结构化分析，具体场景（如提取变量、注释、语法错误检查）只需调整节点类型匹配和信息提取逻辑即可。
*/
func ParseTargetFile(ctx context.Context, filePath string) {
	parser := NewParser()
	defer parser.Close()

	if err := parser.SetLanguage(NewLanguage(tree_sitter_go.Language())); err != nil {
		zap_log.CtxError(ctx, "Failed to set language", err)
	}

	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		zap_log.CtxError(ctx, "Failed to ReadFile", err)
	}

	tree := parser.Parse(fileBytes, nil)
	defer tree.Close()

	rootNode := tree.RootNode()
	childCount := rootNode.ChildCount()
	for i := 0; i < int(childCount); i++ {
		child := rootNode.Child(uint(i))
		zap_log.CtxInfo(ctx, fmt.Sprintf("node %v %v %v", child.Id(), child.GrammarName(), child.Kind()))
	}
	zap_log.CtxInfo(ctx, "successfully parsed target file")
	// 3. 获取根节点，创建 TreeCursor
	cursor := rootNode.Walk() // 调用 Walk() 方法创建游标
	defer cursor.Close()

	// 4. 使用游标遍历节点（适配最新 API）
	for {
		// 关键变化：通过 cursor.Node 直接获取当前节点，无需 CurrentNode() 方法
		currentNode := cursor.Node()
		if currentNode == nil {
			break
		}

		// 打印节点信息
		zap_log.CtxInfo(ctx, fmt.Sprintf("node %v %v %v", currentNode.Id(), currentNode.GrammarName(), currentNode.Kind()))

		// 移动游标逻辑（与之前一致）
		if !cursor.GotoFirstChild() {
			for !cursor.GotoNextSibling() {
				if !cursor.GotoParent() {
					goto end // 遍历结束
				}
			}
		}
	}
end:
}
