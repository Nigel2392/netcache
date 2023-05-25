package server

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Nigel2392/netcache/src/cache"
	"github.com/Nigel2392/netcache/src/logger"
	"github.com/Nigel2392/netcache/src/protocols"
)

type CacheServer struct {
	// The cache to use.
	Cache cache.Cache
	// The address to listen on.
	address string
	// The port to listen on.
	port int
	// The timeout for requests.
	timeout time.Duration
	// The logger to use.
	logger logger.Logger
}

// NewCacheServer creates a new cache server.
func New(address string, port int, timeout time.Duration, c cache.Cache) *CacheServer {

	if c == nil {
		c = cache.NewMemoryCache()
	}

	if timeout <= 0 {
		timeout = time.Minute
	}

	var s = &CacheServer{
		Cache:   c,
		address: address,
		port:    port,
		timeout: timeout,
	}

	return s
}

func (s *CacheServer) SavePeriodically(init_file string, interval time.Duration) (stop func()) {
	var t = time.NewTicker(interval)
	go func() {
		var errs int
		var err error
		for range t.C {
			err = s.Save(init_file)
			if err != nil {
				errs++
			}
			if errs > 5 {
				if s.logger != nil {
					s.logger.Critical(errors.New("Too many errors saving cache, exiting."))
				}
				os.Exit(1)
			}
		}
	}()
	return t.Stop
}

// SaveOnInterrupt saves the cache when the program is interrupted.
func (s *CacheServer) SaveOnInterrupt(init_file string) {
	var c = make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		s.Save(init_file)
		os.Exit(0)
	}()
}

func (s *CacheServer) Save(init_file string) error {
	if s.logger != nil {
		s.logger.Debug("Saving cache...")
	}
	var b, err = s.Cache.Dump()
	if err != nil {
		if s.logger != nil {
			s.logger.Critical(fmt.Errorf("Error dumping cache: %s", err))
		}
		return err
	}
	if s.logger != nil {
		s.logger.Debug("Writing cache...")
	}

	var dir = filepath.Dir(init_file)
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		if s.logger != nil {
			s.logger.Debug("Creating directory...")
		}
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
	}
	var f *os.File
	if f, err = os.OpenFile(init_file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666); err != nil {
		return err
	}
	if err != nil {
		if s.logger != nil {
			s.logger.Critical(fmt.Errorf("Error opening init file: %s", err))
		}
		return err
	}
	defer f.Close()
	_, err = f.Write(b)
	if err != nil {
		if s.logger != nil {
			s.logger.Critical(fmt.Errorf("Error writing init file: %s", err))
		}
	}
	return err
}

// Load loads the cache from the init file.
func (s *CacheServer) Load(init_file string) error {
	var f *os.File
	var err error

	if s.logger != nil {
		s.logger.Debug("Reading init file...")
	}

	if f, err = os.Open(init_file); err != nil {
		if s.logger != nil {
			s.logger.Critical(fmt.Errorf("Error opening init file: %s", err))
		}
		return err
	}

	var b bytes.Buffer
	if _, err = b.ReadFrom(f); err != nil {
		if s.logger != nil {
			s.logger.Critical(fmt.Errorf("Error reading init file: %s", err))
		}
		return err
	}

	if s.logger != nil {
		s.logger.Debug("Loading cache...")
	}

	err = s.Cache.Load(b.Bytes())
	if err != nil {
		if s.logger != nil {
			s.logger.Critical(fmt.Errorf("Error loading cache: %s", err))
		}
		return err
	}
	return nil
}

// NewLogger creates a new logger for the server.
func (s *CacheServer) NewLogger(logger logger.Logger) {
	s.logger = logger
}

// Function to run before the server starts to listen.
func (s *CacheServer) preInit() {
	if s.logger != nil {
		s.logger.Info("Starting cache...")
	}
	s.Cache.Run(time.Minute / 2)
	if s.logger != nil {
		s.logger.Infof("Listening on %s:%d\n", s.address, s.port)
	}
}

