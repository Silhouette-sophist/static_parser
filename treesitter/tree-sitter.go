package treesitter

import (
	"fmt"
	_ "sync/atomic"

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

func ExampleParser_Parse() {
	parser := NewParser()
	defer parser.Close()

	language := NewLanguage(tree_sitter_go.Language())

	parser.SetLanguage(language)

	tree := parser.Parse(
		[]byte(`
			package main


			func main() {
				return
			}
		`),
		nil,
	)
	defer tree.Close()

	rootNode := tree.RootNode()
	fmt.Println(rootNode.ToSexp())
	// Output:
	// (source_file (package_clause (package_identifier)) (function_declaration name: (identifier) parameters: (parameter_list) body: (block (return_statement))))
}

func ExampleParser_ParseWithOptions() {
	parser := NewParser()
	defer parser.Close()

	language := NewLanguage(tree_sitter_go.Language())

	parser.SetLanguage(language)

	sourceCode := []byte(`
			package main

			func main() {
				return
			}
	`)

	readCallback := func(offset int, position Point) []byte {
		return sourceCode[offset:]
	}

	tree := parser.ParseWithOptions(readCallback, nil, nil)
	defer tree.Close()

	rootNode := tree.RootNode()
	fmt.Println(rootNode.ToSexp())
	// Output:
	// (source_file (package_clause (package_identifier)) (function_declaration name: (identifier) parameters: (parameter_list) body: (block (return_statement))))
}
