# Prompt Testing Tool

## Build/Test Commands
To build code and check for errors use `just build`.
You can also use any of the commands at `./justfile`.

## Architecture
This is an OpenAI prompt testing tool with two main functions:
1. **Test Runner**: Tests prompts against predefined test cases using OpenAI API
2. **Test Generator**: Generates new test cases based on existing prompt configurations

## Code Style & Conventions
- Standard Go formatting (`gofmt`)
- Try to use fmt.Errorf when returning errors to add more context
- Struct tags use both `json` and `yaml` for dual format support
- Error handling: Return errors, use `log.Fatal` for critical failures
- OpenAI API integration via `github.com/sashabaranov/go-openai`
- Environment: Requires `OPENAI_API_KEY` env var (auto-loaded via godotenv)
- File formats: Supports both YAML (.yaml/.yml) and JSON for prompts
- Test cases: Always JSON format with `input`/`expected` structure
