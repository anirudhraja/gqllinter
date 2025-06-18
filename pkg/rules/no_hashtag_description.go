package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// NoHashtagDescription checks that descriptions use triple quotes instead of hashtag comments
type NoHashtagDescription struct{}

// NewNoHashtagDescription creates a new instance of the NoHashtagDescription rule
func NewNoHashtagDescription() *NoHashtagDescription {
	return &NoHashtagDescription{}
}

// Name returns the rule name
func (r *NoHashtagDescription) Name() string {
	return "no-hashtag-description"
}

// Description returns what this rule checks
func (r *NoHashtagDescription) Description() string {
	return "Use triple quotes for descriptions instead of hashtag comments, following Yelp guidelines"
}

// Check validates that descriptions use proper syntax
func (r *NoHashtagDescription) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Parse the source to find hashtag comments that should be descriptions
	lines := strings.Split(source.Input, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for hashtag comments that appear to be descriptions
		if strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "# gqllinter") {
			// Check if this appears before a type or field definition
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if r.looksLikeDefinition(nextLine) {
					errors = append(errors, types.LintError{
						Message: fmt.Sprintf("Use triple quotes (\"\"\") for descriptions instead of hashtag comments."),
						Location: types.Location{
							Line:   i + 1, // 1-indexed
							Column: strings.Index(line, "#") + 1,
							File:   source.Name,
						},
						Rule: r.Name(),
					})
				}
			}
		}
	}

	return errors
}

// looksLikeDefinition checks if a line looks like a GraphQL definition
func (r *NoHashtagDescription) looksLikeDefinition(line string) bool {
	return strings.HasPrefix(line, "type ") ||
		strings.HasPrefix(line, "interface ") ||
		strings.HasPrefix(line, "enum ") ||
		strings.HasPrefix(line, "input ") ||
		strings.HasPrefix(line, "scalar ") ||
		strings.HasPrefix(line, "union ") ||
		strings.Contains(line, ":") // field definition
}
