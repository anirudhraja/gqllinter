package rules

import (
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestNoInvalidEnum(t *testing.T) {
	rule := NewNoInvalidEnum()

	tests := []struct {
		name           string
		schema         string
		expectedErrors int
		expectedMsg    string
	}{
		{
			name: "Valid: Enum without INVALID value",
			schema: `
				enum Status {
					ACTIVE
					INACTIVE
					PENDING
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: Multiple enums without INVALID values",
			schema: `
				enum Status {
					ACTIVE
					INACTIVE
					PENDING
				}
				
				enum Priority {
					HIGH
					MEDIUM
					LOW
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: Enum with INVALID value (uppercase)",
			schema: `
				enum Status {
					ACTIVE
					INVALID
					PENDING
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Enum value 'INVALID' is not allowed. Use a different name for enum 'Status'",
		},
		{
			name: "Invalid: Enum with invalid value (lowercase)",
			schema: `
				enum Status {
					ACTIVE
					invalid
					PENDING
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Enum value 'INVALID' is not allowed. Use a different name for enum 'Status'",
		},
		{
			name: "Invalid: Enum with Invalid value (mixed case)",
			schema: `
				enum Status {
					ACTIVE
					Invalid
					PENDING
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Enum value 'INVALID' is not allowed. Use a different name for enum 'Status'",
		},
		{
			name: "Invalid: Multiple enums with INVALID values",
			schema: `
				enum Status {
					ACTIVE
					INVALID
					PENDING
				}
				
				enum Priority {
					HIGH
					INVALID
					LOW
				}
			`,
			expectedErrors: 2,
		},
		{
			name: "Valid: Enum with INVALID-like but different values",
			schema: `
				enum Status {
					ACTIVE
					INVALID_STATE
					NOT_INVALID
					VALIDATION_ERROR
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: Multiple INVALID values in same enum",
			schema: `
				enum Status {
					ACTIVE
					INVALID
					PENDING
					invalid
				}
			`,
			expectedErrors: 2, // Both INVALID and invalid should be caught
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
			schema, gqlErr := gqlparser.LoadSchema(source)
			if gqlErr != nil {
				t.Fatalf("Failed to parse schema: %v", gqlErr)
			}

			// Run the rule
			errors := rule.Check(schema, source)

			// Check error count
			if len(errors) != tt.expectedErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectedErrors, len(errors))
				for i, err := range errors {
					t.Logf("Error %d: %s", i+1, err.Message)
				}
				return
			}

			// Check error message if expected
			if tt.expectedErrors > 0 && tt.expectedMsg != "" {
				found := false
				for _, err := range errors {
					if strings.Contains(err.Message, tt.expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message to contain '%s', but got:", tt.expectedMsg)
					for i, err := range errors {
						t.Logf("Error %d: %s", i+1, err.Message)
					}
				}
			}

			// Verify all errors have correct rule name
			for _, err := range errors {
				if err.Rule != rule.Name() {
					t.Errorf("Expected rule name '%s', got '%s'", rule.Name(), err.Rule)
				}
			}
		})
	}
}
