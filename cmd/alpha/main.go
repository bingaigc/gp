package main

import (
	"github.com/bingaigc/gp/internal/interfaces/cli"
)

var (
	Version   = "1.0.0"
	BuildTime = "unknown"
)

func main() {
	cli.Execute()
}
