# AGENT.md - AI Prompt Testing Tool

## Build/Test Commands
- `go run . test <prompt.yaml> <testcases.json>` - Run tests with custom prompt
- `go run . generate <prompt.yaml> <testcases.json> [num_cases]` - Generate test cases
- `go build` - Build the application
- `go mod tidy` - Update dependencies

## Architecture
This is an OpenAI prompt testing tool with two main functions:
1. **Test Runner**: Tests prompts against predefined test cases using OpenAI API
2. **Test Generator**: Generates new test cases based on existing prompt configurations

### Key Components
- `main.go`: Core test runner and CLI interface
- `generate.go`: Test case generation functionality
- `input/`: Contains prompt configuration files (YAML/JSON format)
- `tests/`: Contains test case files (JSON arrays)
- `schemas/`: JSON schemas for validation (OpenAI chat completion format)

## Code Style & Conventions
- Standard Go formatting (`gofmt`)
- Struct tags use both `json` and `yaml` for dual format support
- Error handling: Return errors, use `log.Fatal` for critical failures
- OpenAI API integration via `github.com/sashabaranov/go-openai`
- Environment: Requires `OPENAI_API_KEY` env var (auto-loaded via godotenv)
- File formats: Supports both YAML (.yaml/.yml) and JSON for prompts
- Test cases: Always JSON format with `input`/`expected` structure
