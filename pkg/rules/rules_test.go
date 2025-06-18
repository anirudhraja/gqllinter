package rules

import (
	"testing"

	"github.com/anirudhraja/gqllinter/pkg/types"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
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

	t.Run("should flag invalid type names", func(t *testing.T) {
		schema := `
		type user_data {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "naming-convention") == 0 {
			t.Error("Expected error for invalid type name")
		}
	})

	t.Run("should flag generic type names", func(t *testing.T) {
		schema := `
		type Data {
			value: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "naming-convention") == 0 {
			t.Error("Expected error for generic type name")
		}
	})

	t.Run("should pass valid type names", func(t *testing.T) {
		schema := `
		type UserProfile {
			id: ID!
			displayName: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "naming-convention") > 0 {
			t.Error("Expected no naming convention errors for valid names")
		}
	})
}

func TestNoFieldNamespacing(t *testing.T) {
	rule := NewNoFieldNamespacing()

	t.Run("should flag namespaced fields", func(t *testing.T) {
		schema := `
		type User {
			userId: ID!
			userName: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-field-namespacing") < 2 {
			t.Error("Expected at least 2 errors for namespaced fields")
		}
	})

	t.Run("should pass non-namespaced fields", func(t *testing.T) {
		schema := `
		type User {
			id: ID!
			name: String!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "no-field-namespacing") > 0 {
			t.Error("Expected no namespacing errors")
		}
	})
}

func TestMinimalTopLevelQueries(t *testing.T) {
	rule := NewMinimalTopLevelQueries()

	t.Run("should flag excessive top-level queries", func(t *testing.T) {
		schema := `
		type Query {
			user1: String
			user2: String
			user3: String
			user4: String
			user5: String
			user6: String
			user7: String
			user8: String
			user9: String
			user10: String
			user11: String
			user12: String
			user13: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "minimal-top-level-queries") == 0 {
			t.Error("Expected error for excessive top-level queries")
		}
	})

	t.Run("should pass reasonable number of queries", func(t *testing.T) {
		schema := `
		type Query {
			user: String
			posts: String
			comments: String
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "minimal-top-level-queries") > 0 {
			t.Error("Expected no errors for reasonable query count")
		}
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

	t.Run("should flag incorrect mutation input naming", func(t *testing.T) {
		schema := `
		type Mutation {
			createUser(data: CreateUserData!): User
		}
		
		input CreateUserData {
			name: String!
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "input-name") < 2 {
			t.Error("Expected at least 2 errors for incorrect input naming")
		}
	})

	t.Run("should pass correct mutation input naming", func(t *testing.T) {
		schema := `
		type Mutation {
			createUser(input: CreateUserInput!): User
		}
		
		input CreateUserInput {
			name: String!
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "input-name") > 0 {
			t.Error("Expected no input-name errors for correct naming")
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

	t.Run("should flag output enums without UNKNOWN", func(t *testing.T) {
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
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "enum-unknown-case") == 0 {
			t.Error("Expected error for enum without UNKNOWN case")
		}
	})

	t.Run("should pass output enums with UNKNOWN", func(t *testing.T) {
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
		if countRuleErrors(errors, "enum-unknown-case") > 0 {
			t.Error("Expected no unknown errors for enum with UNKNOWN case")
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
			UNKNOWN
			ACTIVE
			INACTIVE
		}
		`
		errors := runRule(t, rule, schema)
		// Should flag ACTIVE and INACTIVE but not UNKNOWN
		if countRuleErrors(errors, "enum-descriptions") < 2 {
			t.Error("Expected at least 2 errors for enum values without descriptions")
		}
	})

	t.Run("should pass enum values with descriptions", func(t *testing.T) {
		schema := `
		enum UserStatus {
			UNKNOWN
			"""User is active"""
			ACTIVE
			"""User is inactive"""
			INACTIVE
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "enum-descriptions") > 0 {
			t.Error("Expected no enum description errors for described values")
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
		if countRuleErrors(errors, "enum-reserved-values") < 2 {
			t.Error("Expected at least 2 errors for reserved values")
		}
	})

	t.Run("should pass non-reserved enum values", func(t *testing.T) {
		schema := `
		enum Status {
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
		}
		
		type User {
			id: ID!
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") < 2 {
			t.Error("Expected at least 2 errors for non-null response fields")
		}
	})

	t.Run("should pass nullable mutation response fields", func(t *testing.T) {
		schema := `
		type Mutation {
			createUser: CreateUserResult
		}
		
		type CreateUserResult {
			user: User
			success: Boolean
		}
		
		type User {
			id: ID
		}
		`
		errors := runRule(t, rule, schema)
		if countRuleErrors(errors, "mutation-response-nullable") > 0 {
			t.Error("Expected no nullable errors for nullable response fields")
		}
	})
}
