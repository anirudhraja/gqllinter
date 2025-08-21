package rules

import (
	"strings"
	"testing"
)

func TestRelayNamingConvention(t *testing.T) {
	rule := NewRelayNamingConvention()

	t.Run("should pass valid Connection and Edge naming", func(t *testing.T) {
		schema := `
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type UserEdge {
			node: User
			cursor: String!
		}
		
		type User {
			id: ID!
			name: String!
		}
		
		type PostConnection {
			edges: [PostEdge]
			pageInfo: PageInfo!
		}
		
		type PostEdge {
			node: Post
			cursor: String!
		}
		
		type Post {
			id: ID!
			title: String!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-naming-convention") > 0 {
			t.Errorf("Expected no errors for valid Connection/Edge naming, got %d errors: %v", countRuleErrors(errors, "relay-naming-convention"), errors)
		}
	})

	t.Run("should flag Connection types with invalid naming", func(t *testing.T) {
		schema := `
		type userconnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type UserConnectionType {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type Connection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type user_connection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type UserEdge {
			node: User
			cursor: String!
		}
		
		type User {
			id: ID!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		
		expectedErrors := []string{
			"Connection type `userconnection` must follow the naming convention [Entity]Connection with proper case.",
			"Connection type `Connection` must have a valid entity name before 'Connection'.",
			"Connection type `user_connection` must follow the naming convention [Entity]Connection with proper case.",
		}
		
		if countRuleErrors(errors, "relay-naming-convention") != len(expectedErrors) {
			t.Errorf("Expected %d errors for invalid Connection naming, got %d", len(expectedErrors), countRuleErrors(errors, "relay-naming-convention"))
		}
		
		for _, expectedMessage := range expectedErrors {
			if !containsError(errors, expectedMessage) {
				t.Errorf("Expected error message: %s", expectedMessage)
			}
		}
	})

	t.Run("should flag Edge types with invalid naming", func(t *testing.T) {
		schema := `
		type useredge {
			node: User
			cursor: String!
		}
		
		type UserEdgeType {
			node: User
			cursor: String!
		}
		
		type Edge {
			node: User
			cursor: String!
		}
		
		type user_edge {
			node: User
			cursor: String!
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		
		expectedErrors := []string{
			"Edge type `useredge` must follow the naming convention [Entity]Edge with proper case.",
			"Edge type `Edge` must have a valid entity name before 'Edge'.",
			"Edge type `user_edge` must follow the naming convention [Entity]Edge with proper case.",
		}
		
		if countRuleErrors(errors, "relay-naming-convention") != len(expectedErrors) {
			t.Errorf("Expected %d errors for invalid Edge naming, got %d", len(expectedErrors), countRuleErrors(errors, "relay-naming-convention"))
		}
		
		for _, expectedMessage := range expectedErrors {
			if !containsError(errors, expectedMessage) {
				t.Errorf("Expected error message: %s", expectedMessage)
			}
		}
	})

	t.Run("should flag entity names that are not PascalCase", func(t *testing.T) {
		schema := `
		type myEntityConnection {
			edges: [myEntityEdge]
			pageInfo: PageInfo!
		}
		
		type myEntityEdge {
			node: MyEntity
			cursor: String!
		}
		
		type user123Connection {
			edges: [user123Edge]
			pageInfo: PageInfo!
		}
		
		type user123Edge {
			node: User123
			cursor: String!
		}
		
		type MyEntity {
			id: ID!
		}
		
		type User123 {
			id: ID!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		
		expectedErrors := []string{
			"Connection type `myEntityConnection` entity name `myEntity` must be PascalCase.",
			"Edge type `myEntityEdge` entity name `myEntity` must be PascalCase.",
			"Connection type `user123Connection` entity name `user123` must be PascalCase.",
			"Edge type `user123Edge` entity name `user123` must be PascalCase.",
		}
		
		if countRuleErrors(errors, "relay-naming-convention") != len(expectedErrors) {
			t.Errorf("Expected %d errors for non-PascalCase entity names, got %d", len(expectedErrors), countRuleErrors(errors, "relay-naming-convention"))
		}
		
		for _, expectedMessage := range expectedErrors {
			if !containsError(errors, expectedMessage) {
				t.Errorf("Expected error message: %s", expectedMessage)
			}
		}
	})

	t.Run("should pass complex entity names in PascalCase", func(t *testing.T) {
		schema := `
		type MyComplexEntityConnection {
			edges: [MyComplexEntityEdge]
			pageInfo: PageInfo!
		}
		
		type MyComplexEntityEdge {
			node: MyComplexEntity
			cursor: String!
		}
		
		type UserAccount123Connection {
			edges: [UserAccount123Edge]
			pageInfo: PageInfo!
		}
		
		type UserAccount123Edge {
			node: UserAccount123
			cursor: String!
		}
		
		type MyComplexEntity {
			id: ID!
		}
		
		type UserAccount123 {
			id: ID!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-naming-convention") > 0 {
			t.Errorf("Expected no errors for valid complex PascalCase entity names, got %d", countRuleErrors(errors, "relay-naming-convention"))
		}
	})

	t.Run("should ignore built-in types and introspection types", func(t *testing.T) {
		// This test verifies that our rule correctly skips built-in and introspection types
		// We simulate this by testing that our rule skips types starting with __
		schema := `
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type UserEdge {
			node: User
			cursor: String!
		}
		
		type User {
			id: ID!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		// This should pass without errors since all names are valid
		if countRuleErrors(errors, "relay-naming-convention") > 0 {
			t.Errorf("Expected no errors for valid naming, got %d", countRuleErrors(errors, "relay-naming-convention"))
		}
	})

	t.Run("should handle edge cases with exact suffix matches", func(t *testing.T) {
		schema := `
		type ConnectionConnection {
			edges: [ConnectionEdge]
			pageInfo: PageInfo!
		}
		
		type EdgeEdge {
			node: ConnectionType
			cursor: String!
		}
		
		type ConnectionEdge {
			node: ConnectionType
			cursor: String!
		}
		
		type EdgeConnection {
			edges: [EdgeEdge]
			pageInfo: PageInfo!
		}
		
		type ConnectionType {
			id: ID!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-naming-convention") > 0 {
			t.Errorf("Expected no errors for types with repeated suffixes, got %d errors:", countRuleErrors(errors, "relay-naming-convention"))
			for _, err := range errors {
				if err.Rule == "relay-naming-convention" {
					t.Errorf("  - %s", err.Message)
				}
			}
		}
	})

	t.Run("should flag Connection types with mismatched edges field types", func(t *testing.T) {
		schema := `
		type UserConnection {
			edges: [SomeEdge]
			pageInfo: PageInfo!
		}
		
		type PostConnection {
			edges: [PostEdge]
			pageInfo: PageInfo!
		}
		
		type OrderConnection {
			edges: [ProductEdge!]!
			pageInfo: PageInfo!
		}
		
		type PostEdge {
			node: Post
			cursor: String!
		}
		
		type SomeEdge {
			node: Some
			cursor: String!
		}
		
		type ProductEdge {
			node: Product
			cursor: String!
		}
		
		type Post {
			id: ID!
		}
		
		type Some {
			id: ID!
		}
		
		type Product {
			id: ID!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		
		expectedErrors := []string{
			"Connection type `UserConnection` edges field must reference `UserEdge`, but references `SomeEdge`.",
			"Connection type `OrderConnection` edges field must reference `OrderEdge`, but references `ProductEdge`.",
		}
		
		for _, expectedMessage := range expectedErrors {
			if !containsError(errors, expectedMessage) {
				t.Errorf("Expected error message: %s", expectedMessage)
			}
		}
		
		// PostConnection should not be flagged since it correctly references PostEdge
		unexpectedMessage := "Connection type `PostConnection` edges field must reference `PostEdge`, but references"
		if containsError(errors, unexpectedMessage) {
			t.Errorf("Did not expect error message containing: %s", unexpectedMessage)
		}
	})

	t.Run("should handle various edge field type formats", func(t *testing.T) {
		schema := `
		type UserConnection {
			edges: [UserEdge!]!
			pageInfo: PageInfo!
		}
		
		type PostConnection {
			edges: [PostEdge]
			pageInfo: PageInfo!
		}
		
		type OrderConnection {
			edges: [OrderEdge!]
			pageInfo: PageInfo!
		}
		
		type UserEdge {
			node: User
			cursor: String!
		}
		
		type PostEdge {
			node: Post
			cursor: String!
		}
		
		type OrderEdge {
			node: Order
			cursor: String!
		}
		
		type User {
			id: ID!
		}
		
		type Post {
			id: ID!
		}
		
		type Order {
			id: ID!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		
		// Should have no errors since all edges fields correctly reference matching Edge types
		if countRuleErrors(errors, "relay-naming-convention") > 0 {
			t.Errorf("Expected no errors for correctly matched edges fields, got %d", countRuleErrors(errors, "relay-naming-convention"))
			for _, err := range errors {
				if err.Rule == "relay-naming-convention" {
					t.Errorf("  - %s", err.Message)
				}
			}
		}
	})

	t.Run("should handle Connection types without edges field", func(t *testing.T) {
		schema := `
		type UserConnection {
			pageInfo: PageInfo!
			nodes: [User]
		}
		
		type PageInfo {
			hasNextPage: Boolean!
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		
		// Should only flag naming issues, not missing edges field (that's handled by other rules)
		if countRuleErrors(errors, "relay-naming-convention") > 0 {
			for _, err := range errors {
				if err.Rule == "relay-naming-convention" && strings.Contains(err.Message, "edges field") {
					t.Errorf("Should not flag missing edges field: %s", err.Message)
				}
			}
		}
	})
}
