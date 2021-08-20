package main

import "github.com/ds-test-framework/scheduler/cmd"

func main() {
	c := cmd.RootCmd()
	c.Execute()
}
