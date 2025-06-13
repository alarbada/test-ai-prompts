test PROMPT TESTCASES:
    go run . test {{PROMPT}} {{TESTCASES}}

# Run a single test case by index
run PROMPT TESTCASES INDEX:
    go run . run -prompt {{PROMPT}} -testcases {{TESTCASES}} -index {{INDEX}}

# Generate test cases
generate PROMPT TESTCASES NUM_CASES:
    just build
    tmp/main generate -prompt {{PROMPT}} -testcases {{TESTCASES}} -num {{NUM_CASES}}

build:
    go build -o tmp/main .

tidy:
    go mod tidy

fmt:
    gofmt -w .

# Run golangci-lint to check code quality
lint:
    golangci-lint run

# Fix auto-fixable linting issues
lint-fix:
    golangci-lint run --fix

# Load environment variables and run a command (example: just env test ...)
env *ARGS:
    dotenv -f .env -- just {{ARGS}}