package main

import (
	"flag"
	"io"
	"os"
	"time"

	"github.com/Nigel2392/netcache/src/cache"
	"github.com/Nigel2392/netcache/src/logger"
	"github.com/Nigel2392/netcache/src/server"
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
}

func init() {
	flag.StringVar(&flags.address, "address", "127.0.0.1", "The address to listen on.")
	flag.IntVar(&flags.port, "port", 8080, "The port to listen on.")
	flag.StringVar(&flags.cacheDir, "cache-dir", "./cache", "The directory to store the cache in.")
	flag.IntVar(&flags.timeout, "timeout", 5, "The timeout for requests in seconds.")
	flag.StringVar(&flags.logfile, "logfile", "", "The logfile to write to (none for stdout).")
	flag.StringVar(&flags.loglevel, "loglevel", "INFO", "The log level to use. (\"CRITICAL\", \"ERROR\", \"WARNING\", \"INFO\", \"DEBUG\", \"TEST\")")
	flag.BoolVar(&flags.memcache, "memory", false, "Use an in-memory cache.")
	flag.Parse()
}

func main() {
	var c cache.Cache
	if flags.memcache {
		c = cache.NewMemoryCache()
	} else {
		c = cache.NewFileCache(flags.cacheDir)
	}

	var server = server.New(flags.address, flags.port, flags.cacheDir, time.Duration(flags.timeout)*time.Second, c)
	var std io.Writer
	var err error
	if flags.logfile != "" {
		std, err = logger.NewLogFile(flags.logfile)
	} else {
		std = os.Stdout
	}
	server.NewLogger(
		logger.Newlogger(logger.LoglevelFromString(flags.loglevel), std),
	)
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
