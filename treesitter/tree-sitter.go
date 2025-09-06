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
}
