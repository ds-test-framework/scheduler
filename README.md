# scheduler
Abstract framework to work with intercepted messages of a distributed system. Provides the necessary abstractions to instrument algorithms and define strategies for testing the implementations.

## Build

Run `go install checker.go` while in directory `cmd/checker` for the main binary.

## Run

The runtime has to be specified a config file with name `config.json` which contains json with the following keys
- `run`, specify the configuration related to how many testing iterations need to be run and the timeout for each iteration
- `log` specify the log path if needed
- `engine` specify the strategy that needs to be used to test and any strategy related configuration
- `driver` specify the algo driver that needs to be used and the related driver configuration

Once in `cmd/checker` directory, run `go run checker.go -config <config_dir_path>`. If the binary is built then run `checker -config <config_dir_path>`