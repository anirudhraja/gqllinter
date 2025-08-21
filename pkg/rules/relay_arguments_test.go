package rules

import (
	"testing"
)

func TestRelayArguments(t *testing.T) {
	rule := NewRelayArguments()

	t.Run("should pass valid forward pagination arguments", func(t *testing.T) {
		schema := `
		type Query {
			users(first: Int, after: String): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-arguments") > 0 {
			t.Errorf("Expected no errors for valid forward pagination arguments, got %d errors: %v", countRuleErrors(errors, "relay-arguments"), errors)
		}
	})

	t.Run("should pass valid backward pagination arguments", func(t *testing.T) {
		schema := `
		type Query {
			users(last: Int, before: String): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-arguments") > 0 {
			t.Errorf("Expected no errors for valid backward pagination arguments, got %d errors: %v", countRuleErrors(errors, "relay-arguments"), errors)
		}
	})

	t.Run("should pass valid both forward and backward pagination arguments", func(t *testing.T) {
		schema := `
		type Query {
			users(first: Int, after: String, last: Int, before: String): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-arguments") > 0 {
			t.Errorf("Expected no errors for valid forward and backward pagination arguments, got %d errors: %v", countRuleErrors(errors, "relay-arguments"), errors)
		}
	})

	t.Run("should pass with NonNull argument types", func(t *testing.T) {
		schema := `
		type Query {
			users(first: Int!, after: String!): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-arguments") > 0 {
			t.Errorf("Expected no errors for NonNull argument types, got %d errors: %v", countRuleErrors(errors, "relay-arguments"), errors)
		}
	})

	t.Run("should pass with explicit Cursor type", func(t *testing.T) {
		schema := `
		scalar Cursor
		
		type Query {
			users(first: Int, after: Cursor): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-arguments") > 0 {
			t.Errorf("Expected no errors for explicit Cursor type, got %d errors: %v", countRuleErrors(errors, "relay-arguments"), errors)
		}
	})

	t.Run("should flag Connection field with no pagination arguments", func(t *testing.T) {
		schema := `
		type Query {
			users: UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `Query.users` returns Connection type but lacks proper pagination arguments. Must include forward pagination arguments (first and after), backward pagination arguments (last and before), or both."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for missing pagination arguments, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag incomplete forward pagination (missing after)", func(t *testing.T) {
		schema := `
		type Query {
			users(first: Int): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `Query.users` has `first` argument but is missing `after` argument for complete forward pagination."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for incomplete forward pagination, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag incomplete forward pagination (missing first)", func(t *testing.T) {
		schema := `
		type Query {
			users(after: String): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `Query.users` has `after` argument but is missing `first` argument for complete forward pagination."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for incomplete forward pagination, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag incomplete backward pagination (missing before)", func(t *testing.T) {
		schema := `
		type Query {
			users(last: Int): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `Query.users` has `last` argument but is missing `before` argument for complete backward pagination."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for incomplete backward pagination, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag incomplete backward pagination (missing last)", func(t *testing.T) {
		schema := `
		type Query {
			users(before: String): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `Query.users` has `before` argument but is missing `last` argument for complete backward pagination."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for incomplete backward pagination, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag invalid first argument type", func(t *testing.T) {
		schema := `
		type Query {
			users(first: String, after: String): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `Query.users` argument `first` must be a non-negative integer type (Int), but is String."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for invalid first argument type, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag invalid after argument type", func(t *testing.T) {
		schema := `
		type Query {
			users(first: Int, after: Int): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `Query.users` argument `after` must be a Cursor type (String), but is Int."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for invalid after argument type, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag invalid last argument type", func(t *testing.T) {
		schema := `
		type Query {
			users(last: String, before: String): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `Query.users` argument `last` must be a non-negative integer type (Int), but is String."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for invalid last argument type, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag invalid before argument type", func(t *testing.T) {
		schema := `
		type Query {
			users(last: Int, before: Int): UserConnection
		}
		
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
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `Query.users` argument `before` must be a Cursor type (String), but is Int."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for invalid before argument type, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should check Connection fields in object types", func(t *testing.T) {
		schema := `
		type Query {
			user(id: ID!): User
		}
		
		type User {
			id: ID!
			friends: UserConnection
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type UserEdge {
			node: User
			cursor: String!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `User.friends` returns Connection type but lacks proper pagination arguments. Must include forward pagination arguments (first and after), backward pagination arguments (last and before), or both."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for Connection field in object type, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should check Connection fields in interface types", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
			connections: UserConnection
		}
		
		type User implements Node {
			id: ID!
			connections: UserConnection
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type UserEdge {
			node: User
			cursor: String!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		
		// Both the interface and the implementing type should be flagged
		if countRuleErrors(errors, "relay-arguments") != 2 {
			t.Errorf("Expected exactly 2 errors for Connection fields in interface and implementing type, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		expectedMessage1 := "Field `Node.connections` returns Connection type but lacks proper pagination arguments. Must include forward pagination arguments (first and after), backward pagination arguments (last and before), or both."
		expectedMessage2 := "Field `User.connections` returns Connection type but lacks proper pagination arguments. Must include forward pagination arguments (first and after), backward pagination arguments (last and before), or both."
		
		if !containsError(errors, expectedMessage1) {
			t.Errorf("Expected error message: %s", expectedMessage1)
		}
		
		if !containsError(errors, expectedMessage2) {
			t.Errorf("Expected error message: %s", expectedMessage2)
		}
	})

	t.Run("should handle nested types that return Connection types", func(t *testing.T) {
		schema := `
		type Query {
			users(first: Int, after: String): UserConnection
		}
		
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
			posts: PostConnection
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
			title: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		expectedMessage := "Field `User.posts` returns Connection type but lacks proper pagination arguments. Must include forward pagination arguments (first and after), backward pagination arguments (last and before), or both."
		
		if countRuleErrors(errors, "relay-arguments") != 1 {
			t.Errorf("Expected exactly 1 error for nested Connection field, got %d", countRuleErrors(errors, "relay-arguments"))
		}
		
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should ignore non-Connection fields", func(t *testing.T) {
		schema := `
		type Query {
			user(id: ID!): User
			users: [User]
		}
		
		type User {
			id: ID!
			name: String
			friends: [User]
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-arguments") > 0 {
			t.Errorf("Expected no errors for non-Connection fields, got %d errors: %v", countRuleErrors(errors, "relay-arguments"), errors)
		}
	})
}
