package main

import (
	"fmt"
	"os"

	"github.com/smokyabdulrahman/prayer-times/internal/cli"
)

// version is set at build time via ldflags:
//
//	go build -ldflags "-X main.version=v1.0.0"
var version = "dev"

func main() {
	rootCmd := cli.NewRootCmd(version)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
