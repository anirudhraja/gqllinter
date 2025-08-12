package rules

import (
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestBasicLint(t *testing.T) {
	tests := []struct {
		name           string
		schema         string
		expectedErrors int
		expectedMsg    string
	}{
		{
			name: "Valid: @error types without Error suffix",
			schema: `
				directive @error on OBJECT

				type UserNotFound @error {
					code: String!
					message: String!
				}

				type ProductUnavailable @error {
					reason: String!
				}

				type RegularType {
					id: ID!
					name: String!
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: @error type with Error suffix",
			schema: `
				directive @error on OBJECT

				type UserNotFoundError @error {
					code: String!
					message: String!
				}

				type RegularType {
					id: ID!
					name: String!
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Type 'UserNotFoundError' has @error directive but contains 'Error' in its name",
		},
		{
			name: "Invalid: Multiple @error types with Error suffix",
			schema: `
				directive @error on OBJECT

				type UserNotFoundError @error {
					code: String!
					message: String!
				}

				type ProductError @error {
					reason: String!
				}

				type ValidationError @error {
					field: String!
					message: String!
				}
			`,
			expectedErrors: 3,
			expectedMsg:    "has @error directive but contains 'Error' in its name",
		},
		{
			name: "Invalid: Case-insensitive Error suffix detection",
			schema: `
				directive @error on OBJECT

				type UserNotFoundERROR @error {
					code: String!
					message: String!
				}

				type Producterror @error {
					reason: String!
				}
			`,
			expectedErrors: 2,
			expectedMsg:    "has @error directive but contains 'Error' in its name",
		},
		{
			name: "Valid: Regular types can have Error suffix",
			schema: `
				directive @error on OBJECT

				type UserNotFoundError {
					code: String!
					message: String!
				}

				type SystemError {
					level: String!
					description: String!
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: Type contains PropertiesBy substring",
			schema: `
				directive @error on OBJECT

				type UserPropertiesByLocation {
					location: String!
					users: [String!]!
				}

				type RegularType {
					id: ID!
					name: String!
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Type 'UserPropertiesByLocation' contains 'PropertiesBy' substring which is not allowed",
		},
		{
			name: "Invalid: Multiple types with PropertiesBy substring",
			schema: `
				directive @error on OBJECT

				type UserPropertiesByLocation {
					location: String!
					users: [String!]!
				}

				type ProductPropertiesByCategory {
					category: String!
					products: [String!]!
				}

				type OrderPropertiesByStatus {
					status: String!
					orders: [String!]!
				}
			`,
			expectedErrors: 3,
			expectedMsg:    "contains 'PropertiesBy' substring which is not allowed",
		},
		{
			name: "Invalid: Case-insensitive PropertiesBy detection",
			schema: `
				directive @error on OBJECT

				type UserpropertiesbyLocation {
					location: String!
					users: [String!]!
				}

				type ProductPROPERTIESBYCategory {
					category: String!
					products: [String!]!
				}

				type OrderPropertiesByStatus {
					status: String!
					orders: [String!]!
				}
			`,
			expectedErrors: 3,
			expectedMsg:    "contains 'PropertiesBy' substring which is not allowed",
		},
		{
			name: "Invalid: Both Error suffix and PropertiesBy violations",
			schema: `
				directive @error on OBJECT

				type UserNotFoundError @error {
					code: String!
					message: String!
				}

				type ProductPropertiesByCategory {
					category: String!
					products: [String!]!
				}

				type ValidationError @error {
					field: String!
					message: String!
				}

				type OrderPropertiesByStatusError @error {
					status: String!
					orders: [String!]!
				}
			`,
			expectedErrors: 5,  // OrderPropertiesByStatusError triggers both Error suffix and PropertiesBy violations
			expectedMsg:    "", // Will check individual messages
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create source from schema string
			source := &ast.Source{
				Name:  "test-schema.graphql",
				Input: tt.schema,
			}

			// Parse the schema
			schema, err := gqlparser.LoadSchema(source)
			if err != nil {
				t.Fatalf("Failed to parse schema: %v", err)
			}

			// Create and run the rule
			rule := NewBasicLint()
			errors := rule.Check(schema, source)

			// Check error count
			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, but got %d", tt.expectedErrors, len(errors))
				for i, e := range errors {
					t.Errorf("Error %d: %s", i+1, e.Message)
				}
				return
			}

			// Check error message if specified
			if tt.expectedErrors > 0 && tt.expectedMsg != "" {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message to contain '%s', but got:", tt.expectedMsg)
					for i, e := range errors {
						t.Errorf("Error %d: %s", i+1, e.Message)
					}
				}
			}

			// Check that all errors have the correct rule name
			for _, e := range errors {
				if e.Rule != rule.Name() {
					t.Errorf("Expected rule name '%s', but got '%s'", rule.Name(), e.Rule)
				}
			}
		})
	}
}
