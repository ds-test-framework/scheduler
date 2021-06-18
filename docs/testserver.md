# [TestServer](../testing/server.go)

The `Checker` module runs the same Strategy over the specified number of runs. Combine that with the generic nature of the strategies, any testing that is specific to the algorithm implementation becomes hard to instrument and will rely on extending the scheduler. `testing.Server` serves as a library which embodies the logic of `Checker`.

`Server` is initialized with an array of `TestCase`s. In each run the corresponding `TestCase` is used. `TestCase` here serves the purpose of a strategy but with the goal of exploring a specific execution expecting a certain predefined outcome. This library can be used to specify scenarios that are specific to an algorithm implementation that can be used to test later improvements on the implementation.

`Server` makes used of a custom `Driver` and the `APIServer` to reuse most of the code of the scheduler framework in order to facilitate specifying testing scenarios through the interface `TestCase`

## `TestCase` 
The interface consists of:
- `Initialize` which is called with the `ReplicaStore` containing information of all the replicas and should return a `TestCaseCtx`
- `HandleMessage` invoked for every message that is intercepted, should return true if the message can be allowed to be delivered. Can add new messages as a seconds return value.
- `HandleStateUpdate` is called when a replica posts an update to its internal state
- `HandleLogMessage` is invoked when a replica send a log message
- `Assert` is called at the end of the testcase run and should return `nil` if the expected outcome was achieved, the respective error otherwise.
- `Name` returns the name unique to this testcase. 

`TestCaseCtx` contains a channel which indicates that the testcase is ready to assert. Also contains a timeout value which is used to stop the run of the current testcase

`BaseTestCase` and `BaseTestCaseCtx` serve as abstract implementations of the two interfaces