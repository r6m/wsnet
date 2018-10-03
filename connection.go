package wsnet

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

var connID uint32

// Connection is a wrapper for Conn
type Connection struct {
	sync.Mutex
	values         *sync.Map
	id             uint32
	server         *Server
	conn           *websocket.Conn
	outgoingCh     chan []byte
	outgoingStopCh chan struct{}
	stopped        bool
}

func newConnection(server *Server, conn *websocket.Conn) *Connection {
	atomic.AddUint32(&connID, 1)
	return &Connection{
		id:             connID,
		values:         &sync.Map{},
		server:         server,
		conn:           conn,
		outgoingCh:     make(chan []byte, server.options.OutgoinSize),
		outgoingStopCh: make(chan struct{}),
		stopped:        false,
	}
}

// ID returns the connection id
func (c *Connection) ID() uint32 { return c.id }

// Values returns connection values
func (c *Connection) Values() *sync.Map { return c.values }

// Send marshals the interface using serializer
func (c *Connection) Send(i interface{}) error {
	bytes, err := c.server.serializer.Marshal(i)
	if err != nil {
		return err
	}

	c.SendBytes(bytes)

	return nil
}

// SendBytes send raw byte data
func (c *Connection) SendBytes(bytes []byte) {
	c.Lock()
	if c.stopped {
		c.Unlock()
		return
	}

	select {
	case c.outgoingCh <- bytes:
		c.Unlock()
		return
	default:
		c.Unlock()
		log.Println("could not write message, outgoing queue full")
	}
}

func (c *Connection) readPipe() {
	defer c.Close()

	// read loop
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				if e, ok := err.(*net.OpError); !ok || e.Err.Error() != "use of closed network connection" {
					log.Println("Error reading message from client", err.Error())
				}
			}
			break
		}

		// pass data to message handler
		if err := c.server.messageHandler(c, data); err != nil {
			log.Println("Error message handler:", err.Error())
			break
		}
	}
}

func (c *Connection) writePipe() {
	for {
		select {
		case <-c.outgoingStopCh:
			// Connection is closing, close the outgoingCh process routine.
			return
		case data := <-c.outgoingCh:
			c.Lock()
			if c.stopped {
				// the connection might be stopped while payload being send to queue
				c.Unlock()
				return
			}

			// process outgoing
			c.conn.SetWriteDeadline(time.Now().Add(time.Duration(c.server.options.WriteDeadline) * time.Millisecond))
			err := c.conn.WriteMessage(websocket.BinaryMessage, data)
			if err != nil {
				c.Unlock()
				log.Println("could not write message: ", err.Error())
				return
			}
			c.Unlock()
		}
	}
}

// Close closes the connection, also stopes the channels
func (c *Connection) Close() {
	c.Lock()
	if c.stopped { // already closed?
		c.Unlock()
		return
	}

	// stop
	c.stopped = true
	c.Unlock()

	close(c.outgoingStopCh)
	close(c.outgoingCh)

	// handle close
	c.server.closeHandler(c)

	// send close message to client
	err := c.conn.WriteControl(websocket.CloseMessage, []byte{}, time.Now().Add(time.Duration(c.server.options.WriteDeadline)*time.Millisecond))
	if err != nil {
		log.Println("Could not send close message, closing prematurely", c.conn.RemoteAddr().String(), err.Error())
	}

	// close connection
	c.conn.Close()
	log.Println("connected closed id:", c.id)
}
