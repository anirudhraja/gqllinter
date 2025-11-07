package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
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

	isRootType := collectRootTypeNames(schema)

	for _, def := range schema.Types {
		// skip built-ins and root operation types
		if def.BuiltIn || isRootType[def.Name] {
			continue
		}
		if def.Description == "" {
			inputArr := strings.Split(def.Position.Src.Input, "\n")
			pos := def.Position.Line - 1
			// description with (""") is not supported by GQL for extend type* - hence skipping
			if (pos >= 0 || pos < len(inputArr)) && isExtendType(strings.TrimSpace(inputArr[pos])) {
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

func collectRootTypeNames(schema *ast.Schema) map[string]bool {
	names := map[string]bool{}
	if schema == nil {
		return names
	}

	if schema.Query != nil && schema.Query.Name != "" {
		names[schema.Query.Name] = true
	}
	if schema.Mutation != nil && schema.Mutation.Name != "" {
		names[schema.Mutation.Name] = true
	}
	if schema.Subscription != nil && schema.Subscription.Name != "" {
		names[schema.Subscription.Name] = true
	}
	return names
}

func isExtendType(line string) bool {
	return strings.HasPrefix(line, "extend type ") ||
		strings.HasPrefix(line, "extend interface ") ||
		strings.HasPrefix(line, "extend input ") ||
		strings.HasPrefix(line, "extend enum ") ||
		strings.HasPrefix(line, "extend union ") ||
		strings.HasPrefix(line, "extend scalar ")
}
