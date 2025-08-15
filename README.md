# GraphQL Schema Linter

[![CI](https://github.com/anirudhraja/gqllinter/actions/workflows/ci.yml/badge.svg)](https://github.com/anirudhraja/gqllinter/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/anirudhraja/gqllinter)](https://goreportcard.com/report/github.com/anirudhraja/gqllinter)
[![Go Reference](https://pkg.go.dev/badge/github.com/anirudhraja/gqllinter.svg)](https://pkg.go.dev/github.com/anirudhraja/gqllinter)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A comprehensive GraphQL schema linter that helps enforce schema design best practices, based on [Yelp's GraphQL guidelines](https://yelp.github.io/graphql-guidelines/schema-design.html) and other industry standards.

## Features

- **Extensible Rule System**: Implement custom rules by satisfying the `Rule` interface
- **Multiple Output Formats**: Support for both text and JSON output formats
- **Glob Pattern Support**: Lint multiple files using glob patterns
- **Built-in Rules**: Comprehensive set of rules following industry best practices
- **Command Line Interface**: Easy-to-use CLI similar to existing GraphQL linters

## Installation

```bash
go install github.com/gqllinter@latest
```

Or build from source:

```bash
git clone https://github.com/gqllinter/gqllinter.git
cd gqllinter
go build -o gqllinter
```

## Usage

### Basic Usage

```bash
# Lint a single file
gqllinter schema.graphql

# Lint multiple files with glob patterns
gqllinter schema/*.graphql

# Output as JSON
gqllinter --format json schema.graphql

# Save output to file
gqllinter --format json --output results.json schema.graphql

# Run only specific rules
gqllinter --rules types-have-descriptions,fields-have-descriptions schema.graphql
```

### Command Line Options

```
Usage:
  gqllinter [flags] <schema-files>

Flags:
      --config string              path to configuration file
      --custom-rule-paths string   path to custom rules directory
      --format string              output format (text, json) (default "text")
      --ignore string              comment to ignore linting errors (default "# gqllinter-ignore")
      --output string              output file (default: stdout)
      --rules strings              comma-separated list of rules to run
```

## Rules Overview

## Implemented Validations

| Rule Name | Category | Description | Example Issue Detected |
|-----------|----------|-------------|------------------------|
| **types-have-descriptions** | Documentation | All types must have descriptions | `type User { id: ID! }` missing description |
| **fields-have-descriptions** | Documentation | All fields must have descriptions | `name: String!` missing description |
| **no-hashtag-description** | Documentation | Use triple quotes for descriptions, not hashtag comments | `# This is a user` instead of `"""This is a user"""` |
| **capitalized-descriptions** | Documentation | All descriptions must start with capital letters | `"""user name"""` should be `"""User name"""` |
| **enum-descriptions** | Documentation | All enum values must have descriptions (except UNKNOWN) | `ACTIVE` enum value missing description |
| **naming-convention** | Naming | Enforce UpperCamelCase for types, lowerCamelCase for fields | `type user_data` should be `type UserData` |
| **no-field-namespacing** | Naming | Fields shouldn't repeat their parent type name | `User.userName` should be `User.name` |
| **no-query-prefixes** | Naming | Query fields shouldn't have get/list/find prefixes | `getUser` should be `user` |
| **input-name** | Naming | Mutation inputs should be named consistently | `createUser(data: UserData!)` should be `createUser(input: CreateUserInput!)` |
| **input-enum-suffix** | Naming | Input enums should be distinct and suffixed with "Input" | Input enum `Role` should be `RoleInput` |
| **minimal-top-level-queries** | Schema Design | Keep top-level Query fields to a minimum | Query type with 15+ fields should be reorganized |
| **no-unused-fields** | Schema Design | Remove fields that are never referenced | Unused field `User.oldField` should be removed |
| **no-unused-types** | Schema Design | Remove types that are never referenced | Unused type `UnusedType` should be removed |
| **enum-unknown-case** | Schema Design | Output enums should have UNKNOWN case for extensibility | `enum Status { ACTIVE, INACTIVE }` missing `UNKNOWN` |
| **require-deprecation-reason** | Schema Evolution | Deprecated fields must have meaningful reasons | `@deprecated` should be `@deprecated(reason: "Use newField instead")` |
| **no-scalar-result-type-on-mutation** | Schema Evolution | Mutations should return object types, not scalars | `createUser(): Boolean` should return `CreateUserResult` |
| **mutation-response-nullable** | Schema Evolution | Mutation response fields should be nullable | `user: User!` should be `user: User` for flexibility |
| **alphabetize** | Organization | Fields and enum values should be alphabetically ordered | Fields `[name, id, email]` should be `[email, id, name]` |
| **list-non-null-items** | Type Safety | List types should contain non-null items | `tags: [String]` should be `tags: [String!]!` |
| **enum-reserved-values** | Extensibility | Avoid using reserved enum values | `UNKNOWN`, `INVALID` are reserved for system use |

## Available Rules

### Yelp Guidelines Rules

### types-have-descriptions
All types should have descriptions to explain their purpose.

**Bad:**
```graphql
type User {
  id: ID!
  name: String!
}
```

**Good:**
```graphql
"""
Represents a user in the system
"""
type User {
  id: ID!
  name: String!
}
```

### fields-have-descriptions
All fields should have descriptions to explain their purpose.

**Bad:**
```graphql
type User {
  id: ID!
  name: String!
  email: String!
}
```

**Good:**
```graphql
type User {
  """Unique identifier for the user"""
  id: ID!
  """Full name of the user"""
  name: String!
  """Email address for the user"""
  email: String!
}
```

### no-hashtag-description
Use triple quotes for descriptions instead of hashtag comments.

**Bad:**
```graphql
# This is a user type
type User {
  id: ID!
}
```

**Good:**
```graphql
"""
This is a user type
"""
type User {
  id: ID!
}
```

### naming-convention
Enforce proper naming conventions for types and fields.

**Bad:**
```graphql
type user_data {
  user_id: ID!
  user_name: String!
}
```

**Good:**
```graphql
type UserData {
  userId: ID!
  userName: String!
}
```

### link-via-types
Link via types, not IDs - following Yelp guidelines for better GraphQL design.

**Bad:**
```graphql
type Post {
  id: ID!
  authorId: ID!
  title: String!
}
```

**Good:**
```graphql
type Post {
  id: ID!
  author: User!
  title: String!
}
```

### no-field-namespacing
Fields don't need to be namespaced with their parent type name.

**Bad:**
```graphql
type User {
  userId: ID!
  userName: String!
  userEmail: String!
}
```

**Good:**
```graphql
type User {
  id: ID!
  name: String!
  email: String!
}
```

### minimal-top-level-queries
Keep the top level queries to a minimum - following Yelp guidelines for better schema organization.

**Bad:**
```graphql
type Query {
  getUser: User
  getUserById: User
  getUserByEmail: User
  getUserByName: User
  # ... 20+ more query fields
}
```

**Good:**
```graphql
type Query {
  user(id: ID, email: String, name: String): User
  users(filters: UserFilters): [User!]!
}
```

### use-standard-scalars
Use existing standardized types and scalars.

**Bad:**
```graphql
type User {
  email: String!
  website: String!
  createdAt: String!
}
```

**Good:**
```graphql
type User {
  email: EmailAddress!
  website: URL!
  createdAt: DateTime!
}
```

### Guild-Inspired Rules

### no-unused-fields
Detects fields that are never used or referenced in the schema, following [Guild's no-unused-fields rule](https://the-guild.dev/graphql/eslint/rules/no-unused-fields).

**Bad:**
```graphql
type User {
  id: ID!
  name: String!
  # This field is never used anywhere
  unusedField: String
}
```

**Good:**
```graphql
type User {
  id: ID!
  name: String!
}
```

### strict-id-in-types
Requires output types to have unique identifier fields, following [Guild's strict-id-in-types rule](https://the-guild.dev/graphql/eslint/rules/strict-id-in-types).

**Bad:**
```graphql
type User {
  name: String!
  email: String!
}
```

**Good:**
```graphql
type User {
  id: ID!
  name: String!
  email: String!
}
```

### require-deprecation-reason
Requires meaningful deprecation reasons for deprecated fields, following [Guild's require-deprecation-reason rule](https://the-guild.dev/graphql/eslint/rules/require-deprecation-reason).

**Bad:**
```graphql
type User {
  id: ID!
  name: String! @deprecated
  fullName: String!
}
```

**Good:**
```graphql
type User {
  id: ID!
  name: String! @deprecated(reason: "Use fullName instead")
  fullName: String!
}
```

### no-scalar-result-type-on-mutation
Mutations should return object types instead of scalars, following [Guild's no-scalar-result-type-on-mutation rule](https://the-guild.dev/graphql/eslint/rules/no-scalar-result-type-on-mutation).

**Bad:**
```graphql
type Mutation {
  createUser(input: CreateUserInput!): Boolean
  deleteUser(id: ID!): String
}
```

**Good:**
```graphql
type Mutation {
  createUser(input: CreateUserInput!): CreateUserResult!
  deleteUser(id: ID!): DeleteUserResult!
}

type CreateUserResult {
  user: User
  success: Boolean!
  errors: [String!]!
}
```

### alphabetize
Enforces alphabetical ordering of fields and enum values, following [Guild's alphabetize rule](https://the-guild.dev/graphql/eslint/rules/alphabetize).

**Bad:**
```graphql
type User {
  name: String!
  id: ID!
  email: String!
}

enum Status {
  PENDING
  ACTIVE
  INACTIVE
}
```

**Good:**
```graphql
type User {
  email: String!
  id: ID!
  name: String!
}

enum Status {
  ACTIVE
  INACTIVE
  PENDING
}
```

### input-name
Standardizes mutation input naming conventions, following [Guild's input-name rule](https://the-guild.dev/graphql/eslint/rules/input-name).

**Bad:**
```graphql
type Mutation {
  createUser(data: CreateUserData!): User
  updateUser(id: ID!, name: String): User
}
```

**Good:**
```graphql
type Mutation {
  createUser(input: CreateUserInput!): User
  updateUser(input: UpdateUserInput!): User
}
```

### Additional Comprehensive Rules

### no-unused-types
All declared types must be used somewhere in the schema - custom rule to support Federation.

**Bad:**
```graphql
type User {
  id: ID!
  name: String!
}

# This type is never referenced
type UnusedType {
  value: String!
}
```

**Good:**
```graphql
type User {
  id: ID!
  name: String!
  profile: UserProfile!
}

type UserProfile {
  bio: String!
  avatar: String!
}
```

### capitalized-descriptions
All descriptions must start with a capital letter for consistency.

**Bad:**
```graphql
type User {
  """user identifier"""
  id: ID!
  """user's full name"""
  name: String!
}
```

**Good:**
```graphql
type User {
  """User identifier"""
  id: ID!
  """User's full name"""
  name: String!
}
```

### enum-unknown-case
All enums used in output types must have an UNKNOWN case for future compatibility.

**Bad:**
```graphql
enum UserStatus {
  ACTIVE
  INACTIVE
}

type User {
  status: UserStatus!
}
```

**Good:**
```graphql
enum UserStatus {
  UNKNOWN
  ACTIVE
  INACTIVE
}

type User {
  status: UserStatus!
}
```

### no-query-prefixes
Query fields cannot be prefixed with get/list/find as it's implied by being a query.

**Bad:**
```graphql
type Query {
  getUser(id: ID!): User
  listUsers: [User!]!
  findProducts: [Product!]!
}
```

**Good:**
```graphql
type Query {
  user(id: ID!): User
  users: [User!]!
  products: [Product!]!
}
```

### input-enum-suffix
Input enums must be distinct from output enums and suffixed with 'Input' for clarity.

**Bad:**
```graphql
enum Role {
  USER
  ADMIN
}

input CreateUserInput {
  role: Role!  # Same enum used in input and output
}

type User {
  role: Role!
}
```

**Good:**
```graphql
enum Role {
  UNKNOWN
  USER
  ADMIN
}

enum RoleInput {
  USER
  ADMIN
}

input CreateUserInput {
  role: RoleInput!
}

type User {
  role: Role!
}
```

### enum-descriptions
All enum values must have descriptions except for UNKNOWN case.

**Bad:**
```graphql
enum UserStatus {
  UNKNOWN
  ACTIVE
  INACTIVE
  SUSPENDED
}
```

**Good:**
```graphql
enum UserStatus {
  UNKNOWN
  """User account is active and in good standing"""
  ACTIVE
  """User account is temporarily inactive"""
  INACTIVE
  """User account has been suspended due to violations"""
  SUSPENDED
}
```

### Additional Best Practice Rules

### list-non-null-items
List types should contain non-null items to prevent null pointer issues and improve type safety.

**Bad:**
```graphql
type User {
  friends: [User]
  tags: [String]
}
```

**Good:**
```graphql
type User {
  friends: [User!]!
  tags: [String!]!
}
```

### enum-reserved-values
Enum types should have reserved values for extensibility and future compatibility.

**Bad:**
```graphql
enum Status {
  ACTIVE
  INACTIVE
}
```

**Good:**
```graphql
enum Status {
  UNKNOWN
  RESERVED1
  RESERVED2
  ACTIVE
  INACTIVE
}
```

### mutation-response-nullable
Mutation response fields should be nullable to prevent breaking changes during schema evolution.

**Bad:**
```graphql
type User {
  id: ID!
  name: String!
}

type Mutation {
  createUser(input: CreateUserInput!): User!
}
```

**Good:**
```graphql
type User {
  id: ID
  name: String
}

type Mutation {
  createUser(input: CreateUserInput!): User
}
```

## Custom Rules

You can create custom rules by implementing the `Rule` interface:

```go
package main

import (
    "github.com/nishant-rn/gqlparser/v2/ast"
    "github.com/gqllinter/pkg/linter"
)

type MyCustomRule struct{}

func (r *MyCustomRule) Name() string {
    return "my-custom-rule"
}

func (r *MyCustomRule) Description() string {
    return "Description of what this rule checks"
}

func (r *MyCustomRule) Check(schema *ast.Schema, source *ast.Source) []linter.LintError {
    var errors []linter.LintError
    
    // Your custom validation logic here
    
    return errors
}

// For plugins, export this function
func NewRule() linter.Rule {
    return &MyCustomRule{}
}
```

Compile your custom rule as a plugin:

```bash
go build -buildmode=plugin -o my-rule.so my-rule.go
```

Then use it with the linter:

```bash
gqllinter --custom-rule-paths ./rules/ schema.graphql
```

## Output Format

### Text Format (Default)

```
schema.graphql:5:1: The object type `QueryRoot` is missing a description. (types-have-descriptions)
schema.graphql:6:3: The field `QueryRoot.a` is missing a description. (fields-have-descriptions)
```

### JSON Format

```json
{
  "errors": [
    {
      "message": "The object type `QueryRoot` is missing a description.",
      "location": {
        "line": 5,
        "column": 1,
        "file": "schema.graphql"
      },
      "rule": "types-have-descriptions"
    },
    {
      "message": "The field `QueryRoot.a` is missing a description.",
      "location": {
        "line": 6,
        "column": 3,
        "file": "schema.graphql"
      },
      "rule": "fields-have-descriptions"
    }
  ]
}
```

## Configuration

Create a configuration file to customize the linter behavior:

```yaml
# .gqllinter.yml
rules:
  - types-have-descriptions
  - fields-have-descriptions
  - naming-convention

ignore-patterns:
  - "# gqllinter-ignore"

custom-rules-dir: "./custom-rules"
```

## Integration

### GitHub Actions

```yaml
name: GraphQL Schema Lint
on: [push, pull_request]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - name: Install gqllinter
        run: go install github.com/gqllinter@latest
      - name: Lint GraphQL Schema
        run: gqllinter --format json schema/*.graphql
```

### Pre-commit Hook

```yaml
# .pre-commit-config.yaml
repos:
  - repo: local
    hooks:
      - id: gqllinter
        name: GraphQL Schema Lint
        entry: gqllinter
        language: system
        files: \.graphql$
```

## References

This linter is inspired by and follows guidelines from:

- [Yelp GraphQL Guidelines](https://yelp.github.io/graphql-guidelines/schema-design.html)
- [The Guild GraphQL ESLint Rules](https://the-guild.dev/graphql/eslint/rules)
- [graphql-schema-linter](https://github.com/cjoudrey/graphql-schema-linter)


## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.
