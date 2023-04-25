//go:build !docker
// +build !docker

package main

import "flag"

var flags struct {
	// The address to listen on.
	address string
	// The port to listen on.
	port int
	// The directory to store the cache in.
	cacheDir string
	// The timeout for requests.
	timeout int
	//The logfile to write to.
	logfile string
	// Log level.
	loglevel string
	// Use an in-memory cache.
	memcache bool
	// Use the built-in cli
	cli bool
}

func init() {
	flag.StringVar(&flags.address, "address", "127.0.0.1", "The address to listen on.")
	flag.IntVar(&flags.port, "port", 8080, "The port to listen on.")
	flag.StringVar(&flags.cacheDir, "cache-dir", "./cache", "The directory to store the cache in.")
	flag.IntVar(&flags.timeout, "timeout", 5, "The timeout for requests in seconds.")
	flag.StringVar(&flags.logfile, "logfile", "", "The logfile to write to (none for stdout).")
	flag.StringVar(&flags.loglevel, "loglevel", "INFO", "The log level to use. (\"CRITICAL\", \"ERROR\", \"WARNING\", \"INFO\", \"DEBUG\", \"TEST\")")
	flag.BoolVar(&flags.memcache, "memory", false, "Use an in-memory cache.")
	flag.BoolVar(&flags.cli, "cli", false, "Use the built-in cli.")
	flag.Parse()
}
