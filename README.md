# scheduler
Abstract framework to work with intercepted messages of a distributed system. Provides the necessary abstractions to test consensus algorithm implementations and define strategies for testing the implementations.

The goal is to be able to intercept messages that are exchanged in a distributed system and play with their ordering which will test the algorithm. There are three ways to run the scheduler
- Visualizer: An interactive dashboard with all the messages in the pool visualized in order to manually instruct the order of messages that should be delivered at every replica
- Testlib: A library which can be used to construct unit test cases. Each testcase drives the distributed system to an specific state where we can assert and check its safety
- Strategies: Generic strategies that are ignorant about the details of a specific protocol and use abstract theories to explore different executions of a distributed system

[![Go Reference](https://pkg.go.dev/badge/github.com/ds-test-framework/scheduler.svg)](https://pkg.go.dev/github.com/ds-test-framework/scheduler)

## Documentation

- [Architecture](./docs/arch.md) of the framework
- [Development](./docs/development.md)
- [Testing Server](./docs/testserver.md) - A library to specify test cases for distributed algorithm implementation 