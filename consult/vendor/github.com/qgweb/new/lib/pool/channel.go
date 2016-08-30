package pool

import (
	"errors"
	"fmt"
	"sync"
)

// channelPool implements the Pool interface based on buffered channels.
type channelPool struct {
	// storage for our net.Conn connections
	mu    sync.Mutex
	conns chan interface{}

	// net.Conn generator
	newfactory   NewFactory
	closefactory CloseFactory
}

// Factory is a function to create new connections.
type NewFactory func() (interface{}, error)
type CloseFactory func(conn interface{}) error

// NewChannelPool returns a new pool based on buffered channels with an initial
// capacity and maximum capacity. Factory is used when initial capacity is
// greater than zero to fill the pool. A zero initialCap doesn't fill the Pool
// until a new Get() is called. During a Get(), If there is no new connection
// available in the pool, a new connection will be created via the Factory()
// method.
func NewChannelPool( maxCap int, newfactory NewFactory, closefactory CloseFactory) (Pool, error) {
	c := &channelPool{
		conns:        make(chan interface{}, maxCap),
		newfactory:   newfactory,
		closefactory: closefactory,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < maxCap; i++ {
		conn, err := newfactory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.conns <- conn
	}

	return c, nil
}

func (c *channelPool) getConns() chan interface{} {
	c.mu.Lock()
	conns := c.conns
	c.mu.Unlock()
	return conns
}

// Get implements the Pool interfaces Get() method. If there is no new
// connection available in the pool, a new connection will be created via the
// Factory() method.
func (c *channelPool) Get() (interface{}, error) {
	conns := c.getConns()
	if conns == nil {
		return nil, ErrClosed
	}

	// wrap our connections with out custom net.Conn implementation (wrapConn
	// method) that puts the connection back to the pool if it's closed.
	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}

		return conn, nil
//	default:
//		conn, err := c.newfactory()
//		if err != nil {
//			return nil, err
//		}
//
//		return conn, nil
	}
}

// put puts the connection back to the pool. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (c *channelPool) Put(conn interface{}) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conns == nil {
		// pool is closed, close passed connection
		return c.closefactory(conn)
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case c.conns <- conn:
		return nil
	default:
		// pool is full, close passed connection
		return c.closefactory(conn)
	}
}

func (c *channelPool) Close() {
	c.mu.Lock()
	conns := c.conns
	c.conns = nil
	c.newfactory = nil
	c.mu.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for conn := range conns {
		c.closefactory(conn)
	}
}

func (c *channelPool) Len() int { return len(c.getConns()) }
