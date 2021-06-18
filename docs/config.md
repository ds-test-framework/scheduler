# Configuration

Configuration should be specified in a json file and the parent directory should be given as a command line argument when one runs the project.

Configuration keys

key | desc | default 
--- | ---- | -------
`run.runs` | The number of runs that the scheduler should run | 1000 | 
`run.time` | Timeout value for each run (in seconds) | 5
`log.path` | Path to file to which logs should be appended | `stdout`
`engine.type` | Strategy that should be employed to test | 
`driver.type` | Driver that should be employed | 
`transport.addr` | host:port of the API server | `0.0.0.0:7074`


## Driver config
There can be additional configuration that can be specific to the driver. In the case of `common` driver. You should additionally specify
- `driver.algo` which is used to instantiate a workload injector specific to that algorithm that is being tested
- `driver.num_replicas` to indicate the number of replicas that are used in the test cluster
