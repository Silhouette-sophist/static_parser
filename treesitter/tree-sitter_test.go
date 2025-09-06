package treesitter

import (
	"context"
	"testing"
)

func TestParseTargetFile(t *testing.T) {
	ctx := context.Background()
	filePath := "/Users/silhouette/codeworks/static_parser/visitor/file_func_visitor.go"
	ParseTargetFile(ctx, filePath, "go")
}
