package linter

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"plugin"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"

	"github.com/gqllinter/pkg/rules"
	"github.com/gqllinter/pkg/types"
)

// Linter holds the configuration and rules for linting GraphQL schemas
type Linter struct {
	rules        map[string]types.Rule
	enabledRules []string
}

// New creates a new linter instance with default rules
func New() *Linter {
	l := &Linter{
		rules: make(map[string]types.Rule),
	}

	// Register built-in rules
	l.registerBuiltinRules()

	return l
}

// registerBuiltinRules adds all the built-in rules to the linter
func (l *Linter) registerBuiltinRules() {
	builtinRules := []types.Rule{
		rules.NewTypesHaveDescriptions(),
		rules.NewFieldsHaveDescriptions(),
		rules.NewNoHashtagDescription(),
		rules.NewNamingConvention(),
	}

	for _, rule := range builtinRules {
		l.rules[rule.Name()] = rule
	}
}

// LoadCustomRules loads custom rules from a directory
func (l *Linter) LoadCustomRules(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.so"))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := l.loadCustomRule(file); err != nil {
			// Log warning but don't fail completely
			fmt.Printf("Warning: failed to load custom rule %s: %v\n", file, err)
		}
	}

	return nil
}

// loadCustomRule loads a single custom rule from a shared library
func (l *Linter) loadCustomRule(path string) error {
	p, err := plugin.Open(path)
	if err != nil {
		return err
	}

	symbol, err := p.Lookup("NewRule")
	if err != nil {
		return err
	}

	newRuleFunc, ok := symbol.(func() types.Rule)
	if !ok {
		return fmt.Errorf("invalid NewRule function signature")
	}

	rule := newRuleFunc()
	l.rules[rule.Name()] = rule

	return nil
}

// SetRules configures which rules should be enabled
func (l *Linter) SetRules(ruleNames []string) {
	l.enabledRules = ruleNames
}

// LintFile lints a single GraphQL schema file
func (l *Linter) LintFile(filename string) ([]types.LintError, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return l.LintString(string(content), filename)
}

// LintString lints GraphQL schema content
func (l *Linter) LintString(content, filename string) ([]types.LintError, error) {
	source := &ast.Source{
		Name:    filename,
		Input:   content,
		BuiltIn: false,
	}

	// Parse the schema
	_, gqlErr := parser.ParseSchema(source)
	if gqlErr != nil {
		// Convert parser errors to lint errors
		var errors []types.LintError

		// Handle gqlerror.List
		if gqlErrList, ok := gqlErr.(interface{ GetErrors() []error }); ok {
			for _, err := range gqlErrList.GetErrors() {
				errors = append(errors, types.LintError{
					Message: err.Error(),
					Location: types.Location{
						Line:   1,
						Column: 1,
						File:   filename,
					},
					Rule: "syntax-error",
				})
			}
		} else {
			errors = append(errors, types.LintError{
				Message: gqlErr.Error(),
				Location: types.Location{
					Line:   1,
					Column: 1,
					File:   filename,
				},
				Rule: "syntax-error",
			})
		}
		return errors, nil
	}

	// Build schema
	schema, err := gqlparser.LoadSchema(source)
	if err != nil {
		return nil, fmt.Errorf("failed to build schema: %w", err)
	}

	// Run enabled rules
	var allErrors []types.LintError
	rulesToRun := l.getRulesToRun()

	for _, ruleName := range rulesToRun {
		if rule, exists := l.rules[ruleName]; exists {
			errors := rule.Check(schema, source)
			allErrors = append(allErrors, errors...)
		}
	}

	return allErrors, nil
}

// getRulesToRun returns the list of rules that should be executed
func (l *Linter) getRulesToRun() []string {
	if len(l.enabledRules) > 0 {
		return l.enabledRules
	}

	// Return all available rules if none specified
	var ruleNames []string
	for name := range l.rules {
		ruleNames = append(ruleNames, name)
	}

	return ruleNames
}

// GetAvailableRules returns all available rules
func (l *Linter) GetAvailableRules() map[string]types.Rule {
	return l.rules
}
