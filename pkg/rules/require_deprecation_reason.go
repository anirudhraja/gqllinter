package rules

import (
	"fmt"
	"strings"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2/ast"
)

// RequireDeprecationReason checks that deprecated fields have proper deprecation reasons
type RequireDeprecationReason struct{}

// NewRequireDeprecationReason creates a new instance of the RequireDeprecationReason rule
func NewRequireDeprecationReason() *RequireDeprecationReason {
	return &RequireDeprecationReason{}
}

// Name returns the rule name
func (r *RequireDeprecationReason) Name() string {
	return "require-deprecation-reason"
}

// Description returns what this rule checks
func (r *RequireDeprecationReason) Description() string {
	return "Require deprecation reasons for deprecated fields - following Guild best practices"
}

// Check validates that deprecated fields have meaningful deprecation reasons
func (r *RequireDeprecationReason) Check(schema *ast.Schema, source *ast.Source) []types.LintError {
	var errors []types.LintError

	// Check fields in object types and interfaces
	for _, def := range schema.Types {
		if def.Kind == ast.Object || def.Kind == ast.Interface {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			for _, field := range def.Fields {
				// Check if field has @deprecated directive
				deprecatedDirective := r.findDeprecatedDirective(field.Directives)
				if deprecatedDirective != nil {
					reason := r.getDeprecationReason(deprecatedDirective)

					if reason == "" {
						line, column := 1, 1
						if field.Position != nil {
							line = field.Position.Line
							column = field.Position.Column
						}

						errors = append(errors, types.LintError{
							Message: fmt.Sprintf("Deprecated field `%s.%s` must include a deprecation reason explaining why it's deprecated and what to use instead.", def.Name, field.Name),
							Location: types.Location{
								Line:   line,
								Column: column,
								File:   source.Name,
							},
							Rule: r.Name(),
						})
					} else if r.isGenericReason(reason) {
						line, column := 1, 1
						if field.Position != nil {
							line = field.Position.Line
							column = field.Position.Column
						}

						errors = append(errors, types.LintError{
							Message: fmt.Sprintf("Deprecated field `%s.%s` has a generic deprecation reason '%s'. Provide specific guidance on what to use instead.", def.Name, field.Name, reason),
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
		}

		// Check enum values
		if def.Kind == ast.Enum {
			// Skip introspection types
			if strings.HasPrefix(def.Name, "__") {
				continue
			}

			for _, enumValue := range def.EnumValues {
				deprecatedDirective := r.findDeprecatedDirective(enumValue.Directives)
				if deprecatedDirective != nil {
					reason := r.getDeprecationReason(deprecatedDirective)

					if reason == "" {
						line, column := 1, 1
						if enumValue.Position != nil {
							line = enumValue.Position.Line
							column = enumValue.Position.Column
						}

						errors = append(errors, types.LintError{
							Message: fmt.Sprintf("Deprecated enum value `%s.%s` must include a deprecation reason.", def.Name, enumValue.Name),
							Location: types.Location{
								Line:   line,
								Column: column,
								File:   source.Name,
							},
							Rule: r.Name(),
						})
					} else if r.isGenericReason(reason) {
						line, column := 1, 1
						if enumValue.Position != nil {
							line = enumValue.Position.Line
							column = enumValue.Position.Column
						}

						errors = append(errors, types.LintError{
							Message: fmt.Sprintf("Deprecated enum value `%s.%s` has a generic deprecation reason '%s'. Provide specific guidance.", def.Name, enumValue.Name, reason),
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
		}
	}

	return errors
}

// findDeprecatedDirective finds the @deprecated directive in a list of directives
func (r *RequireDeprecationReason) findDeprecatedDirective(directives ast.DirectiveList) *ast.Directive {
	for _, directive := range directives {
		if directive.Name == "deprecated" {
			return directive
		}
	}
	return nil
}

// getDeprecationReason extracts the reason from a @deprecated directive
func (r *RequireDeprecationReason) getDeprecationReason(directive *ast.Directive) string {
	// Look for the "reason" argument
	for _, arg := range directive.Arguments {
		if arg.Name == "reason" && arg.Value != nil {
			if arg.Value.Kind == ast.StringValue {
				return strings.TrimSpace(arg.Value.Raw)
			}
		}
	}
	return ""
}

// isGenericReason checks if a deprecation reason is too generic to be helpful
func (r *RequireDeprecationReason) isGenericReason(reason string) bool {
	genericReasons := []string{
		"deprecated", "no longer supported", "legacy", "old", "unused",
		"removed", "obsolete", "outdated", "use something else", "will be removed",
	}

	reasonLower := strings.ToLower(strings.TrimSpace(reason))

	// Check for exact matches or very short reasons
	if len(reasonLower) < 10 {
		return true
	}

	for _, generic := range genericReasons {
		if reasonLower == generic || strings.Contains(reasonLower, generic) {
			return true
		}
	}

	// Check if it doesn't provide alternative guidance
	hasGuidance := strings.Contains(reasonLower, "use ") ||
		strings.Contains(reasonLower, "instead") ||
		strings.Contains(reasonLower, "replace") ||
		strings.Contains(reasonLower, "migrate") ||
		strings.Contains(reasonLower, "switch to")

	return !hasGuidance
}
