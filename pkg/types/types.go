package types

import (
	"github.com/nishant-rn/gqlparser/v2/ast"
)

// LintError represents a linting error with location information
type LintError struct {
	Message  string   `json:"message"`
	Location Location `json:"location"`
	Rule     string   `json:"rule"`
}

// Location represents the position of an error in a file
type Location struct {
	Line   int    `json:"line"`
	Column int    `json:"column"`
	File   string `json:"file"`
}

// Rule interface that all linting rules must implement
type Rule interface {
	// Name returns the unique identifier for this rule
	Name() string

	// Description returns a human-readable description of what this rule checks
	Description() string

	// Check validates the schema and returns any errors found
	Check(schema *ast.Schema, source *ast.Source) []LintError
}
