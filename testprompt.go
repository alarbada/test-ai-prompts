package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type TestCase struct {
	Input    string `json:"input" yaml:"input"`
	Expected string `json:"expected" yaml:"expected"`
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

func runTestWithPrompt(testCase TestCase, promptConfig PromptConfig) bool {
	result, err := callOpenAIWithPrompt(promptConfig, testCase.Input)
	if err != nil {
		fmt.Printf("  ERROR: %v\n", err)
		return false
	}

	passed := strings.EqualFold(strings.TrimSpace(result), strings.TrimSpace(testCase.Expected))

	fmt.Printf("Input: %s\n", testCase.Input)
	fmt.Printf("Expected: %s\n", testCase.Expected)
	fmt.Printf("Got: %s\n", result)
	fmt.Printf("PASSED: %v\n\n", passed)

	return passed
}

func testMain() error {
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	promptFile := testCmd.String("prompt", "", "prompt config file (required)")
	testFile := testCmd.String("testcases", "", "test cases file (required)")

	testCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s test -prompt <file> -testcases <file>\n", os.Args[0])
		testCmd.PrintDefaults()
	}

	if err := testCmd.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *promptFile == "" || *testFile == "" {
		testCmd.Usage()
		return fmt.Errorf("missing required flags: prompt and testcases are required")
	}

	// Load prompt config
	promptConfig, err := loadPromptConfig(*promptFile)
	if err != nil {
		return fmt.Errorf("loading prompt file %s: %w", *promptFile, err)
	}

	// Load test cases
	testCases, err := loadTestCases(*testFile)
	if err != nil {
		return fmt.Errorf("loading test cases: %w", err)
	}

	fmt.Printf("Running %d test cases with prompt from %s:\n\n", len(testCases), *promptFile)

	passed := 0
	for i, testCase := range testCases {
		fmt.Printf("Test %d:\n", i+1)
		if runTestWithPrompt(testCase, promptConfig) {
			passed++
		}
	}

	fmt.Printf("Results: %d/%d tests passed\n", passed, len(testCases))
	return nil
}

func runMain() error {
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	promptFile := runCmd.String("prompt", "", "prompt config file (required)")
	testFile := runCmd.String("testcases", "", "test cases file (required)")
	indexStr := runCmd.String("index", "", "test case index (required)")

	runCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s run -prompt <file> -testcases <file> -index <n>\n", os.Args[0])
		runCmd.PrintDefaults()
	}

	if err := runCmd.Parse(os.Args[2:]); err != nil {
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *promptFile == "" || *testFile == "" || *indexStr == "" {
		runCmd.Usage()
		return fmt.Errorf("missing required flags: prompt, testcases, and index are required")
	}

	index, err := strconv.Atoi(*indexStr)
	if err != nil {
		return fmt.Errorf("invalid index %s: %w", *indexStr, err)
	}

	// Load prompt config
	promptConfig, err := loadPromptConfig(*promptFile)
	if err != nil {
		return fmt.Errorf("loading prompt file %s: %w", *promptFile, err)
	}

	// Load test cases
	testCases, err := loadTestCases(*testFile)
	if err != nil {
		return fmt.Errorf("loading test cases: %w", err)
	}

	if index < 0 || index >= len(testCases) {
		return fmt.Errorf("index %d out of range (0-%d)", index, len(testCases)-1)
	}

	fmt.Printf("Running test case %d:\n", index)
	testCase := testCases[index]

	result := runTestWithPrompt(testCase, promptConfig)
	if result {
		fmt.Printf("✓ Test case %d PASSED\n", index)
	} else {
		fmt.Printf("✗ Test case %d FAILED\n", index)
	}
	return nil
}
