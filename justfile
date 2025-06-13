test PROMPT TESTCASES:
    go run . test {{PROMPT}} {{TESTCASES}}

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

# Load environment variables and run a command (example: just env test ...)
env *ARGS:
    dotenv -f .env -- just {{ARGS}}