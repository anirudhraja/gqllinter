package rules

import (
	"fmt"
	"strings"
	"testing"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/nishant-rn/gqlparser/v2"
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// Helper function to parse GraphQL schema from string
func parseSchema(t *testing.T, schemaStr string) (*ast.Schema, *ast.Source) {
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

// Helper function to run a rule and return errors
func runRule(t *testing.T, rule types.Rule, schemaStr string) []types.LintError {
	schema, source := parseSchema(t, schemaStr)
	return rule.Check(schema, source)
}

// Helper function to check if an error contains expected message
func containsError(errors []types.LintError, expectedMessage string) bool {
	for _, err := range errors {
		if err.Message == expectedMessage {
			return true
		}
	}
	return false
}

// Helper function to count errors for a specific rule
func countRuleErrors(errors []types.LintError, ruleName string) int {
	count := 0
	for _, err := range errors {
		if err.Rule == ruleName {
			count++
		}
	}
	return count
}

func TestTypesHaveDescriptions(t *testing.T) {
	rule := NewTypesHaveDescriptions()

	t.Run("should flag types without descriptions", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if len(errors) == 0 {
			t.Error("Expected error for type without description")
		}
		if !containsError(errors, "The object type `User` is missing a description.") {
			t.Error("Expected specific error message about missing description")
		}
	})

	t.Run("should pass types with descriptions", func(t *testing.T) {
		schema := `
		"""A user in the system"""
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "types-have-descriptions") > 0 {
			t.Error("Expected no errors for type with description")
		}
	})
}

func TestFieldsHaveDescriptions(t *testing.T) {
	rule := NewFieldsHaveDescriptions()

	t.Run("should flag fields without descriptions", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		userFieldErrors := 0
		for _, err := range errors {
			if err.Message == "The field `User.id` is missing a description." ||
				err.Message == "The field `User.name` is missing a description." {
				userFieldErrors++
			}
		}
		if userFieldErrors < 2 {
			t.Errorf("Expected at least 2 errors for fields without descriptions, got %d", userFieldErrors)
		}
	})

	t.Run("should pass fields with descriptions", func(t *testing.T) {
		schema := `
		type User {
			"""User identifier"""
			id: ID!
			"""User's full name"""
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		userFieldErrors := 0
		for _, err := range errors {
			if err.Message == "The field `User.id` is missing a description." ||
				err.Message == "The field `User.name` is missing a description." {
				userFieldErrors++
			}
		}
		if userFieldErrors > 0 {
			t.Errorf("Expected no errors for User fields with descriptions, got %d", userFieldErrors)
		}
	})
}

func TestNoHashtagDescription(t *testing.T) {
	rule := NewNoHashtagDescription()

	t.Run("should flag hashtag comments", func(t *testing.T) {
		schema := `
		# This is a user type
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-hashtag-description") == 0 {
			t.Error("Expected error for hashtag comment")
		}
	})

	t.Run("should pass triple quote descriptions", func(t *testing.T) {
		schema := `
		"""This is a user type"""
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-hashtag-description") > 0 {
			t.Error("Expected no hashtag errors for triple quote description")
		}
	})
}

func TestNamingConvention(t *testing.T) {
	rule := NewNamingConvention()

	t.Run("Type Names - PascalCase Validation", func(t *testing.T) {
		t.Run("should flag non-PascalCase type names", func(t *testing.T) {
			schema := `
			type user_data {
				id: ID!
			}
			type userProfile {
				id: ID!
			}
			type USER_DATA {
				id: ID!
			}
			`
			errors := runRule(t, rule, schema)

			expectedErrors := []string{
				"Type name `user_data` should be PascalCase.",
				"Type name `userProfile` should be PascalCase.",
				"Type name `USER_DATA` should be PascalCase.",
			}

			for _, expectedMsg := range expectedErrors {
				if !containsError(errors, expectedMsg) {
					t.Errorf("Expected error: %s", expectedMsg)
				}
			}
		})

		t.Run("should pass valid PascalCase type names", func(t *testing.T) {
			schema := `
			type UserProfile {
				id: ID!
			}
			type AccountSettings {
				id: ID!
			}
			type PaymentMethod123 {
				id: ID!
			}
			`
			errors := runRule(t, rule, schema)
			pascalCaseErrors := 0
			for _, err := range errors {
				if strings.Contains(err.Message, "should be PascalCase") {
					pascalCaseErrors++
				}
			}
			if pascalCaseErrors > 0 {
				t.Errorf("Expected no PascalCase errors, got %d", pascalCaseErrors)
			}
		})
	})

	t.Run("Object Types - Suffix/Prefix Validation", func(t *testing.T) {
		t.Run("should flag Object types with Type/Object suffix", func(t *testing.T) {
			schema := `
			type UserType {
				id: ID!
			}
			type DataObject {
				id: ID!
			}
			type TypeUser {
				id: ID!
			}
			type ObjectData {
				id: ID!
			}
			`
			errors := runRule(t, rule, schema)

			expectedErrors := []string{
				"Type name `UserType` should be PascalCase and should not start/end with `Type` or `Object`",
				"Type name `DataObject` should be PascalCase and should not start/end with `Type` or `Object`",
				"Type name `TypeUser` should be PascalCase and should not start/end with `Type` or `Object`",
				"Type name `ObjectData` should be PascalCase and should not start/end with `Type` or `Object`",
			}

			for _, expectedMsg := range expectedErrors {
				if !containsError(errors, expectedMsg) {
					t.Errorf("Expected error: %s", expectedMsg)
				}
			}
		})

		t.Run("should pass Object types without Type/Object suffix", func(t *testing.T) {
			schema := `
			type User {
				id: ID!
			}
			type Account {
				id: ID!
			}
			type PaymentMethod {
				id: ID!
			}
			`
			errors := runRule(t, rule, schema)
			objectSuffixErrors := 0
			for _, err := range errors {
				if strings.Contains(err.Message, "should not start/end with `Type` or `Object`") {
					objectSuffixErrors++
				}
			}
			if objectSuffixErrors > 0 {
				t.Errorf("Expected no Object suffix errors, got %d", objectSuffixErrors)
			}
		})
	})

	t.Run("Interface Types - Suffix/Prefix Validation", func(t *testing.T) {
		t.Run("should flag Interface types with Interface suffix", func(t *testing.T) {
			schema := `
			interface NodeInterface {
				id: ID!
			}
			interface InterfaceNode {
				id: ID!
			}
			`
			errors := runRule(t, rule, schema)

			expectedErrors := []string{
				"Interface name `NodeInterface` should be PascalCase and should not start/end with `Interface`",
				"Interface name `InterfaceNode` should be PascalCase and should not start/end with `Interface`",
			}

			for _, expectedMsg := range expectedErrors {
				if !containsError(errors, expectedMsg) {
					t.Errorf("Expected error: %s", expectedMsg)
				}
			}
		})

		t.Run("should pass Interface types without Interface suffix", func(t *testing.T) {
			schema := `
			interface Node {
				id: ID!
			}
			interface Timestamped {
				createdAt: String!
			}
			`
			errors := runRule(t, rule, schema)
			interfaceSuffixErrors := 0
			for _, err := range errors {
				if strings.Contains(err.Message, "should not start/end with `Interface`") {
					interfaceSuffixErrors++
				}
			}
			if interfaceSuffixErrors > 0 {
				t.Errorf("Expected no Interface suffix errors, got %d", interfaceSuffixErrors)
			}
		})
	})

	t.Run("Enum Types - Suffix/Prefix Validation", func(t *testing.T) {
		t.Run("should flag Enum types with Enum suffix", func(t *testing.T) {
			schema := `
			enum StatusEnum {
				ACTIVE
				INACTIVE
			}
			enum EnumStatus {
				ACTIVE
				INACTIVE
			}
			enum statuses_enum {
				ACTIVE
				INACTIVE
			}
			`
			errors := runRule(t, rule, schema)

			// Should have errors for enum naming
			enumNamingErrors := 0
			for _, err := range errors {
				if strings.Contains(err.Message, "should not start or end with `Enum`") {
					enumNamingErrors++
				}
			}
			if enumNamingErrors < 3 {
				t.Errorf("Expected at least 3 enum naming errors, got %d", enumNamingErrors)
			}
		})

		t.Run("should pass Enum types without Enum suffix", func(t *testing.T) {
			schema := `
			enum Status {
				ACTIVE
				INACTIVE
			}
			enum UserRole {
				ADMIN
				USER
			}
			`
			errors := runRule(t, rule, schema)
			enumSuffixErrors := 0
			for _, err := range errors {
				if strings.Contains(err.Message, "should not start or end with `Enum`") {
					enumSuffixErrors++
				}
			}
			if enumSuffixErrors > 0 {
				t.Errorf("Expected no Enum suffix errors, got %d", enumSuffixErrors)
			}
		})
	})

	t.Run("Enum Values - UPPER_CASE Validation", func(t *testing.T) {
		t.Run("should flag non-UPPER_CASE enum values", func(t *testing.T) {
			schema := `
			enum Status {
				active
				Inactive
				PENDING
				in_progress
				Done
			}
			`
			errors := runRule(t, rule, schema)

			expectedErrors := []string{
				"Enum value `Status.active` should be UPPER_CASE",
				"Enum value `Status.Inactive` should be UPPER_CASE",
				"Enum value `Status.in_progress` should be UPPER_CASE",
				"Enum value `Status.Done` should be UPPER_CASE",
			}

			for _, expectedMsg := range expectedErrors {
				if !containsError(errors, expectedMsg) {
					t.Errorf("Expected error: %s", expectedMsg)
				}
			}
		})

		t.Run("should pass valid UPPER_CASE enum values", func(t *testing.T) {
			schema := `
			enum Status {
				ACTIVE
				INACTIVE
				PENDING
				IN_PROGRESS
				DONE
				STATUS_123
			}
			`
			errors := runRule(t, rule, schema)
			upperCaseErrors := 0
			for _, err := range errors {
				if strings.Contains(err.Message, "should be UPPER_CASE") {
					upperCaseErrors++
				}
			}
			if upperCaseErrors > 0 {
				t.Errorf("Expected no UPPER_CASE errors, got %d", upperCaseErrors)
			}
		})
	})

	t.Run("Field Names - camelCase Validation", func(t *testing.T) {
		t.Run("should flag non-camelCase field names", func(t *testing.T) {
			schema := `
			type User {
				User_id: ID!
				display_name: String!
				FirstName: String!
				LAST_NAME: String!
			}
			`
			errors := runRule(t, rule, schema)

			expectedErrors := []string{
				"Field name `User.User_id` should be camelCase.",
				"Field name `User.display_name` should be camelCase.",
				"Field name `User.FirstName` should be camelCase.",
				"Field name `User.LAST_NAME` should be camelCase.",
			}

			for _, expectedMsg := range expectedErrors {
				if !containsError(errors, expectedMsg) {
					t.Errorf("Expected error: %s", expectedMsg)
				}
			}
		})

		t.Run("should pass valid camelCase field names", func(t *testing.T) {
			schema := `
			type User {
				id: ID!
				displayName: String!
				firstName: String!
				lastName: String!
				createdAt: String!
				field123: String!
			}
			`
			errors := runRule(t, rule, schema)
			camelCaseErrors := 0
			for _, err := range errors {
				if strings.Contains(err.Message, "should be camelCase") {
					camelCaseErrors++
				}
			}
			if camelCaseErrors > 0 {
				t.Errorf("Expected no camelCase errors, got %d", camelCaseErrors)
			}
		})
	})

	t.Run("Edge Cases", func(t *testing.T) {
		t.Run("should handle empty names gracefully", func(t *testing.T) {
			// This test mainly ensures the functions don't crash on edge cases
			// The GraphQL parser should prevent truly empty names, but we test our functions
			rule := NewNamingConvention()

			// Test empty string handling in helper functions
			if rule.isPascalCase("") {
				t.Error("Empty string should not be valid PascalCase")
			}
			if rule.isCamelCase("") {
				t.Error("Empty string should not be valid camelCase")
			}
			if rule.isUpperCase("") {
				t.Error("Empty string should not be valid UPPER_CASE")
			}
		})

		t.Run("should handle numbers in names correctly", func(t *testing.T) {
			schema := `
			type User123 {
				field456: String!
			}
			enum Status789 {
				VALUE_123
			}
			`
			errors := runRule(t, rule, schema)

			// These should be valid
			invalidErrors := 0
			for _, err := range errors {
				if strings.Contains(err.Message, "User123") ||
					strings.Contains(err.Message, "field456") ||
					strings.Contains(err.Message, "Status789") ||
					strings.Contains(err.Message, "VALUE_123") {
					invalidErrors++
				}
			}
			if invalidErrors > 0 {
				t.Errorf("Numbers in names should be allowed, got %d errors", invalidErrors)
			}
		})

		t.Run("should skip built-in types", func(t *testing.T) {
			// Built-in types should be skipped automatically by the GraphQL parser
			// This test ensures our rule doesn't process them
			schema := `
			type User {
				id: ID!
				name: String!
			}
			`
			errors := runRule(t, rule, schema)

			// Should not have errors for built-in types like String, ID, etc.
			for _, err := range errors {
				if strings.Contains(err.Message, "String") ||
					strings.Contains(err.Message, "ID") ||
					strings.Contains(err.Message, "Int") ||
					strings.Contains(err.Message, "Float") ||
					strings.Contains(err.Message, "Boolean") {
					t.Errorf("Should not validate built-in types: %s", err.Message)
				}
			}
		})
	})

	t.Run("Complex Schema Integration", func(t *testing.T) {
		t.Run("should handle multiple violations in one schema", func(t *testing.T) {
			schema := `
			type user_type {
				User_id: ID!
				display_name: String!
			}
			
			interface NodeInterface {
				id: ID!
			}
			
			enum statusEnum {
				active
				inactive
			}
			`
			errors := runRule(t, rule, schema)

			// Should have multiple different types of errors
			expectedMinErrors := 6 // At least 6 different violations
			if len(errors) < expectedMinErrors {
				t.Errorf("Expected at least %d errors for complex schema, got %d", expectedMinErrors, len(errors))
			}
		})

		t.Run("should pass fully compliant schema", func(t *testing.T) {
			schema := `
			type User {
				id: ID!
				displayName: String!
				email: String!
			}
			
			interface Node {
				id: ID!
			}
			
			enum UserStatus {
				ACTIVE
				INACTIVE
				PENDING
			}
			
			input UserInput {
				displayName: String!
				email: String!
			}
			`
			errors := runRule(t, rule, schema)

			namingErrors := countRuleErrors(errors, "naming-convention")
			if namingErrors > 0 {
				t.Errorf("Expected no naming convention errors for compliant schema, got %d", namingErrors)
				for _, err := range errors {
					if err.Rule == "naming-convention" {
						t.Logf("Error: %s", err.Message)
					}
				}
			}
		})
	})
}

func TestNoUnusedFields(t *testing.T) {
	rule := NewNoUnusedFields()

	t.Run("should flag unused fields", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			unusedField: String
		}
		`
		errors := runRule(t, rule, schema)
		unusedFieldErrors := 0
		for _, err := range errors {
			if err.Rule == "no-unused-fields" && err.Message == "Field `User.unusedField` is never used and can be removed." {
				unusedFieldErrors++
			}
		}
		if unusedFieldErrors == 0 {
			t.Error("Expected error for unused field")
		}
	})
}

func TestRequireDeprecationReason(t *testing.T) {
	rule := NewRequireDeprecationReason()

	t.Run("should flag deprecated fields without reason", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			oldField: String @deprecated
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "require-deprecation-reason") == 0 {
			t.Error("Expected error for deprecated field without reason")
		}
	})

	t.Run("should pass deprecated fields with reason", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			oldField: String @deprecated(reason: "Use newField instead")
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "require-deprecation-reason") > 0 {
			t.Error("Expected no deprecation errors for field with reason")
		}
	})
}

func TestNoScalarResultTypeOnMutation(t *testing.T) {
	rule := NewNoScalarResultTypeOnMutation()

	t.Run("should flag mutations returning scalars", func(t *testing.T) {
		schema := `
		type Mutation {
			createUser: Boolean
			deleteUser: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-scalar-result-type-on-mutation") < 2 {
			t.Error("Expected at least 2 errors for scalar mutations")
		}
	})

	t.Run("should pass mutations returning objects", func(t *testing.T) {
		schema := `
		type Mutation {
			createUser: CreateUserResult
		}
		
		type CreateUserResult {
			success: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-scalar-result-type-on-mutation") > 0 {
			t.Error("Expected no scalar errors for object mutations")
		}
	})
}

func TestAlphabetize(t *testing.T) {
	rule := NewAlphabetize()

	t.Run("should flag unordered fields", func(t *testing.T) {
		schema := `
		type User {
			name: String!
			id: ID!
			email: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "alphabetize") == 0 {
			t.Error("Expected error for unordered fields")
		}
	})

	t.Run("should pass ordered fields", func(t *testing.T) {
		schema := `
		type User {
			email: String!
			id: ID!
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "alphabetize") > 0 {
			t.Error("Expected no alphabetize errors for ordered fields")
		}
	})
}

func TestInputName(t *testing.T) {
	rule := NewInputName()

	t.Run("should flag input types with Request suffix", func(t *testing.T) {
		schema := `
		type Query {
			getUser(request: GetUserRequest!): User
			listUsers(request: ListUsersRequest!): [User!]!
		}
		
		type Mutation {
			createUser(request: CreateUserRequest!): User!
			updateUser(request: UpdateUserRequest!): User!
		}
		
		type User {
			id: ID!
			name: String!
		}
		
		input GetUserRequest {
			id: ID!
		}
		
		input ListUsersRequest {
			limit: Int
			offset: Int
		}
		
		input CreateUserRequest {
			name: String!
			email: String!
		}
		
		input UpdateUserRequest {
			id: ID!
			name: String
			email: String
		}
		`
		errors := runRule(t, rule, schema)
		errorCount := countRuleErrors(errors, "operation-input-name")
		if errorCount != 4 {
			t.Errorf("Expected 4 errors for forbidden Request suffix types, got %d", errorCount)
		}
	})

	t.Run("should flag versioned Request input types", func(t *testing.T) {
		schema := `
		type Query {
			getUser(request: GetUserRequestV2!): User
			listUsers(request: ListUsersRequestVersion3!): [User!]!
		}
		
		type Mutation {
			createUser(request: CreateUserRequestV1!): User!
		}
		
		type User {
			id: ID!
		}
		
		input GetUserRequestV2 {
			id: ID!
			includeProfile: Boolean
		}
		
		input ListUsersRequestVersion3 {
			limit: Int
			offset: Int
			filter: String
		}
		
		input CreateUserRequestV1 {
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		errorCount := countRuleErrors(errors, "operation-input-name")
		if errorCount != 3 {
			t.Errorf("Expected 3 errors for forbidden versioned Request suffix types, got %d", errorCount)
		}
	})

	t.Run("should pass with operations without arguments", func(t *testing.T) {
		schema := `
		type Query {
			allUsers: [User!]!
			currentTime: String!
		}
		
		type Mutation {
			refreshCache: Boolean!
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "operation-input-name") > 0 {
			t.Error("Expected no errors for operations without arguments")
		}
	})

	t.Run("should flag incorrect argument names and Request suffix types", func(t *testing.T) {
		schema := `
		type Query {
			getUser(input: GetUserRequest!): User
			listUsers(params: ListUsersRequest!): [User!]!
		}
		
		type Mutation {
			createUser(data: CreateUserRequest!): User!
		}
		
		type User {
			id: ID!
		}
		
		input GetUserRequest {
			id: ID!
		}
		
		input ListUsersRequest {
			limit: Int
		}
		
		input CreateUserRequest {
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		errorCount := countRuleErrors(errors, "operation-input-name")
		if errorCount != 6 {
			t.Errorf("Expected 6 errors (3 wrong argument names + 3 forbidden Request suffix types), got %d", errorCount)
		}
	})

	t.Run("should pass with proper input type names without Request suffix", func(t *testing.T) {
		schema := `
		type Query {
			getUser(request: UserInput!): User
			listUsers(request: UsersFilter!): [User!]!
		}
		
		type Mutation {
			createUser(request: NewUserData!): User!
		}
		
		type User {
			id: ID!
		}
		
		input UserInput {
			id: ID!
		}
		
		input UsersFilter {
			limit: Int
		}
		
		input NewUserData {
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "operation-input-name") > 0 {
			t.Error("Expected no errors for input types without Request suffix")
		}
	})

	t.Run("should suggest consolidating multiple arguments", func(t *testing.T) {
		schema := `
		type Query {
			getUser(id: ID!, includeProfile: Boolean): User
			listUsers(limit: Int, offset: Int, filter: String): [User!]!
		}
		
		type Mutation {
			createUser(name: String!, email: String!, age: Int): User!
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		errorCount := countRuleErrors(errors, "operation-input-name")
		if errorCount != 3 {
			t.Errorf("Expected 3 errors suggesting argument consolidation, got %d", errorCount)
		}
	})

	t.Run("should handle mixed valid and invalid cases", func(t *testing.T) {
		schema := `
		type Query {
			getUser(request: GetUserRequest!): User
			listUsers(input: ListUsersRequest!): [User!]!
			searchUsers(query: String!, limit: Int): [User!]!
		}
		
		type Mutation {
			createUser(request: CreateUserRequest!): User!
			updateUser(data: UpdateUserInput!): User!
		}
		
		type User {
			id: ID!
		}
		
		input GetUserRequest {
			id: ID!
		}
		
		input ListUsersRequest {
			limit: Int
		}
		
		input CreateUserRequest {
			name: String!
		}
		
		input UpdateUserInput {
			id: ID!
			name: String
		}
		`
		errors := runRule(t, rule, schema)
		errorCount := countRuleErrors(errors, "operation-input-name")
		if errorCount != 6 {
			t.Errorf("Expected 6 errors (3 Request suffix types + 2 wrong argument names + 1 multiple arguments), got %d", errorCount)
		}
	})
}

func TestNoUnusedTypes(t *testing.T) {
	rule := NewNoUnusedTypes()

	t.Run("should flag unused types", func(t *testing.T) {
		schema := `
		type Query {
			user: User
		}
		
		type User {
			id: ID!
		}
		
		type UnusedType {
			value: String!
		}
		`
		errors := runRule(t, rule, schema)
		unusedTypeErrors := 0
		for _, err := range errors {
			if err.Rule == "no-unused-types" && err.Message == "Type `UnusedType` is declared but never used. Consider removing it or using it in the schema." {
				unusedTypeErrors++
			}
		}
		if unusedTypeErrors == 0 {
			t.Error("Expected error for unused type")
		}
	})

	t.Run("should pass used types", func(t *testing.T) {
		schema := `
		type Query {
			user: User
		}
		
		type User {
			id: ID!
			profile: Profile
		}
		
		type Profile {
			bio: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-unused-types") > 0 {
			t.Error("Expected no unused type errors for used types")
		}
	})
}

func TestCapitalizedDescriptions(t *testing.T) {
	rule := NewCapitalizedDescriptions()

	t.Run("should flag lowercase descriptions", func(t *testing.T) {
		schema := `
		type User {
			"""user identifier"""
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "capitalized-descriptions") == 0 {
			t.Error("Expected error for lowercase description")
		}
	})

	t.Run("should pass capitalized descriptions", func(t *testing.T) {
		schema := `
		type User {
			"""User identifier"""
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "capitalized-descriptions") > 0 {
			t.Error("Expected no capitalization errors for proper descriptions")
		}
	})
}

func TestEnumUnknownCase(t *testing.T) {
	rule := NewEnumUnknownCase()

	t.Run("should flag enums with UNKNOWN values", func(t *testing.T) {
		schema := `
		type Query {
			user: User
		}
		
		type User {
			status: UserStatus!
		}
		
		enum UserStatus {
			UNKNOWN
			ACTIVE
			INACTIVE
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "enum-unknown-case") == 0 {
			t.Error("Expected error for enum with UNKNOWN value")
		}

		// Check that the error message is correct
		expectedMessage := "Enum `UserStatus` contains an UNKNOWN value. UNKNOWN as a enum value is not allowed."
		if !containsError(errors, expectedMessage) {
			t.Error("Expected specific error message about UNKNOWN value")
		}
	})

	t.Run("should pass enums without UNKNOWN values", func(t *testing.T) {
		schema := `
		type Query {
			user: User
		}
		
		type User {
			status: UserStatus!
		}
		
		enum UserStatus {
			ACTIVE
			INACTIVE
			PENDING
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "enum-unknown-case") > 0 {
			t.Error("Expected no errors for enum without UNKNOWN value")
		}
	})

	t.Run("should flag multiple enums with UNKNOWN values", func(t *testing.T) {
		schema := `
		enum Status1 {
			UNKNOWN
			ACTIVE
		}
		
		enum Status2 {
			UNKNOWN
			INACTIVE
		}
		
		enum Status3 {
			ACTIVE
			INACTIVE
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "enum-unknown-case") != 2 {
			t.Errorf("Expected exactly 2 errors for enums with UNKNOWN values, got %d", countRuleErrors(errors, "enum-unknown-case"))
		}
	})
}

func TestNoQueryPrefixes(t *testing.T) {
	rule := NewNoQueryPrefixes()

	t.Run("should flag query fields with prefixes", func(t *testing.T) {
		schema := `
		type Query {
			getUser: User
			listUsers: [User!]!
			findProducts: [String!]!
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-query-prefixes") < 3 {
			t.Error("Expected at least 3 errors for prefixed queries")
		}
	})

	t.Run("should pass query fields without prefixes", func(t *testing.T) {
		schema := `
		type Query {
			user: User
			users: [User!]!
			products: [String!]!
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-query-prefixes") > 0 {
			t.Error("Expected no prefix errors for clean queries")
		}
	})
}

func TestInputEnumSuffix(t *testing.T) {
	rule := NewInputEnumSuffix()

	t.Run("should flag input enums without Input suffix", func(t *testing.T) {
		schema := `
		input CreateUserInput {
			role: Role!
		}
		
		enum Role {
			USER
			ADMIN
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "input-enum-suffix") == 0 {
			t.Error("Expected error for input enum without suffix")
		}
	})

	t.Run("should pass input enums with Input suffix", func(t *testing.T) {
		schema := `
		input CreateUserInput {
			role: RoleInput!
		}
		
		enum RoleInput {
			USER
			ADMIN
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "input-enum-suffix") > 0 {
			t.Error("Expected no suffix errors for proper input enum")
		}
	})
}

func TestEnumDescriptions(t *testing.T) {
	rule := NewEnumDescriptions()

	t.Run("should flag enum values without descriptions", func(t *testing.T) {
		schema := `
		enum UserStatus {
			ACTIVE
			INACTIVE
		}
		`
		errors := runRule(t, rule, schema)
		// Should flag ACTIVE and INACTIVE but not UNKNOWN
		if countRuleErrors(errors, "enum-descriptions") != 2 {
			t.Error("Expected at least 2 errors for enum values without descriptions")
		}
	})

	t.Run("should pass enum values with descriptions", func(t *testing.T) {
		schema := `
		enum UserStatus {
			"""User is active"""
			ACTIVE
			"""User is inactive"""
			INACTIVE
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "enum-descriptions") > 0 {
			t.Error("Expected no errors for described values")
		}
	})
}

func TestListNonNullItems(t *testing.T) {
	rule := NewListNonNullItems()

	t.Run("should flag lists with nullable items", func(t *testing.T) {
		schema := `
		type User {
			tags: [String]
			friends: [User]
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "list-non-null-items") < 2 {
			t.Error("Expected at least 2 errors for nullable list items")
		}
	})

	t.Run("should pass lists with non-null items", func(t *testing.T) {
		schema := `
		type User {
			tags: [String!]!
			friends: [User!]!
			users : [[User!]!]
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "list-non-null-items") > 0 {
			t.Error("Expected no list errors for non-null items")
		}
	})
}

func TestEnumReservedValues(t *testing.T) {
	rule := NewEnumReservedValues()

	t.Run("should flag reserved enum values", func(t *testing.T) {
		schema := `
		enum Status {
			UNKNOWN
			INVALID
			ACTIVE
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "enum-reserved-values") != 1 {
			t.Errorf("Expected exactly 1 error for reserved values, got %d", countRuleErrors(errors, "enum-reserved-values"))
		}
		// Verify the specific error message for INVALID
		found := false
		for _, err := range errors {
			if err.Rule == "enum-reserved-values" && err.Message == "Enum value `Status.INVALID` uses a reserved name." {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected error message for INVALID enum value not found")
		}
	})

	t.Run("should pass non-reserved enum values", func(t *testing.T) {
		schema := `
		enum Status {
			UNKNOWN
			ACTIVE
			INACTIVE
			PENDING
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "enum-reserved-values") > 0 {
			t.Error("Expected no reserved errors for clean enum values")
		}
	})
}

func TestMutationResponseNullable(t *testing.T) {
	rule := NewMutationResponseNullable()

	t.Run("should flag non-null mutation response fields", func(t *testing.T) {
		schema := `
		type Mutation {
			createUser: CreateUserResult!
		}
		
		type CreateUserResult {
			user: User!
			success: Boolean!
			message: String!
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") != 3 {
			t.Error("Expected exactly 3 errors for non-null response fields")
		}

		// Check specific error messages for response fields
		expectedFields := []string{"user", "success", "message"}
		for _, field := range expectedFields {
			found := false
			for _, err := range errors {
				if err.Rule == "mutation-response-nullable" &&
					strings.Contains(err.Message, fmt.Sprintf("Mutation response field `CreateUserResult.%s`", field)) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error for non-null response field %s", field)
			}
		}
	})

	t.Run("should pass valid mutation with non-null return and nullable response fields", func(t *testing.T) {
		schema := `
		type Mutation {
			createUser: CreateUserResult!
			updateUser: UpdateUserResult!
		}
		
		type CreateUserResult {
			user: User
			success: Boolean
			errors: [String]
		}
		
		type UpdateUserResult {
			user: User
			message: String
		}
		
		type User {
			id: ID!
			name: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") > 0 {
			t.Error("Expected no errors for valid mutation structure")
		}
	})

	t.Run("should handle mutations returning scalars", func(t *testing.T) {
		schema := `
		type Mutation {
			deleteUser: Boolean!
			getUserCount: Int!
		}
		`
		errors := runRule(t, rule, schema)
		// No response type fields to check, only mutation fields themselves
		if countRuleErrors(errors, "mutation-response-nullable") > 0 {
			t.Error("Expected no errors for scalar mutation returns")
		}
	})

	//t.Run("should handle list types in mutation responses", func(t *testing.T) {
	//	schema := `
	//	type Mutation {
	//		createUsers: CreateUsersResult!
	//	}
	//
	//	type CreateUsersResult {
	//		users: [User!]!
	//		errors: [String!]!
	//		successCount: Int!
	//	}
	//
	//	type User {
	//		id: ID!
	//	}
	//	`
	//	errors := runRule(t, rule, schema)
	//	if countRuleErrors(errors, "mutation-response-nullable") != 3 {
	//		t.Error("Expected exactly 3 errors for non-null list response fields")
	//	}
	//})

	t.Run("should handle nested object types", func(t *testing.T) {
		schema := `
		type Mutation {
			createOrder: CreateOrderResult!
		}
		
		type CreateOrderResult {
			order: Order!
			payment: Payment!
		}
		
		type Order {
			id: ID!
			items: [OrderItem!]!
		}
		
		type Payment {
			id: ID!
			amount: Float!
		}
		
		type OrderItem {
			id: ID!
			quantity: Int!
		}
		`
		errors := runRule(t, rule, schema)
		// Should only flag CreateOrderResult fields, not nested type fields
		if countRuleErrors(errors, "mutation-response-nullable") != 2 {
			t.Errorf("Expected exactly 2 errors for CreateOrderResult fields, got %d", countRuleErrors(errors, "mutation-response-nullable"))
		}
	})

	t.Run("should handle schema without mutations", func(t *testing.T) {
		schema := `
		type Query {
			user: User
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") > 0 {
			t.Error("Expected no errors for schema without mutations")
		}
	})

	t.Run("should handle union and interface return types", func(t *testing.T) {
		schema := `
		type Mutation {
			createContent: CreateContentResult!
		}
		
		type CreateContentResult {
			content: Content!
			success: Boolean!
		}
		
		union Content = Article | Video
		
		type Article {
			id: ID!
			title: String!
		}
		
		type Video {
			id: ID!
			duration: Int!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") < 2 {
			t.Error("Expected at least 2 errors for non-null response fields with union types")
		}
	})
}

func TestFieldsNullableExceptId(t *testing.T) {
	rule := NewFieldsNullableExceptId()

	t.Run("should flag non-null fields except ID", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			name: String!
			email: String!
			age: Int!
		}
		
		type Product {
			productId: ID!
			title: String!
			price: Float!
		}
		`
		errors := runRule(t, rule, schema)
		// Should flag name, email, age, title, price (5 total) but not id or productId
		if countRuleErrors(errors, "fields-nullable-except-id") != 5 {
			t.Errorf("Expected exactly 5 errors for non-null non-ID fields, got %d", countRuleErrors(errors, "fields-nullable-except-id"))
		}

		// Verify specific error messages
		expectedMessages := []string{
			"Field `User.name` should be nullable (`String` instead of `String!`)",
			"Field `User.email` should be nullable (`String` instead of `String!`)",
			"Field `User.age` should be nullable (`Int` instead of `Int!`)",
			"Field `Product.title` should be nullable (`String` instead of `String!`)",
			"Field `Product.price` should be nullable (`Float` instead of `Float!`)",
		}

		for _, expectedMsg := range expectedMessages {
			found := false
			for _, err := range errors {
				if err.Rule == "fields-nullable-except-id" && strings.Contains(err.Message, expectedMsg) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error message containing '%s' not found", expectedMsg)
			}
		}
	})

	t.Run("should pass nullable fields and ID fields", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			name: String
			email: String
			age: Int
			profile: UserProfile
		}
		
		type UserProfile {
			profileId: ID!
			bio: String
			avatar: String
		}
		
		type Query {
			user: User
			users: [User!]!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "fields-nullable-except-id") > 0 {
			t.Errorf("Expected no errors for nullable fields and ID fields, got %d", countRuleErrors(errors, "fields-nullable-except-id"))
		}
	})

	t.Run("should handle different ID field naming patterns", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			userId: ID!
			userID: ID!
			name: String!
		}
		
		type Order {
			orderId: ID!
			customerId: ID!
			amount: Float!
		}
		`
		errors := runRule(t, rule, schema)
		// Should only flag name and amount (2 errors), not the ID fields
		if countRuleErrors(errors, "fields-nullable-except-id") != 2 {
			t.Errorf("Expected exactly 2 errors for non-ID fields, got %d", countRuleErrors(errors, "fields-nullable-except-id"))
		}

		// Verify it flags the right fields
		expectedFields := []string{"User.name", "Order.amount"}
		for _, expectedField := range expectedFields {
			found := false
			for _, err := range errors {
				if err.Rule == "fields-nullable-except-id" && strings.Contains(err.Message, expectedField) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error for field '%s' not found", expectedField)
			}
		}
	})

	t.Run("should handle list types correctly", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			tags: [String!]!
			friends: [User!]!
			optionalTags: [String!]
		}
		`
		errors := runRule(t, rule, schema)
		// Should flag the non-null list fields (tags, friends) but not optionalTags
		if errors == nil || len(errors) != 2 {
			t.Errorf("UnExpected errors got %v", errors)
		}
	})

	t.Run("should skip root operation types", func(t *testing.T) {
		schema := `
		type Query {
			user: User!
			product: Product!
		}
		
		type Mutation {
			createUser: User!
			deleteUser: Boolean!
		}
		
		type Subscription {
			userUpdated: User!
		}
		
		type User {
			id: ID!
			name: String
		}
		
		type Product {
			id: ID!
			title: String
		}
		`
		errors := runRule(t, rule, schema)
		// Should not flag Query, Mutation, or Subscription fields
		if countRuleErrors(errors, "fields-nullable-except-id") > 0 {
			t.Errorf("Expected no errors for root types and nullable fields, got %d", countRuleErrors(errors, "fields-nullable-except-id"))
		}
	})

	t.Run("should handle non-ID fields with ID in name correctly", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			videoId: String!
			ideaTitle: String!
			identifier: Int!
		}
		`
		errors := runRule(t, rule, schema)
		// videoId ends with Id but is String type, so should be flagged
		// ideaTitle contains "id" but doesn't end with it, should be flagged
		// identifier contains "id" but doesn't end with Id/ID, should be flagged
		if countRuleErrors(errors, "fields-nullable-except-id") != 3 {
			t.Errorf("Expected exactly 3 errors for non-ID type fields, got %d", countRuleErrors(errors, "fields-nullable-except-id"))
		}
	})

	t.Run("should handle custom scalar types", func(t *testing.T) {
		schema := `
		scalar DateTime
		scalar UUID
		
		type User {
			id: ID!
			createdAt: DateTime!
			uuid: UUID!
			customField: String!
		}
		`
		errors := runRule(t, rule, schema)
		// Should flag all non-ID fields (createdAt, uuid, customField)
		if countRuleErrors(errors, "fields-nullable-except-id") != 3 {
			t.Errorf("Expected exactly 3 errors for non-ID scalar fields, got %d", countRuleErrors(errors, "fields-nullable-except-id"))
		}
	})

	t.Run("should skip excluded types like PageInfo", func(t *testing.T) {
		schema := `
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
			startCursor: String!
			endCursor: String!
		}
		
		type User {
			id: ID!
			name: String!
			email: String!
		}
		
		type Connection {
			edges: [Edge!]!
			pageInfo: PageInfo!
		}
		
		type Edge {
			node: User!
			cursor: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "fields-nullable-except-id") != 6 {
			t.Errorf("Expected exactly 6 errors (excluding PageInfo fields), got %d", countRuleErrors(errors, "fields-nullable-except-id"))
		}

		// Verify PageInfo fields are not flagged
		for _, err := range errors {
			if err.Rule == "fields-nullable-except-id" && strings.Contains(err.Message, "PageInfo.") {
				t.Errorf("PageInfo fields should be excluded from the rule, but got error: %s", err.Message)
			}
		}

		// Verify other types are still flagged (excluding list types)
		expectedFields := []string{"User.name", "User.email", "Connection.pageInfo", "Edge.node", "Edge.cursor", "Connection.edges"}
		for _, expectedField := range expectedFields {
			found := false
			for _, err := range errors {
				if err.Rule == "fields-nullable-except-id" && strings.Contains(err.Message, expectedField) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error for field '%s' not found", expectedField)
			}
		}
	})
}

func TestRelayPageInfo(t *testing.T) {
	rule := NewRelayPageInfo()

	t.Run("should flag PageInfo missing required fields", func(t *testing.T) {
		schema := `
		type PageInfo {
			hasNextPage: Boolean!
			# Missing hasPreviousPage, startCursor, endCursor
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-pageinfo") != 3 {
			t.Errorf("Expected exactly 3 errors for missing fields, got %d", countRuleErrors(errors, "relay-pageinfo"))
		}

		expectedMissingFields := []string{"hasPreviousPage", "startCursor", "endCursor"}
		for _, field := range expectedMissingFields {
			found := false
			for _, err := range errors {
				if err.Rule == "relay-pageinfo" && strings.Contains(err.Message, fmt.Sprintf("must contain field `%s`", field)) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected error for missing field '%s'", field)
			}
		}
	})

	t.Run("should flag PageInfo with incorrect field types", func(t *testing.T) {
		schema := `
		type PageInfo {
			hasNextPage: String!
			hasPreviousPage: Boolean
			startCursor: Int
			endCursor: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-pageinfo") != 4 {
			t.Errorf("Expected exactly 4 errors for incorrect field types, got %d", countRuleErrors(errors, "relay-pageinfo"))
		}

		// Check specific type errors
		expectedTypeErrors := map[string]string{
			"hasNextPage":     "Boolean!",
			"hasPreviousPage": "Boolean!",
			"startCursor":     "String",
			"endCursor":       "String",
		}

		for field, expectedType := range expectedTypeErrors {
			found := false
			for _, err := range errors {
				if err.Rule == "relay-pageinfo" &&
					strings.Contains(err.Message, fmt.Sprintf("field `%s` must return %s", field, expectedType)) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected type error for field '%s' (should be %s)", field, expectedType)
			}
		}
	})

	t.Run("should pass valid PageInfo", func(t *testing.T) {
		schema := `
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-pageinfo") > 0 {
			t.Errorf("Expected no errors for valid PageInfo, got %d", countRuleErrors(errors, "relay-pageinfo"))
		}
	})

	t.Run("should allow additional fields on PageInfo", func(t *testing.T) {
		schema := `
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
			startCursor: String
			endCursor: String
			totalCount: Int
			hasmore: Boolean
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-pageinfo") > 0 {
			t.Errorf("Expected no errors for PageInfo with additional fields, got %d", countRuleErrors(errors, "relay-pageinfo"))
		}
	})
}

func TestOperationResponseName(t *testing.T) {
	rule := NewOperationResponseName()

	t.Run("should flag response types with Response suffix", func(t *testing.T) {
		schema := `
		type Query {
			getUser: GetUserResponse!
			listUsers: ListUsersResponse!
			searchUser: SearchUserResponse!
		}
		
		type Mutation {
			createUser(request: CreateUserRequest!): CreateUserResponse!
			updateUser(request: UpdateUserRequest!): UpdateUserResponse!
		}
		
		type GetUserResponse {
			user: User
		}
		
		type ListUsersResponse {
			users: [User!]!
		}
		
		type SearchUserResponse {
			users: [User!]!
		}
		
		type CreateUserResponse {
			user: User
		}
		
		type UpdateUserResponse {
			user: User
		}
		
		type User {
			id: ID!
			name: String!
		}
		
		input CreateUserRequest {
			name: String!
		}
		
		input UpdateUserRequest {
			id: ID!
			name: String!
		}
		`

		errors := runRule(t, rule, schema)
		ruleErrors := countRuleErrors(errors, "operation-response-name")
		if ruleErrors != 5 {
			t.Errorf("Expected 5 errors for forbidden Response suffix types, got %d", ruleErrors)
		}
	})

	t.Run("should pass with response type names without Response suffix", func(t *testing.T) {
		schema := `
		type Query {
			getUser: UserResult!
			listUsers: UsersData!
		}
		
		type Mutation {
			createUser: CreateUserResult!
		}
		
		type UserResult {
			user: User
		}
		
		type UsersData {
			users: [User!]!
		}
		
		type CreateUserResult {
			user: User
		}
		
		type User {
			id: ID!
			name: String!
		}
		`

		errors := runRule(t, rule, schema)
		ruleErrors := countRuleErrors(errors, "operation-response-name")

		if ruleErrors != 0 {
			t.Errorf("Expected no errors for response types without Response suffix, got %d", ruleErrors)
		}
	})

	t.Run("should flag versioned Response response types", func(t *testing.T) {
		schema := `
		type Query {
			getUser: GetUserResponseV2!
			listUsers: ListUsersResponseVersion3!
		}
		
		type Mutation {
			createUser: CreateUserResponseV1!
		}
		
		type GetUserResponseV2 {
			user: User
		}
		
		type ListUsersResponseVersion3 {
			users: [User!]!
		}
		
		type CreateUserResponseV1 {
			user: User
		}
		
		type User {
			id: ID!
			name: String!
		}
		`

		errors := runRule(t, rule, schema)
		ruleErrors := countRuleErrors(errors, "operation-response-name")

		if ruleErrors != 3 {
			t.Errorf("Expected 3 errors for forbidden versioned Response suffix types, got %d", ruleErrors)
		}
	})
}

func TestRelayEdgeTypes(t *testing.T) {
	rule := NewRelayEdgeTypes()

	t.Run("should pass valid Edge types referenced by Connection types", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		type User implements Node {
			id: ID!
			name: String
		}
		
		type UserEdge {
			node: User
			cursor: String!
		}
		
		type PostEdge {
			node: Post!
			cursor: String
		}
		
		type Post implements Node {
			id: ID!
			title: String
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PostConnection {
			edges: [PostEdge!]!
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors for valid Edge types, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should flag Edge type that is not Object type", func(t *testing.T) {
		schema := `
		interface UserEdge {
			node: User
			cursor: String
		}
		
		type User {
			id: ID!
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-edge-types") != 1 {
			t.Errorf("Expected exactly 1 error for non-Object Edge type, got %d", countRuleErrors(errors, "relay-edge-types"))
		}

		expectedMessage := "Edge type `UserEdge` must be an Object type, but is INTERFACE."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag Edge type missing node field", func(t *testing.T) {
		schema := `
		type UserEdge {
			cursor: String!
			data: User
		}
		
		type User {
			id: ID!
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-edge-types") != 1 {
			t.Errorf("Expected exactly 1 error for missing node field, got %d", countRuleErrors(errors, "relay-edge-types"))
		}

		expectedMessage := "Edge type `UserEdge` must contain a field `node` that returns either Scalar, Enum, Object, Interface, Union, or a non-null wrapper around one of those types."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag Edge type missing cursor field", func(t *testing.T) {
		schema := `
		type UserEdge {
			node: User!
			id: String!
		}
		
		type User {
			id: ID!
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-edge-types") != 1 {
			t.Errorf("Expected exactly 1 error for missing cursor field, got %d", countRuleErrors(errors, "relay-edge-types"))
		}

		expectedMessage := "Edge type `UserEdge` must contain a field `cursor` that returns either String, Scalar, or a non-null wrapper around one of those types."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag Edge type with node field returning list", func(t *testing.T) {
		schema := `
		type UserEdge {
			node: [User!]!
			cursor: String!
		}
		
		type User {
			id: ID!
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-edge-types") != 1 {
			t.Errorf("Expected exactly 1 error for node field returning list, got %d", countRuleErrors(errors, "relay-edge-types"))
		}

		expectedMessage := "Edge type `UserEdge` field `node` cannot return a list type, but returns [User!]!."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	//t.Run("should flag Edge type with cursor field returning invalid type", func(t *testing.T) {
	//	schema := `
	//	type UserEdge {
	//		node: User!
	//		cursor: Int!
	//	}
	//
	//	type User {
	//		id: ID!
	//	}
	//	`
	//	errors := runRule(t, rule, schema)
	//	if countRuleErrors(errors, "relay-edge-types") != 1 {
	//		t.Errorf("Expected exactly 1 error for invalid cursor type, got %d", countRuleErrors(errors, "relay-edge-types"))
	//	}
	//
	//	expectedMessage := "Edge type `UserEdge` field `cursor` must return String, Scalar, or a non-null wrapper around one of those types, but returns Int!."
	//	if !containsError(errors, expectedMessage) {
	//		t.Errorf("Expected error message: %s", expectedMessage)
	//	}
	//})

	t.Run("should ignore types not referenced by Connection edges field", func(t *testing.T) {
		schema := `
		type UserDetail {
			node: User!
			cursor: String!
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		// This should NOT be flagged because it's not referenced by any Connection type
		// so it won't be detected as an Edge type
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors for type not referenced by Connection, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should flag Edge type node field not implementing Node interface", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		type User {
			id: ID!
			name: String
		}
		
		type UserEdge {
			node: User!
			cursor: String!
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-edge-types") != 1 {
			t.Errorf("Expected exactly 1 error for node not implementing Node interface, got %d", countRuleErrors(errors, "relay-edge-types"))
		}

		expectedMessage := "Edge type `UserEdge` field `node` type `User` must implement Node interface."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should pass when Node interface doesn't exist", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			name: String
		}
		
		type UserEdge {
			node: User!
			cursor: String!
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		// Should pass because if Node interface doesn't exist, we can't enforce this rule
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors when Node interface doesn't exist, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should handle nullable cursor field", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		type User implements Node {
			id: ID!
			name: String
		}
		
		type UserEdge {
			node: User!
			cursor: String
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors for nullable cursor field, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should handle interface node type", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		interface ContentNode implements Node {
			id: ID!
			title: String
		}
		
		type ContentEdge {
			node: ContentNode!
			cursor: String!
		}
		
		type ContentConnection {
			edges: [ContentEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors for interface node type implementing Node, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should handle union node type", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		type User implements Node {
			id: ID!
			name: String
		}
		
		type Post implements Node {
			id: ID!
			title: String
		}
		
		union Content = User | Post
		
		type ContentEdge {
			node: Content!
			cursor: String!
		}
		
		type ContentConnection {
			edges: [ContentEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		// Union types don't implement interfaces, so this should pass
		// because the rule only checks Object and Interface types for Node implementation
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors for union node type, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should handle enum node type", func(t *testing.T) {
		schema := `
		enum Status {
			ACTIVE
			INACTIVE
		}
		
		type StatusEdge {
			node: Status!
			cursor: String!
		}
		
		type StatusConnection {
			edges: [StatusEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		// Enum types don't implement interfaces, so this should pass
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors for enum node type, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should handle scalar node type", func(t *testing.T) {
		schema := `
		scalar DateTime
		
		type TimeEdge {
			node: DateTime!
			cursor: String!
		}
		
		type TimeConnection {
			edges: [TimeEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		// Scalar types don't implement interfaces, so this should pass
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors for scalar node type, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should handle multiple errors for same Edge type", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		type User {
			id: ID!
			name: String
		}
		
		type UserEdge {
			data: [User!]!
			id: Int!
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		// Should have errors for: missing node field, missing cursor field
		if countRuleErrors(errors, "relay-edge-types") != 2 {
			t.Errorf("Expected exactly 2 errors for multiple violations, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should handle Edge type with additional fields", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		type User implements Node {
			id: ID!
			name: String
		}
		
		type UserEdge {
			node: User!
			cursor: String!
			metadata: String
			createdAt: String
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		// Should pass even with additional fields
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors for Edge type with additional fields, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should handle schema without any Connection types", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			name: String
		}
		
		type Post {
			id: ID!
			title: String
		}
		
		type UserEdge {
			node: User!
			cursor: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors for schema without Connection types, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should detect Edge types from Connection edges field", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		type User implements Node {
			id: ID!
			name: String
		}
		
		type UserDetail {
			node: User!
			cursor: String!
		}
		
		type UserConnection {
			edges: [UserDetail]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		// UserDetail should be detected as Edge type from UserConnection.edges
		// and should pass all Edge type validations
		if countRuleErrors(errors, "relay-edge-types") > 0 {
			t.Errorf("Expected no errors for Edge type detected from Connection, got %d", countRuleErrors(errors, "relay-edge-types"))
		}
	})

	t.Run("should flag Edge type detected from Connection with violations", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		type User {
			id: ID!
			name: String
		}
		
		type UserDetail {
			node: User!
			cursor: String!
		}
		
		type UserConnection {
			edges: [UserDetail]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		// UserDetail should be detected as Edge type and flagged for User not implementing Node
		if countRuleErrors(errors, "relay-edge-types") != 1 {
			t.Errorf("Expected exactly 1 error for Edge type detected from Connection with violations, got %d", countRuleErrors(errors, "relay-edge-types"))
		}

		expectedMessage := "Edge type `UserDetail` field `node` type `User` must implement Node interface."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag union node type members not implementing Node", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		type User implements Node {
			id: ID!
			name: String
		}
		
		type Post {
			id: ID!
			title: String
		}
		
		union Content = User | Post
		
		type ContentEdge {
			node: Content!
			cursor: String!
		}
		
		type ContentConnection {
			edges: [ContentEdge]
			pageInfo: PageInfo!
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
		}
		`
		errors := runRule(t, rule, schema)
		// Should flag Post for not implementing Node interface
		if countRuleErrors(errors, "relay-edge-types") != 1 {
			t.Errorf("Expected exactly 1 error for union member not implementing Node, got %d", countRuleErrors(errors, "relay-edge-types"))
		}

		expectedMessage := "Edge type `ContentEdge` field `node` union type `Content` member `Post` must implement Node interface."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})
}

func TestRelayConnectionTypes(t *testing.T) {
	rule := NewRelayConnectionTypes()

	t.Run("should pass valid Connection types", func(t *testing.T) {
		schema := `
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
			startCursor: String
			endCursor: String
		}
		
		type UserEdge {
			node: User
			cursor: String!
		}
		
		type User {
			id: ID!
			name: String
		}
		
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type PostConnection {
			edges: [PostEdge!]!
			pageInfo: PageInfo!
			totalCount: Int
		}
		
		type PostEdge {
			node: Post!
			cursor: String!
		}
		
		type Post {
			id: ID!
			title: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") > 0 {
			t.Errorf("Expected no errors for valid Connection types, got %d", countRuleErrors(errors, "relay-connection-types"))
		}
	})

	t.Run("should flag Connection type that is not Object type", func(t *testing.T) {
		schema := `
		interface UserConnection {
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
			hasPreviousPage: Boolean!
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") != 1 {
			t.Errorf("Expected exactly 1 error for non-Object Connection type, got %d", countRuleErrors(errors, "relay-connection-types"))
		}

		expectedMessage := "Connection type `UserConnection` must be an Object type, but is INTERFACE."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag Connection type missing edges field", func(t *testing.T) {
		schema := `
		type UserConnection {
			pageInfo: PageInfo!
			items: [User!]!
		}
		
		type User {
			id: ID!
			name: String
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") != 1 {
			t.Errorf("Expected exactly 1 error for missing edges field, got %d", countRuleErrors(errors, "relay-connection-types"))
		}

		expectedMessage := "Connection type `UserConnection` must contain a field `edges` that returns a list type."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag Connection type with edges field not returning list", func(t *testing.T) {
		schema := `
		type UserConnection {
			edges: UserEdge!
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
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") != 1 {
			t.Errorf("Expected exactly 1 error for edges field not returning list, got %d", countRuleErrors(errors, "relay-connection-types"))
		}

		expectedMessage := "Connection type `UserConnection` field `edges` must return a list type, but returns UserEdge!."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag Connection type missing pageInfo field", func(t *testing.T) {
		schema := `
		type UserConnection {
			edges: [UserEdge]
			pagination: PageInfo!
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
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") != 1 {
			t.Errorf("Expected exactly 1 error for missing pageInfo field, got %d", countRuleErrors(errors, "relay-connection-types"))
		}

		expectedMessage := "Connection type `UserConnection` must contain a field `pageInfo` that returns a non-null PageInfo Object type."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag Connection type with nullable pageInfo field", func(t *testing.T) {
		schema := `
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo
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
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") != 1 {
			t.Errorf("Expected exactly 1 error for nullable pageInfo field, got %d", countRuleErrors(errors, "relay-connection-types"))
		}

		expectedMessage := "Connection type `UserConnection` must contain a field `pageInfo` that returns a non-null PageInfo Object type."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should handle multiple Connection types with different violations", func(t *testing.T) {
		schema := `
		interface PostConnection {
			edges: [PostEdge]
			pageInfo: PageInfo!
		}
		
		type UserConnection {
			items: [User!]!
			pageInfo: PageInfo
		}
		
		type ProductConnection {
			edges: Product
			pageInfo: PageInfo!
		}
		
		type User {
			id: ID!
			name: String
		}
		
		type Product {
			id: ID!
			name: String
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
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)

		// PostConnection: Interface type (1 error)
		// UserConnection: Missing edges field + nullable pageInfo (2 errors)
		// ProductConnection: edges not a list type (1 error)
		expectedErrors := 4
		if countRuleErrors(errors, "relay-connection-types") != expectedErrors {
			t.Errorf("Expected exactly %d errors for multiple violations, got %d", expectedErrors, countRuleErrors(errors, "relay-connection-types"))
		}

		// Check specific error messages
		expectedMessages := []string{
			"Connection type `PostConnection` must be an Object type, but is INTERFACE.",
			"Connection type `UserConnection` must contain a field `edges` that returns a list type.",
			"Connection type `UserConnection` must contain a field `pageInfo` that returns a non-null PageInfo Object type.",
			"Connection type `ProductConnection` field `edges` must return a list type, but returns Product.",
		}

		for _, expectedMsg := range expectedMessages {
			if !containsError(errors, expectedMsg) {
				t.Errorf("Expected error message: %s", expectedMsg)
			}
		}
	})

	t.Run("should ignore types that don't end with Connection", func(t *testing.T) {
		schema := `
		type UserList {
			items: [User!]!
		}
		
		type UserContainer {
			data: User
		}
		
		type User {
			id: ID!
			name: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") > 0 {
			t.Errorf("Expected no errors for types not ending with Connection, got %d", countRuleErrors(errors, "relay-connection-types"))
		}
	})

	t.Run("should ignore built-in and introspection types", func(t *testing.T) {
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
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)

		// Should not have errors for built-in types like String, ID, Boolean, etc.
		for _, err := range errors {
			if strings.Contains(err.Message, "String") ||
				strings.Contains(err.Message, "ID") ||
				strings.Contains(err.Message, "Boolean") ||
				strings.Contains(err.Message, "__") {
				t.Errorf("Should not validate built-in or introspection types: %s", err.Message)
			}
		}
	})

	t.Run("should allow additional fields on Connection types", func(t *testing.T) {
		schema := `
		type UserConnection {
			edges: [UserEdge]
			pageInfo: PageInfo!
			totalCount: Int
			hasMore: Boolean
			aggregates: UserAggregates
		}
		
		type UserEdge {
			node: User
			cursor: String!
		}
		
		type User {
			id: ID!
			name: String
		}
		
		type UserAggregates {
			count: Int
			avgAge: Float
		}
		
		type PageInfo {
			hasNextPage: Boolean!
			hasPreviousPage: Boolean!
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") > 0 {
			t.Errorf("Expected no errors for Connection with additional fields, got %d", countRuleErrors(errors, "relay-connection-types"))
		}
	})

	t.Run("should handle different list type variations", func(t *testing.T) {
		schema := `
		type Connection1 {
			edges: [UserEdge]
			pageInfo: PageInfo!
		}
		
		type Connection2 {
			edges: [UserEdge!]
			pageInfo: PageInfo!
		}
		
		type Connection3 {
			edges: [UserEdge]!
			pageInfo: PageInfo!
		}
		
		type Connection4 {
			edges: [UserEdge!]!
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
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") > 0 {
			t.Errorf("Expected no errors for different list type variations, got %d", countRuleErrors(errors, "relay-connection-types"))
		}
	})

	t.Run("should flag Connection type with nested list edges field", func(t *testing.T) {
		schema := `
		type UserConnection {
			edges: [[UserEdge]]
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
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") != 1 {
			t.Errorf("Expected exactly 1 error for nested list edges field, got %d", countRuleErrors(errors, "relay-connection-types"))
		}

		expectedMessage := "Connection type `UserConnection` field `edges` must return a single-level list type, but returns a nested list [[UserEdge]]."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})

	t.Run("should flag Connection type with nested non-null list edges field", func(t *testing.T) {
		schema := `
		type UserConnection {
			edges: [[UserEdge!]!]!
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
			startCursor: String
			endCursor: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "relay-connection-types") != 1 {
			t.Errorf("Expected exactly 1 error for nested non-null list edges field, got %d", countRuleErrors(errors, "relay-connection-types"))
		}

		expectedMessage := "Connection type `UserConnection` field `edges` must return a single-level list type, but returns a nested list [[UserEdge!]!]!."
		if !containsError(errors, expectedMessage) {
			t.Errorf("Expected error message: %s", expectedMessage)
		}
	})
}

func TestUnsupportedDirectives(t *testing.T) {
	rule := NewUnsupportedDirectives()

	t.Run("should pass when no unsupported directives are defined and used", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			name: String @deprecated(reason: "Use fullName instead")
		}
		
		enum Status {
			ACTIVE
			INACTIVE
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "unsupported-directives") > 0 {
			t.Error("Expected no errors when no unsupported directives are used")
		}
	})

	t.Run("should flag multiple unsupported directives defined", func(t *testing.T) {
		schema := `
		directive @inaccessible on FIELD_DEFINITION
		directive @external on FIELD_DEFINITION
		directive @requires(fields: String!) on FIELD_DEFINITION
		
		type User {
			id: ID!
			name: String @inaccessible @external @requires(fields: "id")
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "unsupported-directives") == 3 {
			t.Error("Expected at least 3 errors for multiple unsupported directives defined")
		}
	})
}

func TestNoUnimplementedInterface(t *testing.T) {
	rule := NewNoUnimplementedInterface()

	t.Run("should flag interfaces with no implementations", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		interface Timestamped {
			createdAt: String!
			updatedAt: String!
		}
		
		type User {
			id: ID!
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-unimplemented-interface") != 2 {
			t.Errorf("Expected exactly 2 errors for unimplemented interfaces, got %d", countRuleErrors(errors, "no-unimplemented-interface"))
		}

		expectedMessages := []string{
			"Interface 'Node' is not implemented by any type",
			"Interface 'Timestamped' is not implemented by any type",
		}

		for _, expectedMessage := range expectedMessages {
			if !containsError(errors, expectedMessage) {
				t.Errorf("Expected error message: %s", expectedMessage)
			}
		}
	})

	t.Run("should pass interfaces that are implemented by object types", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		interface Timestamped {
			createdAt: String!
			updatedAt: String!
		}
		
		type User implements Node & Timestamped {
			id: ID!
			name: String!
			createdAt: String!
			updatedAt: String!
		}
		
		type Post implements Node {
			id: ID!
			title: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-unimplemented-interface") > 0 {
			t.Errorf("Expected no errors for implemented interfaces, got %d", countRuleErrors(errors, "no-unimplemented-interface"))
		}
	})

	t.Run("should pass interfaces that are implemented by other interfaces", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		interface ContentNode implements Node {
			id: ID!
			title: String!
		}
		
		type Post implements ContentNode & Node {
			id: ID!
			title: String!
			content: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-unimplemented-interface") > 0 {
			t.Errorf("Expected no errors for interfaces implemented by other interfaces, got %d", countRuleErrors(errors, "no-unimplemented-interface"))
		}
	})

	t.Run("should handle mixed scenario with some implemented and some unimplemented interfaces", func(t *testing.T) {
		schema := `
		interface Node {
			id: ID!
		}
		
		interface Timestamped {
			createdAt: String!
			updatedAt: String!
		}
		
		interface Unused {
			value: String!
		}
		
		type User implements Node {
			id: ID!
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-unimplemented-interface") != 2 {
			t.Errorf("Expected exactly 2 errors for unimplemented interfaces, got %d", countRuleErrors(errors, "no-unimplemented-interface"))
		}

		expectedMessages := []string{
			"Interface 'Timestamped' is not implemented by any type",
			"Interface 'Unused' is not implemented by any type",
		}

		for _, expectedMessage := range expectedMessages {
			if !containsError(errors, expectedMessage) {
				t.Errorf("Expected error message: %s", expectedMessage)
			}
		}

		// Node should not be flagged as it's implemented by User
		unexpectedMessage := "Interface 'Node' is not implemented by any type"
		if containsError(errors, unexpectedMessage) {
			t.Errorf("Did not expect error message: %s", unexpectedMessage)
		}
	})
}
