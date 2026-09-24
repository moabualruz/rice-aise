ci:
    go vet ./...
    go test ./... -count=1 -cover
    go test ./tests/integration/... -tags integration -count=1
    go test ./tests/e2e/... -tags e2e -count=1
    go build -o raise ./cmd/raise
    ./raise version