// ListenAndServe starts the server.
func (s *CacheServer) ListenAndServe() error {
	s.preInit()
	var l, err = net.Listen("tcp", s.address+":"+strconv.Itoa(s.port))
	if err != nil {
		return err
	}
	defer l.Close()
	if s.logger != nil {
		s.logger.Info("Waiting for connections...")
	}
	return s.listen(l)
}

// ListenAndServeTLS starts the server with TLS.
func (s *CacheServer) ListenAndServeTLS(conf *tls.Config) error {
	s.preInit()
	var l, err = tls.Listen("tcp", s.address+":"+strconv.Itoa(s.port), conf)
	if err != nil {
		return err
	}
	defer l.Close()
	if s.logger != nil {
		s.logger.Info("Waiting for TLS connections...")
	}
	return s.listen(l)
}

// Function to run to listen for connections.
func (s *CacheServer) listen(l net.Listener) error {
	for {
		var c, err = l.Accept()
		if err != nil {
			if s.logger != nil {
				s.logger.Errorf("Error accepting connection: %s (%s)\n", err, c.RemoteAddr().String())
			}
			return err
		}

		if s.logger != nil {
			s.logger.Infof("Connection from %s\n", c.RemoteAddr().String())
		}
		go s.handle(c)
	}
}

func (s *CacheServer) handle(c net.Conn) {
	for {
		var message = new(protocols.Message)
		if s.logger != nil {
			s.logger.Debug("Waiting for message...")
		}
		_, err := message.ReadFrom(c)
		if err != nil {
			if s.logger != nil {
				s.logger.Warningf("Error reading message: %s, disconnecting. (%s)\n", err, c.RemoteAddr().String())
			}
			return
		}
		// Handle the message with a set timeout.
		err = runWithTimeout(func() error {
			switch message.Type {
			case protocols.TypeGET:
				if s.logger != nil {
					s.logger.Debugf("Received GET request for key %s\n", message.Key)
				}
				err = s.handleGet(c, message)
			case protocols.TypeSET:
				if s.logger != nil {
					s.logger.Debugf("Received SET request for key %s\n", message.Key)
				}
				err = s.handleSet(c, message)
			case protocols.TypeDELETE:
				if s.logger != nil {
					s.logger.Debugf("Received DELETE request for key %s\n", message.Key)
				}
				err = s.handleDelete(c, message)
			case protocols.TypeCLEAR:
				if s.logger != nil {
					s.logger.Debug("Received CLEAR request")
				}
				err = s.handleClear(c)
			case protocols.TypeHAS:
				if s.logger != nil {
					s.logger.Debugf("Received HAS request for key %s\n", message.Key)
				}
				err = s.handleHas(c, message)
			case protocols.TypePING:
				if s.logger != nil {
					s.logger.Debug("Received PING request")
				}
				err = s.handlePing(c)
			case protocols.TypeKEYS:
				if s.logger != nil {
					s.logger.Debug("Received KEYS request")
				}
				err = s.handleKeys(c)
			}
			if err != nil {
				err = writeErrorMessage(c, err)
				if err != nil {
					if s.logger != nil {
						s.logger.Warningf("Error writing error message: %s, disconnecting. (%s)\n", err, c.RemoteAddr().String())
					}
					return err
				}
				return nil
			}
			s.logger.Debug("Writing end message...")
			err = protocols.WriteEnd(c)
			if err != nil {
				if s.logger != nil {
					s.logger.Warningf("Error writing end message: %s, disconnecting. (%s)\n", err, c.RemoteAddr().String())
				}
				return err
			}
			return nil
		}, s.timeout)
		if err != nil {
			return
		}
	}
}

func runWithTimeout(f func() error, timeout time.Duration) error {
	if timeout <= 0 {
		f()
		return nil
	}
	var done = make(chan error)
	defer close(done)
	go func() {
		done <- f()
	}()
	select {
	case <-time.After(timeout):
		return errors.New("timeout")
	case err := <-done:
		return err
	}
}
