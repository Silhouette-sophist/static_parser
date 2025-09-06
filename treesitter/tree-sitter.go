package treesitter

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"github.com/Silhouette-sophist/static_parser/zap_log"
	sitter "github.com/tree-sitter/go-tree-sitter"
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

func getLanguage(name string) *sitter.Language {
	switch name {
	case "c":
		return sitter.NewLanguage(tree_sitter_c.Language())
	case "cpp":
		return sitter.NewLanguage(tree_sitter_cpp.Language())
	case "embedded-template":
		return sitter.NewLanguage(tree_sitter_embedded_template.Language())
	case "go":
		return sitter.NewLanguage(tree_sitter_go.Language())
	case "html":
		return sitter.NewLanguage(tree_sitter_html.Language())
	case "java":
		return sitter.NewLanguage(tree_sitter_java.Language())
	case "javascript":
		return sitter.NewLanguage(tree_sitter_javascript.Language())
	case "json":
		return sitter.NewLanguage(tree_sitter_json.Language())
	case "php":
		return sitter.NewLanguage(tree_sitter_php.LanguagePHP())
	case "python":
		return sitter.NewLanguage(tree_sitter_python.Language())
	case "ruby":
		return sitter.NewLanguage(tree_sitter_ruby.Language())
	case "rust":
		return sitter.NewLanguage(tree_sitter_rust.Language())
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
func ParseTargetFile(ctx context.Context, filePath, langType string) {
	parser := sitter.NewParser()
	defer parser.Close()

	if err := parser.SetLanguage(getLanguage(langType)); err != nil {
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

	// 4. 获取根节点并创建游标
	cursor := rootNode.Walk()
	defer cursor.Close()

	// 5. 遍历所有节点并获取 content
	fmt.Println("节点类型\t内容")
	fmt.Println("------------------------")
	traverseNodes(cursor, fileBytes)
}

// traverse 遍历所有节点
func traverseNodes(cursor *sitter.TreeCursor, code []byte) {
	for {
		// 获取当前节点（根据实际 API 调整）
		node := cursor.Node() // 尝试方法调用

		// 检查节点是否为空
		if isNull(node) {
			break
		}

		// 获取节点类型和内容
		nodeType := getNodeType(node)
		content := getNodeContent(node, code)

		// 打印信息
		fmt.Printf("%-18s %q\n", nodeType, content)

		// 递归处理子节点
		if cursor.GotoFirstChild() {
			traverseNodes(cursor, code)
			cursor.GotoParent()
		}

		// 移动到下一个兄弟节点
		if !cursor.GotoNextSibling() {
			break
		}
	}
}

// 安全获取节点类型（处理空指针和类型转换问题）
func getNodeType(node *sitter.Node) string {
	if node == nil {
		return "null_node"
	}

	// 方法1：尝试通过反射获取内部类型信息（适用于部分版本）
	rv := reflect.ValueOf(*node)
	if rv.Kind() == reflect.Struct {
		// 尝试查找名为"type_"或"typ"的字段（不同版本可能有不同命名）
		for _, fieldName := range []string{"type_", "typ", "type"} {
			if field := rv.FieldByName(fieldName); field.IsValid() {
				if field.Kind() == reflect.Uintptr && field.Uint() != 0 {
					// 尝试将类型指针转换为字符串（谨慎操作）
					typePtr := unsafe.Pointer(uintptr(field.Uint()))
					return parseTypeFromPointer(typePtr)
				}
			}
		}
	}

	// 方法2：作为最后的 fallback，返回节点的位置哈希（至少避免崩溃）
	return fmt.Sprintf("node_%d_%d", node.StartByte(), node.EndByte())
}

// 从类型指针解析类型字符串（增加安全判断）
func parseTypeFromPointer(ptr unsafe.Pointer) string {
	if ptr == nil {
		return "unknown_type"
	}

	// 不同版本的 Tree-sitter 类型指针结构不同，这里尝试两种常见结构
	// 结构1：直接指向字符串
	strPtr := (*string)(ptr)
	if strPtr != nil && *strPtr != "" {
		return *strPtr
	}

	// 结构2：指向包含字符串的结构体（如 { name: "function_declaration", ... }）
	type typeInfo struct {
		name string
	}
	infoPtr := (*typeInfo)(ptr)
	if infoPtr != nil && infoPtr.name != "" {
		return infoPtr.name
	}

	return "unknown_type"
}

// 判断节点是否为空
func isNull(node *sitter.Node) bool {
	return node == nil || (node.StartByte() == 0 && node.EndByte() == 0)
}

// 通过位置信息提取节点内容
func getNodeContent(node *sitter.Node, code []byte) string {
	if isNull(node) {
		return ""
	}
	start := int(node.StartByte())
	end := int(node.EndByte())
	if start < 0 || end > len(code) || start >= end {
		return ""
	}
	return string(code[start:end])
}
