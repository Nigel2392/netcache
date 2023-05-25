//go:build docker
// +build docker

package main

import (
	"fmt"
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

	// The init file to use.
	initFile string
	// period to save cache in milliseconds
	savePeriod int
	// save on interrupt
	saveOnInterrupt bool
}

func setup() {
	var (
		err1, err2, err3 error
	)
	flags.address = "0.0.0.0"
	flags.port, err1 = strconv.Atoi(getEnv("PORT", "2392"))
	flags.cacheDir = getEnv("CACHE_DIR", "/netcache/cache")
	flags.timeout, err2 = strconv.Atoi(getEnv("TIMEOUT", "60"))
	flags.logfile = getEnv("LOGFILE")
	flags.loglevel = getEnv("LOGLEVEL", "INFO")
	flags.memcache, err3 = strconv.ParseBool(getEnv("MEMCACHE", "false"))
	flags.cli = false
	flags.initFile = getEnv("INIT_FILE", "/netcache/init.netcache")
	flags.savePeriod, _ = strconv.Atoi(getEnv("SAVE_PERIOD", "500"))
	flags.saveOnInterrupt, _ = strconv.ParseBool(getEnv("SAVE_ON_INTERRUPT", "false"))

	if err1 != nil || err2 != nil || err3 != nil {
		panic("Invalid environment variables")
	}

	if flags.savePeriod == 0 {
		fmt.Println("Save period is set to 0, cache will not be saved periodically.")
	}

	if flags.savePeriod < 0 {
		flags.savePeriod = 500
	}

}

func getEnv(key string, def ...string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}
