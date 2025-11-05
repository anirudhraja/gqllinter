package rules

import (
	"fmt"
	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// CommonSchemaRules checks lint issues present in common schema shared by all subgraphs
type CommonSchemaRules struct{}

// NewCommonSchemaRules creates a new instance of the CommonSchemaRules rule
func NewCommonSchemaRules() *CommonSchemaRules {
	return &CommonSchemaRules{}
}

// Name returns the rule name
func (r *CommonSchemaRules) Name() string {
	return "common-schema-lint"
}

// Description returns what this rule checks
func (r *CommonSchemaRules) Description() string {
	return "CommonSchemaRules contains basic rules that a common schema should follow like .. it shouldn't contain entities, Unions, RPCs.."
}

// Check validates common directive rules
func (r *CommonSchemaRules) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	for _, typeDef := range schema.Types {
		line, column := r.getPositionOfDefinition(typeDef)
		switch typeDef.Kind {
		case ast.Object, ast.Interface:
			if hasKeyDirective(typeDef) {
				errors = append(errors,
					buildLintError(fmt.Sprintf("The Definition of Entity %v is not allowed in Common schema- since no Entity is allowed to be inside common schema", typeDef.Name), r.Name(), source, line, column))
			}
		}
	}

	if schema.Query != nil {
		line, column := r.getPositionOfDefinition(schema.Query)
		errors = append(errors,
			buildLintError("The Defining of Query is restricted inside common schema", r.Name(), source, line, column))
	}
	if schema.Mutation != nil {
		line, column := r.getPositionOfDefinition(schema.Mutation)
		errors = append(errors,
			buildLintError("The Defining of Mutation is restricted inside common schema", r.Name(), source, line, column))
	}

	return errors
}

func (r *CommonSchemaRules) getPositionOfDefinition(typedef *ast.Definition) (int, int) {
	line, column := 1, 1
	if typedef.Position != nil {
		line = typedef.Position.Line
		column = typedef.Position.Column
	}
	return line, column
}

func buildLintError(message string, ruleName string, source *ast.Source, line int, column int) types.LintError {
	return types.LintError{
		Message: message,
		Location: types.Location{
			Line:   line,
			Column: column,
			File:   source.Name,
		},
		Rule: ruleName,
	}
}

func hasKeyDirective(typeDef *ast.Definition) bool {
	for _, directive := range typeDef.Directives {
		if directive.Name == "key" {
			return true
		}
	}
	return false
}
