package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// UnsupportedDirectives checks that no unsupported directives are used
type UnsupportedDirectives struct{}

// NewUnsupportedDirectives creates a new instance of the UnsupportedDirectives rule
func NewUnsupportedDirectives() *UnsupportedDirectives {
	return &UnsupportedDirectives{}
}

// Name returns the rule name
func (r *UnsupportedDirectives) Name() string {
	return "unsupported-directives"
}

// Description returns what this rule checks
func (r *UnsupportedDirectives) Description() string {
	return "No unsupported directives should be used in the schemas"
}

// Check validates no unsupported directives are used
func (r *UnsupportedDirectives) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// List of unsupported directives
	unsupportedDirectives := []string{"inaccessible", "external", "requires", "provides"}

	// Check directives on schema types
	for _, def := range schema.Types {
		// Skip built-in types and introspection types
		if def.BuiltIn || strings.HasPrefix(def.Name, "__") {
			continue
		}

		// Check directives on the type itself
		errors = append(errors, r.checkDirectivesOnElement(def.Directives, fmt.Sprintf("type %s", def.Name), def.Position, source, unsupportedDirectives)...)

		// Check directives on object, interface, and input object fields
		if def.Kind == ast.Object || def.Kind == ast.Interface || def.Kind == ast.InputObject {
			for _, field := range def.Fields {
				// Skip built-in fields and introspection fields
				if strings.HasPrefix(field.Name, "__") {
					continue
				}

				errors = append(errors, r.checkDirectivesOnElement(field.Directives, fmt.Sprintf("field %s.%s", def.Name, field.Name), field.Position, source, unsupportedDirectives)...)

				// Check directives on field arguments
				for _, arg := range field.Arguments {
					errors = append(errors, r.checkDirectivesOnElement(arg.Directives, fmt.Sprintf("argument %s.%s(%s:)", def.Name, field.Name, arg.Name), arg.Position, source, unsupportedDirectives)...)
				}
			}
		}

		// Check directives on enum values
		if def.Kind == ast.Enum {
			for _, enumValue := range def.EnumValues {
				errors = append(errors, r.checkDirectivesOnElement(enumValue.Directives, fmt.Sprintf("enum value %s.%s", def.Name, enumValue.Name), enumValue.Position, source, unsupportedDirectives)...)
			}
		}
	}

	return errors
}

// checkDirectivesOnElement checks if any directives in the list are unsupported
func (r *UnsupportedDirectives) checkDirectivesOnElement(directives ast.DirectiveList, elementName string, position *ast.Position, source *ast.Source, unsupportedDirectives []string) []types.LintError {
	var errors []types.LintError

	for _, directive := range directives {
		for _, unsupported := range unsupportedDirectives {
			if directive.Name == unsupported {
				line, column := 1, 1
				if position != nil {
					line = position.Line
					column = position.Column
				}

				errors = append(errors, types.LintError{
					Message: fmt.Sprintf("The %s uses unsupported directive @%s. This directive is not supported in this schema.", elementName, unsupported),
					Location: types.Location{
						Line:   line,
						Column: column,
						File:   source.Name,
					},
					Rule: r.Name(),
				})
			}
		}
	}

	return errors
}
