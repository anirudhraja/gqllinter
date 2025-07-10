package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// TypesHaveDescriptions checks that all types have descriptions
type TypesHaveDescriptions struct{}

// NewTypesHaveDescriptions creates a new instance of the TypesHaveDescriptions rule
func NewTypesHaveDescriptions() *TypesHaveDescriptions {
	return &TypesHaveDescriptions{}
}

// Name returns the rule name
func (r *TypesHaveDescriptions) Name() string {
	return "types-have-descriptions"
}

// Description returns what this rule checks
func (r *TypesHaveDescriptions) Description() string {
	return "All types should have descriptions to explain their purpose"
}

// Check validates that all types have descriptions
func (r *TypesHaveDescriptions) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check object types
	for _, def := range schema.Types {
		if def.Description == "" && !def.BuiltIn {
			inputArr := strings.Split(def.Position.Src.Input, "\n")
			pos := def.Position.Line - 1
			// description with (""") is not supported by GQL for extend type - hence skipping
			if (pos >= 0 || pos < len(inputArr)) && strings.HasPrefix(strings.TrimSpace(inputArr[pos]), "extend type") {
				continue
			}
			// For types, position information might not be available in the schema built from source
			// We'll use line 1 as default
			line, column := 1, 1
			if def.Position != nil {
				line = def.Position.Line
				column = def.Position.Column
			}

			errors = append(errors, types.LintError{
				Message: fmt.Sprintf("The object type `%s` is missing a description.", def.Name),
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
