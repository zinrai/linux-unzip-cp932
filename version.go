package main

import "fmt"

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func printVersion() {
	fmt.Printf("linux-unzip-cp932 %s (commit %s, built %s)\n", version, commit, date)
}
