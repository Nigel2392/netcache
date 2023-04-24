package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Nigel2392/netcache/src/cache"
	"github.com/Nigel2392/netcache/src/client"
	"github.com/Nigel2392/netcache/src/logger"
	"github.com/Nigel2392/netcache/src/protocols"
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
	// Use the built-in cli
	cli bool
	// The cli-serializer to use.
	cliSerializer string
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
	flag.StringVar(&flags.cliSerializer, "cli-serializer", "", "The cli-serializer to use. (\"json\", \"gob\", \"xml\", \"\")")
	flag.Parse()
}

func main() {

	if flags.cli {
		startCLI()
		return
	}

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

func startCLI() {
	var client = client.New(fmt.Sprintf("%s:%d", flags.address, flags.port), getSerializer())
	var err = client.Connect()
	if err != nil {
		panic(err)
	}
	var cmd string
	printHelp()
	fmt.Println()
	for {
		fmt.Printf("%snetcache> %s", logger.Purple, logger.Reset)
		_, err = fmt.Scanln(&cmd)
		if err != nil {
			fmt.Println(err)
			continue
		}
		switch strings.ToLower(cmd) {
		case "quit", "exit", "q", "leave":
			return
		case "get":
			var key string
			fmt.Printf("%skey> %s", logger.Blue, logger.Reset)
			_, err := fmt.Scanln(&key)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%s%s%s\n", logger.Green, "GETTING", logger.Reset)
			value, err := client.Get(key, nil)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%s%s%s\n", logger.Green, value.Value(), logger.Reset)
		case "set":
			var (
				key   string
				value string
			)
			fmt.Printf("%skey> %s", logger.Blue, logger.Reset)
			_, err := fmt.Scanln(&key)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%svalue> %s", logger.Blue, logger.Reset)
			_, err = fmt.Scanln(&value)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Printf("%sttl> %s", logger.Blue, logger.Reset)
			var ttl int
			_, err = fmt.Scanln(&ttl)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = client.Set(key, value, int64(time.Duration(ttl)*time.Second))
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%s%s%s\n", logger.Green, "OK", logger.Reset)
		case "delete":
			var key string
			fmt.Printf("%skey> %s", logger.Blue, logger.Reset)
			_, err := fmt.Scanln(&key)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = client.Delete(key)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%s%s%s\n", logger.Green, "OK", logger.Reset)
		case "clear":
			err := client.Clear()
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%s%s%s\n", logger.Green, "OK", logger.Reset)
		case "keys":
			keys, err := client.Keys()
			if err != nil {
				fmt.Println(err)
				continue
			}
			for _, key := range keys {
				fmt.Printf("%s%s%s\n", logger.Green, key, logger.Reset)
			}
		case "help":
			printHelp()
		default:
			fmt.Printf("%sUnknown command. Type \"help\" for a list of commands.%s\n", logger.Red, logger.Reset)
		}
	}
}

func printHelp() {
	fmt.Println(logger.Purple + "netcache - Available Commands" + logger.Reset)
	fmt.Printf("\t%sget%s    args: [KEY]\n", logger.Green, logger.Reset)
	fmt.Printf("\t%sset%s    args: [KEY, VALUE, TTL]\n", logger.Green, logger.Reset)
	fmt.Printf("\t%sdelete%s args: [KEY]\n", logger.Green, logger.Reset)
	fmt.Printf("\t%sclear%s\n", logger.Green, logger.Reset)
	fmt.Printf("\t%skeys%s\n", logger.Green, logger.Reset)
	fmt.Printf("\t%shelp%s\n", logger.Green, logger.Reset)
	fmt.Printf("\t%squit%s\n", logger.Green, logger.Reset)
}

func getSerializer() protocols.Serializer {
	switch strings.ToLower(flags.cliSerializer) {
	case "json":
		return &protocols.JsonSerializer{}
	case "gob":
		return &protocols.GobSerializer{}
	case "xml":
		return &protocols.XmlSerializer{}
	}
	return nil
}
