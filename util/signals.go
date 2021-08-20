package util

import (
	"os"
	"os/signal"
	"syscall"
)

func Term() chan os.Signal {
	termCh := make(chan os.Signal, 1)
	signal.Notify(termCh, os.Interrupt, syscall.SIGTERM)
	return termCh
}
