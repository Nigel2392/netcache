package server

import (
	"errors"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/Nigel2392/netcache/src/cache"
	"github.com/Nigel2392/netcache/src/internal/protocols"
	"github.com/Nigel2392/netcache/src/logger"
)

type CacheServer struct {
	// The cache to use.
	cache cache.Cache
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
func New(address string, port int, cacheDir string, timeout time.Duration, c cache.Cache) *CacheServer {
	var logger = logger.Newlogger(logger.INFO, os.Stdout)
	if c == nil {
		c = cache.NewFileCache(cacheDir)
	}
	return &CacheServer{
		cache:   c,
		address: address,
		port:    port,
		timeout: timeout,
		logger:  logger,
	}
}

// NewLogger creates a new logger for the server.
func (s *CacheServer) NewLogger(logger logger.Logger) error {
	if s.logger != nil {
		s.logger = logger
	}
	return nil
}

// ListenAndServe starts the server.
func (s *CacheServer) ListenAndServe() error {
	if s.logger != nil {
		s.logger.Info("Starting cache...")
	}
	s.cache.Run(time.Minute / 2)
	if s.logger != nil {
		s.logger.Infof("Listening on %s:%d\n", s.address, s.port)
	}
	var l, err = net.Listen("tcp", s.address+":"+strconv.Itoa(s.port))
	if err != nil {
		return err
	}
	defer l.Close()
	if s.logger != nil {
		s.logger.Info("Waiting for connection...")
	}
	for {
		var c, err = l.Accept()
		if err != nil {
			if s.logger != nil {
				s.logger.Warningf("Error accepting connection: %s (%s)\n", err, c.RemoteAddr().String())
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
