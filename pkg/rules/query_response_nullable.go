package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// QueryResponseNullable checks that query root fields are nullable
type QueryResponseNullable struct{}

// NewQueryResponseNullable creates a new instance of the QueryResponseNullable rule
func NewQueryResponseNullable() *QueryResponseNullable {
	return &QueryResponseNullable{}
}

// Name returns the rule name
func (r *QueryResponseNullable) Name() string {
	return "query-response-nullable"
}

// Description returns what this rule checks
func (r *QueryResponseNullable) Description() string {
	return "Query root response fields should be nullable."
}

// Check validates that root level query response fields are nullable
func (r *QueryResponseNullable) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find the query type
	queryType := schema.Types["Query"]
	if queryType == nil {
		return errors // No queries to check
	}

	for _, field := range queryType.Fields {
		// Skip introspection fields
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		// Enforce that the top-level field is nullable. Inner/nested types are not checked.
		if field.Type != nil && field.Type.NonNull {
			line, column := 1, 1
			if field.Position != nil {
				line = field.Position.Line
				column = field.Position.Column
			}

			suggestion := r.makeNullable(field.Type)

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf(
					"Query root field `%s` should be nullable (`%s` instead of `%s`) to prevent nulling out entire query response due to missing data.",
					field.Name, suggestion, field.Type.String(),
				),
				Location: types.Location{
					Line:   line,
					Column: column,
					File:   source.Name,
				},
				Rule: r.Name(),
			})
		}
	}

	return errors
}

// makeNullable converts a non-null type to nullable
func (r *QueryResponseNullable) makeNullable(fieldType *ast.Type) string {
	typeStr := fieldType.String()
	// Remove the ! at the end if present
	if strings.HasSuffix(typeStr, "!") {
		return typeStr[:len(typeStr)-1]
	}

	return typeStr
}
