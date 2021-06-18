# Development

## Setup
```go install .``` installs the required dependencies. Ensure that the codebase is in `$GOPATH/github.com/ds-test-framework/scheduler`

[`main.go`](../main.go) is the entry-point and it runs the [`cmd/checker`](../cmd/checker/main.go) command.

## Run
```go run main.go --config <path_to_config_dir>```

## Build and run
```go build -o build/checker```
```cd build && ./checker --config <path_to_config_dir>```