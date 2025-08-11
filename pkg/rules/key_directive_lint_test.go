package rules

import (
	"strings"
	"testing"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

func TestKeyDirectivesLint(t *testing.T) {
	rule := NewKeyDirectivesLint()

	tests := []struct {
		name           string
		schema         string
		expectedErrors int
		expectedMsg    string
	}{
		{
			name: "Valid: @key with existing field",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "id") {
					id: ID!
					name: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: @key with multiple existing fields",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "id name") {
					id: ID!
					name: String
					email: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: @key with non-existing field",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "nonExistentField") {
					id: ID!
					name: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Field 'nonExistentField' specified in @key directive does not exist in object type 'User'",
		},
		{
			name: "Invalid: @key with mix of primitive and non-primitive fields",
			schema: `
				directive @key(fields: String!) on OBJECT

				type Product @key(fields: "id version { code }") {
				  id: ID!
				  version: Version!
				  name: String
				}
				
				type Version {
				  code: String!
				  releaseDate: String
				}
			`,
			expectedErrors: 1, // 'version' is object type
			expectedMsg:    "Field 'version' specified in @key directive must be a primitive or scalar type, but is of type 'Version'",
		},
		{
			name: "Invalid: Multiple non-existing fields in @key",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "badField1 badField2") {
					id: ID!
					name: String
				}
			`,
			expectedErrors: 2,
		},
		{
			name: "Invalid: @key with nested field selection (object type not allowed)",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "profile { id }") {
					id: ID!
					profile: UserProfile
				}
				
				type UserProfile {
					id: ID!
					bio: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Field 'profile' specified in @key directive must be a primitive or scalar type, but is of type 'UserProfile'",
		},
		{
			name: "Valid: Multiple @key directives on different types",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "id") {
					id: ID!
					name: String
				}
				
				type Product @key(fields: "sku") {
					sku: String!
					title: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: Multiple @key directives with errors",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "badField") {
					id: ID!
					name: String
				}
				
				type Product @key(fields: "anotherBadField") {
					sku: String!
					title: String
				}
			`,
			expectedErrors: 2,
		},
		{
			name: "Valid: Object without @key directive",
			schema: `
				type User {
					id: ID!
					name: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: @key on interface (should be ignored)",
			schema: `
				directive @key(fields: String!) on OBJECT | INTERFACE
				
				interface Node @key(fields: "id") {
					id: ID!
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: @key with comma-separated fields",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "id,name") {
					id: ID!
					name: String
					email: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: Empty @key fields (edge case)",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "") {
					id: ID!
					name: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: @key field with object type",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "profile") {
					id: ID!
					profile: UserProfile
				}
				
				type UserProfile {
					bio: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Field 'profile' specified in @key directive must be a primitive or scalar type, but is of type 'UserProfile'",
		},
		{
			name: "Invalid: @key field with enum type",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "status") {
					id: ID!
					status: UserStatus
				}
				
				enum UserStatus {
					ACTIVE
					INACTIVE
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Field 'status' specified in @key directive must be a primitive or scalar type, but is of type 'UserStatus'",
		},
		{
			name: "Invalid: @key field with list type",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "tags") {
					id: ID!
					tags: [String!]!
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Field 'tags' specified in @key directive must be a primitive or scalar type",
		},
		{
			name: "Valid: @key with custom scalar",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				scalar DateTime
				
				type User @key(fields: "createdAt") {
					id: ID!
					createdAt: DateTime
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: @key with all built-in scalar types",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "id name age score active") {
					id: ID!
					name: String!
					age: Int
					score: Float
					active: Boolean
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: @key field with interface type",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "node") {
					id: ID!
					node: Node
				}
				
				interface Node {
					id: ID!
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Field 'node' specified in @key directive must be a primitive or scalar type, but is of type 'Node'",
		},
		{
			name: "Invalid: @key field with union type",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "result") {
					id: ID!
					result: SearchResult
				}
				
				union SearchResult = User | Product
				
				type Product {
					id: ID!
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Field 'result' specified in @key directive must be a primitive or scalar type, but is of type 'SearchResult'",
		},
		{
			name: "Valid: @key with NonNull scalar type",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "email") {
					id: ID!
					email: String!
				}
			`,
			expectedErrors: 0,
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

			// Run the rule
			errors := rule.Check(schema, source)

			// Check the number of errors
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

			// Verify rule name in errors
			for _, err := range errors {
				if err.Rule != rule.Name() {
					t.Errorf("Expected rule name '%s', got '%s'", rule.Name(), err.Rule)
				}
			}
		})
	}
}
