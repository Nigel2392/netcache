//go:build docker
// +build docker

package main

import (
	"os"
	"strconv"
)

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
	var (
		err1, err2, err3 error
	)
	flags.address = "127.0.0.1"
	flags.port, err1 = strconv.Atoi(os.Getenv("PORT"))
	flags.cacheDir = os.Getenv("CACHE_DIR")
	flags.timeout, err2 = strconv.Atoi(os.Getenv("TIMEOUT"))
	flags.logfile = os.Getenv("LOGFILE")
	flags.loglevel = os.Getenv("LOGLEVEL")
	flags.memcache, err3 = strconv.ParseBool(os.Getenv("MEMCACHE"))

	if err1 != nil || err2 != nil || err3 != nil {
		panic("Invalid environment variables")
	}
}
