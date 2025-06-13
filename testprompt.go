package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
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

func testMain() {
	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	promptFile := testCmd.String("prompt", "", "prompt config file (required)")
	testFile := testCmd.String("testcases", "", "test cases file (required)")

	testCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s test -prompt <file> -testcases <file>\n", os.Args[0])
		testCmd.PrintDefaults()
	}

	testCmd.Parse(os.Args[2:])

	if *promptFile == "" || *testFile == "" {
		testCmd.Usage()
		os.Exit(1)
	}

	// Load prompt config
	promptConfig, err := loadPromptConfig(*promptFile)
	if err != nil {
		fmt.Printf("Error loading prompt file %s: %v\n", *promptFile, err)
		os.Exit(1)
	}

	// Load test cases
	testData, err := os.ReadFile(*testFile)
	if err != nil {
		fmt.Printf("Error reading test file %s: %v\n", *testFile, err)
		os.Exit(1)
	}

	var testCases []TestCase
	if err := json.Unmarshal(testData, &testCases); err != nil {
		fmt.Printf("Error parsing test cases: %v\n", err)
		os.Exit(1)
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
}
