package main

import (
	"fmt"
	"os"

	"github.com/ds-test-framework/scheduler/cmd/checker"
)

func main() {
	cmd := checker.CheckerCommand()
	if err := cmd.Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
