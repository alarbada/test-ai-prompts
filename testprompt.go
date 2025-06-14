package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"
)

type TestCase struct {
	Input    string `json:"input" yaml:"input"`
	Expected string `json:"expected" yaml:"expected"`
}

// Evaluator defines the interface for different evaluation strategies
type Evaluator interface {
	Evaluate(expected, actual string) (bool, error)
}

// StrictEvaluator performs case-insensitive string comparison
type StrictEvaluator struct{}

func (e StrictEvaluator) Evaluate(expected, actual string) (bool, error) {
	return strings.EqualFold(strings.TrimSpace(expected), strings.TrimSpace(actual)), nil
}

// JSONEvaluator parses and compares JSON structures
type JSONEvaluator struct{}

func (e JSONEvaluator) Evaluate(expected, actual string) (bool, error) {
	var expectedJSON, actualJSON any

	if err := json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		return false, fmt.Errorf("failed to parse expected JSON: %w", err)
	}

	if err := json.Unmarshal([]byte(actual), &actualJSON); err != nil {
		return false, fmt.Errorf("failed to parse actual JSON: %w", err)
	}

	if cmp.Equal(expectedJSON, actualJSON) {
		return true, nil
	}

	diff := cmp.Diff(expectedJSON, actualJSON)
	return false, fmt.Errorf("JSON mismatch:\n%s", diff)
}

func loadTestCases(filename string) ([]TestCase, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var testCases []TestCase
	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &testCases)
		if err != nil {
			return nil, fmt.Errorf("failed to parse YAML test cases: %w", err)
		}
	case ".json":
		err = json.Unmarshal(data, &testCases)
		if err != nil {
			return nil, fmt.Errorf("failed to parse JSON test cases: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported file format: %s (use .json, .yaml, or .yml)", ext)
	}

	return testCases, nil
}

func runTestWithPrompt(testCase TestCase, promptConfig PromptConfig, evaluator Evaluator) bool {
	result, err := callOpenAIWithPrompt(promptConfig, testCase.Input)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return false
	}

	fmt.Printf("Input: %s\n", testCase.Input)
	fmt.Printf("Expected: %s\n", testCase.Expected)
	fmt.Printf("Got: %s\n", result)

	passed, err := evaluator.Evaluate(testCase.Expected, result)
	if err != nil {
		fmt.Printf("EVAL ERROR: %v\n\n", err)
		return false
	}

	fmt.Printf("PASSED: %v\n\n", passed)

	return passed
}

func evalMain(evalFile string) error {
	// Load eval config
	evalConfig, err := loadEvalConfig(evalFile)
	if err != nil {
		return fmt.Errorf("loading eval file %s: %w", evalFile, err)
	}

	fmt.Printf("Running evaluation: %s\n", evalConfig.Name)
	fmt.Printf("Found %d test suites\n\n", len(evalConfig.Tests))

	totalPassed := 0
	totalTests := 0

	for _, test := range evalConfig.Tests {
		fmt.Printf("=== Running %s ===\n", test.Name)

		// Load prompt config
		promptConfig, err := loadPromptConfig(test.Prompt)
		if err != nil {
			return fmt.Errorf("loading prompt file %s: %w", test.Prompt, err)
		}

		// Load test cases
		testCases, err := loadTestCases(test.Samples)
		if err != nil {
			return fmt.Errorf("loading test cases %s: %w", test.Samples, err)
		}

		// Create evaluator based on type
		var evaluator Evaluator
		evalType := test.EvalType
		if evalType == "" {
			evalType = "strict"
		}

		switch evalType {
		case "json":
			evaluator = JSONEvaluator{}
		case "strict":
			evaluator = StrictEvaluator{}
		default:
			return fmt.Errorf("unknown evaluation type: %s (use 'strict' or 'json')", evalType)
		}

		fmt.Printf("Running %d test cases (eval: %s):\n\n", len(testCases), evalType)

		passed := 0
		for i, testCase := range testCases {
			fmt.Printf("Test %d:\n", i+1)
			if runTestWithPrompt(testCase, promptConfig, evaluator) {
				passed++
			}
		}

		fmt.Printf("Suite Results: %d/%d tests passed\n", passed, len(testCases))
		fmt.Printf("=== %s Complete ===\n\n", test.Name)

		totalPassed += passed
		totalTests += len(testCases)
	}

	fmt.Printf("Overall Results: %d/%d tests passed\n", totalPassed, totalTests)
	return nil
}
