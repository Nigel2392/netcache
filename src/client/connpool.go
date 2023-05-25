package client

import (
	"net"
	"sync"
	"time"

	"github.com/Nigel2392/go-datastructures/stack"
)

func init() {
	// here to make sure the cache client implements the cache interface
	var _ = Cache(&CacheClient{})
}

type connectionPool struct {
	// The address of the server.
	ServerAddr string

	// the connection pool
	pool *stack.Stack[net.Conn]

	// mutex
	mu sync.Mutex
}

func newPool(serverAddr string, connections int) (*connectionPool, error) {

	var p = &connectionPool{
		ServerAddr: serverAddr,
		pool:       &stack.Stack[net.Conn]{},
	}

	for i := 0; i < connections; i++ {
		var conn, err = net.Dial("tcp", serverAddr)
		if err != nil {
			return nil, err
		}
		p.pool.Push(conn)
	}

	return p, nil
}

// get a connection from the pool
func (p *connectionPool) get(deadline time.Duration) net.Conn {
	if deadline == 0 {
		deadline = 5 * time.Second
	}

	if p == nil {
		return nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	var conn, ok = p.pool.PopOK()
	if !ok {
		var c, okChan = p.pool.PopOKDeadline(deadline / 2)
		select {
		case <-okChan:
			return nil
		case conn = <-c:
			deadline = deadline / 2
		}
	}

	if err := conn.SetDeadline(time.Now().Add(deadline)); err != nil {
		return nil
	}

	return conn
}

func (p *connectionPool) put(conn net.Conn) {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pool.Push(conn)
}
