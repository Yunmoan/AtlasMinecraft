//go:build !windows

package server

// Go heap/runtime metrics remain available on all platforms. Native process
// CPU and resident memory are currently collected on Windows only.
func readProcessUsage() processUsage { return processUsage{} }
