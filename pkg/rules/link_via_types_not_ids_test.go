package rules

import (
	"strings"
	"testing"

	"github.com/nishant-rn/gqlparser/v2"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

func TestLinkViaTypesNotIds(t *testing.T) {
	tests := []struct {
		name          string
		schema        string
		expectedCount int
		expectedMsg   string
	}{
		{
			name: "Invalid - field uses ID instead of type",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type Product @key(fields: "id") {
					id: ID!
					name: String
					sellerId: String
				}
				
				type Seller @key(fields: "id") {
					id: ID!
					name: String
				}
			`,
			expectedCount: 1,
			expectedMsg:   "Field `Product.sellerId` should reference the `Seller` type directly instead of storing its ID",
		},
		{
			name: "Invalid - field uses ID! instead of type",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type Order @key(fields: "id") {
					id: ID!
					CustomerId: ID!
				}
				
				type Customer @key(fields: "id") {
					id: ID!
					name: String
				}
			`,
			expectedCount: 1,
			expectedMsg:   "Field `Order.CustomerId` should reference the `Customer` type directly instead of storing its ID",
		},
		{
			name: "Valid - field already uses type reference",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type Product @key(fields: "id") {
					id: ID!
					name: String
					seller: Seller
				}
				
				type Seller @key(fields: "id") {
					id: ID!
					name: String
				}
			`,
			expectedCount: 0,
		},
		{
			name: "Valid - no matching entity type",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type Product @key(fields: "id") {
					id: ID!
					name: String
					externalId: String
				}
			`,
			expectedCount: 0,
		},
		{
			name: "Valid - field doesn't end with Id",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type Product @key(fields: "id") {
					id: ID!
					name: String
					identifier: String
				}
			`,
			expectedCount: 0,
		},
		{
			name: "Invalid - multiple ID fields referencing entities",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type Order @key(fields: "id") {
					id: ID!
					customerId: String
					productId: String
				}
				
				type Customer @key(fields: "id") {
					id: ID!
					name: String
				}
				
				type Product @key(fields: "id") {
					id: ID!
					name: String
				}
			`,
			expectedCount: 2,
		},
		{
			name: "Valid - entity without @key directive should not trigger",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type Product @key(fields: "id") {
					id: ID!
					name: String
					categoryId: String
				}
				
				type Category {
					id: ID!
					name: String
				}
			`,
			expectedCount: 0,
		},
		{
			name: "Valid - ID field on entity itself is fine",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type Product @key(fields: "id") {
					id: ID!
					name: String
				}
			`,
			expectedCount: 0,
		},
		{
			name: "Invalid - field ending with ID (uppercase)",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type Product @key(fields: "id") {
					id: ID!
					name: String
					sellerID: String
				}
				
				type Seller @key(fields: "id") {
					id: ID!
					name: String
				}
			`,
			expectedCount: 1,
			expectedMsg:   "Field `Product.sellerID` should reference the `Seller` type directly instead of storing its ID",
		},
		{
			name: "Valid - Int type field ending with Id but not ID type",
			schema: `
				directive @key(fields: String!) on OBJECT
				
				type Product @key(fields: "id") {
					id: ID!
					name: String
					vendorId: Int
				}
				
				type Vendor @key(fields: "id") {
					id: ID!
					name: String
				}
			`,
			expectedCount: 1, // Int is also a scalar, so it should still trigger
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := &ast.Source{
				Name:  "test.graphql",
				Input: tt.schema,
			}

			schema, err := gqlparser.LoadSchema(source)
			if err != nil {
				t.Fatalf("Failed to parse schema: %v", err)
			}

			rule := NewLinkViaTypesNotIds()
			errors := rule.Check(schema, source)

			if len(errors) != tt.expectedCount {
				t.Errorf("Expected %d errors, got %d", tt.expectedCount, len(errors))
				for _, err := range errors {
					t.Logf("Error: %s", err.Message)
				}
			}

			if tt.expectedCount > 0 && tt.expectedMsg != "" {
				found := false
				for _, err := range errors {
					if strings.Contains(err.Message, tt.expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error message containing '%s', but didn't find it in any error", tt.expectedMsg)
					for _, err := range errors {
						t.Logf("Got: %s", err.Message)
					}
				}
			}
		})
	}
}
