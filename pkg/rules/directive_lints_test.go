package rules

import (
	"testing"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

// Helper function to parse GraphQL schema from string for directive_lints tests
func parseSchemaForDirectiveLints(t *testing.T, schemaStr string) (*ast.Schema, *ast.Source) {
	source := &ast.Source{
		Name:  "test.graphql",
		Input: schemaStr,
	}

	schema, err := gqlparser.LoadSchema(source)
	if err != nil {
		t.Fatalf("Failed to parse schema: %v", err)
	}

	return schema, source
}

// Helper function to run directive_lints rule and return errors
func runDirectiveLintRule(t *testing.T, rule types.Rule, schemaStr string) []types.LintError {
	schema, source := parseSchemaForDirectiveLints(t, schemaStr)
	return rule.Check(schema, source)
}

func TestDirectivesCommonLint(t *testing.T) {
	rule := NewDirectivesCommonLint()

	t.Run("should flag object with both @key and @shareable directives", func(t *testing.T) {
		schema := `
		directive @key(fields: String!) on OBJECT
		directive @shareable on OBJECT
		
		type User @key(fields: "id") @shareable {
			id: ID!
			name: String!
		}
		`
		errors := runDirectiveLintRule(t, rule, schema)
		if countRuleErrors(errors, "common-directives-lint") != 1 {
			t.Errorf("Expected exactly 1 error for object with both @key and @shareable, got %d", countRuleErrors(errors, "common-directives-lint"))
		}

		// Check the error message
		if len(errors) > 0 {
			expectedMessage := "The object User cannot have both @key and @shareable directives. These directives are not supported together."
			if errors[0].Message != expectedMessage {
				t.Errorf("Expected error message '%s', got '%s'", expectedMessage, errors[0].Message)
			}
		}
	})

	t.Run("should pass when object has only @key directive", func(t *testing.T) {
		schema := `
		directive @key(fields: String!) on OBJECT
		
		type User @key(fields: "id") {
			id: ID!
			name: String!
		}
		`
		errors := runDirectiveLintRule(t, rule, schema)
		if countRuleErrors(errors, "common-directives-lint") != 0 {
			t.Error("Expected no errors for object with only @key directive")
		}
	})

	t.Run("should pass when object has neither @key nor @shareable directives", func(t *testing.T) {
		schema := `
		directive @entity on OBJECT
		
		type User @entity {
			id: ID!
			name: String!
		}
		`
		errors := runDirectiveLintRule(t, rule, schema)
		if countRuleErrors(errors, "common-directives-lint") != 0 {
			t.Error("Expected no errors for object with neither @key nor @shareable directives")
		}
	})

	t.Run("should pass when @shareable is only on object types", func(t *testing.T) {
		schema := `
		directive @shareable on OBJECT
		
		type User @shareable {
			id: ID!
			name: String!
		}
		
		type Product @shareable {
			sku: String!
			name: String!
		}
		
		scalar DateTime
		enum Status { ACTIVE INACTIVE }
		interface Node { id: ID! }
		input UserInput { name: String! }
		`
		errors := runDirectiveLintRule(t, rule, schema)
		if countRuleErrors(errors, "common-directives-lint") != 0 {
			t.Error("Expected no errors when @shareable is only used on object types")
		}
	})
}
