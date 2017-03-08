package main

import (
	"fmt"
	"runtime"
)

// values set with linker flags
// don't you dare modifying this values!
var version string
var buildDate string
var commit string

func printVersion() {
	fmt.Println("sunoKB Version: " + version)
	fmt.Println("UTC Build Time: " + buildDate)
	fmt.Println("Git Commit Hash: " + commit)
	fmt.Println("Go Version: " + runtime.Version())
	fmt.Printf("Go OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
