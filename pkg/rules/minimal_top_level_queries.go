package rules

import (
	"fmt"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// MinimalTopLevelQueries checks that the Query type doesn't have too many top-level fields
type MinimalTopLevelQueries struct{}

// NewMinimalTopLevelQueries creates a new instance of the MinimalTopLevelQueries rule
func NewMinimalTopLevelQueries() *MinimalTopLevelQueries {
	return &MinimalTopLevelQueries{}
}

// Name returns the rule name
func (r *MinimalTopLevelQueries) Name() string {
	return "minimal-top-level-queries"
}

// Description returns what this rule checks
func (r *MinimalTopLevelQueries) Description() string {
	return "Keep the top level queries to a minimum - following Yelp guidelines for better schema organization"
}

// Check validates that the Query type doesn't have excessive top-level fields
func (r *MinimalTopLevelQueries) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Find the Query type
	queryType := schema.Query
	if queryType == nil {
		return errors
	}

	// Count non-introspection fields
	fieldCount := 0
	for _, field := range queryType.Fields {
		if !r.isIntrospectionField(field.Name) {
			fieldCount++
		}
	}

	// Recommend maximum of 10-15 top-level query fields
	maxRecommended := 12
	if fieldCount > maxRecommended {
		line, column := 1, 1
		if queryType.Position != nil {
			line = queryType.Position.Line
			column = queryType.Position.Column
		}

		errors = append(errors, types.LintError{
			Message: fmt.Sprintf("Query type has %d fields, consider reducing to %d or fewer. Group related queries under core types instead.", fieldCount, maxRecommended),
			Location: types.Location{
				Line:   line,
				Column: column,
				File:   source.Name,
			},
			Rule: r.Name(),
		})
	}

	return errors
}

// isIntrospectionField checks if a field is part of GraphQL introspection
func (r *MinimalTopLevelQueries) isIntrospectionField(fieldName string) bool {
	introspectionFields := []string{"__schema", "__type"}
	for _, introspectionField := range introspectionFields {
		if fieldName == introspectionField {
			return true
		}
	}
	return false
}
