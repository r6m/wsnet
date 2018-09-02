package wsnet

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type (
	handleConnectionFunc func(*Connection)
	handleMessageFunc    func(*Connection, []byte) error
	handlePongFunc       func(*Connection)
	handleErrorFunc      func(*Connection, error)
	filterFunc           func(*Connection) bool
)

// Server struct
type Server struct {
	options        *Options
	upgrader       *websocket.Upgrader
	serializer     Serializer
	connectHandler handleConnectionFunc
	messageHandler handleMessageFunc
	pongHandler    handlePongFunc
	closeHandler   handleConnectionFunc
	errorHandler   handleErrorFunc
}

// New returns a server, options is optional to pass
func New(options ...*Options) *Server {
	opts := defaultOptions()
	if len(options) > 0 {
		opts = options[0]
	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize:  opts.ReadBufferSize,
		WriteBufferSize: opts.WriteBufferSize,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return &Server{
		options:        opts,
		upgrader:       upgrader,
		serializer:     NewJSONSerializer(), // default is json
		connectHandler: func(*Connection) {},
		messageHandler: func(*Connection, []byte) error { return nil },
		closeHandler:   func(*Connection) {},
		errorHandler:   func(*Connection, error) {},
	}
}

// HandleRequest should be set to upgrade it to Websocket
func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	// upgrade websocket
	conn, err := s.upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		return err
	}

	connection := newConnection(s, conn)
	s.connectHandler(connection)

	// process wirtes in background
	go connection.writePipe()

	// process read blocks here
	connection.readPipe()

	connection.Close()

	return nil
}

// HandleConnect is called on client connect
func (s *Server) HandleConnect(fn func(*Connection)) {
	s.connectHandler = fn
}

// HandleMessage called on every message
func (s *Server) HandleMessage(fn func(*Connection, []byte) error) {
	s.messageHandler = fn
}

// HandleClose is called when connection closed
func (s *Server) HandleClose(fn func(*Connection)) {
	s.closeHandler = fn
}

// HandleError called if connection send an error
func (s *Server) HandleError(fn func(*Connection, error)) {
	s.errorHandler = fn
}

// HandlePong is called on every pong
func (s *Server) HandlePong(fn func(*Connection)) {
	s.pongHandler = fn
}

// SetSerializer sets the serializer, default is json
func (s *Server) SetSerializer(serializer Serializer) {
	s.serializer = serializer
}
