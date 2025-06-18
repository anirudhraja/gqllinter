package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// NoQueryPrefixes checks that Query fields don't have unnecessary prefixes
type NoQueryPrefixes struct{}

// NewNoQueryPrefixes creates a new instance of the NoQueryPrefixes rule
func NewNoQueryPrefixes() *NoQueryPrefixes {
	return &NoQueryPrefixes{}
}

// Name returns the rule name
func (r *NoQueryPrefixes) Name() string {
	return "no-query-prefixes"
}

// Description returns what this rule checks
func (r *NoQueryPrefixes) Description() string {
	return "Query fields cannot be prefixed with get/list/find as it's implied by being a query"
}

// Check validates that Query fields don't have unnecessary prefixes
func (r *NoQueryPrefixes) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check if there's a Query type
	if schema.Query == nil {
		return errors
	}

	// Forbidden prefixes for query fields
	forbiddenPrefixes := []string{"get", "list", "find", "fetch", "retrieve", "load", "read"}

	// Check each query field
	for _, field := range schema.Query.Fields {
		// Skip introspection fields
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		fieldNameLower := strings.ToLower(field.Name)

		// Check if the field starts with any forbidden prefix
		for _, prefix := range forbiddenPrefixes {
			if strings.HasPrefix(fieldNameLower, prefix) {
				// Make sure it's actually a prefix (not just starts with same letters)
				if len(field.Name) > len(prefix) {
					// Check if the character after the prefix is uppercase (indicating it's a real prefix)
					charAfterPrefix := field.Name[len(prefix)]
					if charAfterPrefix >= 'A' && charAfterPrefix <= 'Z' {
						line, column := 1, 1
						if field.Position != nil {
							line = field.Position.Line
							column = field.Position.Column
						}

						// Suggest a better name
						suggestedName := r.suggestBetterName(field.Name, prefix)

						errors = append(errors, types.LintError{
							Message: fmt.Sprintf("Query field `%s` should not be prefixed with '%s' as it's implied by being a query. Consider `%s` instead.", field.Name, prefix, suggestedName),
							Location: types.Location{
								Line:   line,
								Column: column,
								File:   source.Name,
							},
							Rule: r.Name(),
						})
						break // Only report the first matching prefix
					}
				}
			}
		}
	}

	return errors
}

// suggestBetterName removes the prefix and suggests a better field name
func (r *NoQueryPrefixes) suggestBetterName(fieldName, prefix string) string {
	// Remove the prefix and make the first letter lowercase
	remainder := fieldName[len(prefix):]
	if len(remainder) == 0 {
		return fieldName // Shouldn't happen, but just in case
	}

	// Convert first character to lowercase
	return strings.ToLower(remainder[:1]) + remainder[1:]
}
