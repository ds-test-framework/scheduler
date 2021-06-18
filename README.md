# scheduler
Abstract framework to work with intercepted messages of a distributed system. Provides the necessary abstractions to test consensus algorithm implementations and define strategies for testing the implementations.

The goal is to be able to intercept messages that are exchanged in a distributed system and feed it to a strategy that decides the ordering or messages that need to be delivered. The strategy can then choose to reorder/drop or change the contents of the messages in order to test the distributed system

[![Go Reference](https://pkg.go.dev/badge/github.com/ds-test-framework/scheduler.svg)](https://pkg.go.dev/github.com/ds-test-framework/scheduler)

## Documentation

- [Architecture](./docs/arch.md) of the framework
- [Development](./docs/development.md)
- [Testing Server](./docs/testserver.md) - A library to specify test cases for distributed algorithm implementation 