package main

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Nigel2392/netcache/src/cache"
	"github.com/Nigel2392/netcache/src/client"
	"github.com/Nigel2392/netcache/src/logger"
	"github.com/Nigel2392/netcache/src/server"
)

func main() {
	setup()

	if flags.cli {
		startCLI()
		return
	}

	if flags.cacheDir == "" {
		flags.cacheDir = "./netcache-data"
	}

	var c cache.Cache
	if flags.memcache {
		c = cache.NewMemoryCache()
	} else {
		c = cache.NewFileCache(flags.cacheDir)
	}

	var shouldLoad bool
	if flags.initFile == "" {
		flags.initFile = fmt.Sprintf("%s/init.netcache", flags.cacheDir)
	} else {
		shouldLoad = true
	}

	var savePeriod time.Duration
	if flags.savePeriod > 0 {
		savePeriod = time.Duration(flags.savePeriod) * time.Millisecond
	}

	var server = server.New(flags.address, flags.port, time.Duration(flags.timeout)*time.Second, c)
	var std io.Writer
	var err error
	if flags.logfile != "" {
		std, err = logger.NewLogFile(flags.logfile)
	} else {
		std = os.Stdout
	}
	var logger = logger.Newlogger(logger.LoglevelFromString(flags.loglevel), std)
	server.NewLogger(
		logger,
	)

	dumpFlags(logger)

	if shouldLoad {
		err = server.Load(flags.initFile)
		if err != nil {
			logger.Error(err)
		}
	}

	if flags.saveOnInterrupt {
		server.SaveOnInterrupt(flags.initFile)
	}

	if savePeriod > 0 {
		server.SavePeriodically(flags.initFile, savePeriod)
	}

	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func dumpFlags(logger logger.Logger) {
	logger.Info("Flags:")
	logger.Infof("  Address: %s\n", flags.address)
	logger.Infof("  Port: %d\n", flags.port)
	logger.Infof("  CacheDir: %s\n", flags.cacheDir)
	logger.Infof("  Timeout: %d\n", flags.timeout)
	logger.Infof("  LogLevel: %s\n", flags.loglevel)
	logger.Infof("  LogFile: %s\n", flags.logfile)
	logger.Infof("  Memcache: %t\n", flags.memcache)
	logger.Infof("  InitFile: %s\n", flags.initFile)
	logger.Infof("  SavePeriod: %d\n", flags.savePeriod)
	logger.Infof("  SaveOnInterrupt: %t\n", flags.saveOnInterrupt)
}

func startCLI() {
	var client = client.New(fmt.Sprintf("%s:%d", flags.address, flags.port), nil, time.Duration(flags.timeout)*time.Second, 10)
	var err = client.Connect()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
			fmt.Printf("%s%s%s\n", logger.Green, "GOODBYE...", logger.Reset)
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
			err = client.Set(key, value, time.Duration(ttl)*time.Second)
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
