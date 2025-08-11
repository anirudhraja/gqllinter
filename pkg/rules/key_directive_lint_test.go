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
			name: "Invalid: @key with comma-separated fields (multiple commas)",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "id, name, email") {
					id: ID!
					name: String
					email: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "@key directive fields must be space-separated, not comma-separated. Found comma in fields: 'id, name, email' for object 'User'",
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
			name: "Valid: Empty @key fields (edge case)",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type User @key(fields: "") {
					id: ID!
					name: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Missing or invalid 'fields' argument in @key directive for object 'User'",
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
		{
			name: "Invalid: Single @key directive with resolvable: false missing fields",
			schema: `
				directive @key(fields: String!, resolvable: Boolean = true) on OBJECT
				
				type User @key(fields: "id", resolvable: false) {
					id: ID!
					name: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Object type 'User' has a single @key directive with 'resolvable: false' but the key does not include all object fields. Missing fields in @key: [name]. All fields must be included when using 'resolvable: false'.",
		},
		{
			name: "Valid: Multiple @key directives without resolvable: false",
			schema: `
				directive @key(fields: String!, resolvable: Boolean = true) on OBJECT
				
				type User @key(fields: "id") @key(fields: "email") {
					id: ID!
					email: String
					name: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: Multiple @key directives with resolvable: true",
			schema: `
				directive @key(fields: String!, resolvable: Boolean = true) on OBJECT
				
				type User @key(fields: "id", resolvable: true) @key(fields: "email", resolvable: true) {
					id: ID!
					email: String
					name: String
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Invalid: Multiple @key directives with one having resolvable: false",
			schema: `
				directive @key(fields: String!, resolvable: Boolean = true) on OBJECT
				
				type User @key(fields: "id", resolvable: false) @key(fields: "email") {
					id: ID!
					email: String
					name: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Object type 'User' has a @key directive with 'resolvable: false' but also has 2 total @key directives",
		},
		{
			name: "Invalid: Multiple @key directives with all having resolvable: false",
			schema: `
				directive @key(fields: String!, resolvable: Boolean = true) on OBJECT
				
				type User @key(fields: "id", resolvable: false) @key(fields: "email", resolvable: false) {
					id: ID!
					email: String
					name: String
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Object type 'User' has a @key directive with 'resolvable: false' but also has 2 total @key directives",
		},
		{
			name: "Invalid: Different objects with different @key configurations",
			schema: `
				directive @key(fields: String!, resolvable: Boolean = true) on OBJECT
				
				type User @key(fields: "id", resolvable: false) {
					id: ID!
					name: String
				}
				
				type Product @key(fields: "sku") @key(fields: "title") {
					sku: String!
					title: String!
				}
			`,
			expectedErrors: 1, // User missing name field in resolvable: false key
			expectedMsg:    "Object type 'User' has a single @key directive with 'resolvable: false' but the key does not include all object fields",
		},
		{
			name: "Invalid: Mixed valid and invalid objects",
			schema: `
				directive @key(fields: String!, resolvable: Boolean = true) on OBJECT
				
				type User @key(fields: "id", resolvable: false) {
					id: ID!
					name: String
				}
				
				type Product @key(fields: "sku", resolvable: false) @key(fields: "title") {
					sku: String!
					title: String!
				}
			`,
			expectedErrors: 2, // User missing name field + Product multiple @key with resolvable: false
		},
		{
			name: "Valid: Single @key with resolvable: false includes all fields",
			schema: `
				directive @key(fields: String!, resolvable: Boolean = true) on OBJECT
				
				type User @key(fields: "id name email", resolvable: false) {
					id: ID!
					name: String!
					email: String!
				}
			`,
			expectedErrors: 0,
		},
		{
			name: "Valid: Single @key with resolvable: false and nested field selection (only top-level fields matter)",
			schema: `
				directive @key(fields: String!, resolvable: Boolean = true) on OBJECT
				
				type User @key(fields: "id profile { nested }", resolvable: false) {
					id: ID!
					profile: Profile
				}
				
				type Profile {
					nested: String!
					other: String!
				}
			`,
			expectedErrors: 1,
			expectedMsg:    "Field 'profile' specified in @key directive must be a primitive or scalar type, but is of type 'Profile'",
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
