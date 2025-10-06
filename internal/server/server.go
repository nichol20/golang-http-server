package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/nichol20/http-server/internal/request"
	"github.com/nichol20/http-server/internal/response"
)

type Server struct {
	Addr     string
	listener *net.Listener
	closed   *atomic.Bool
	handler  Handler
}

type HandlerError struct {
	StatusCode int16
	Message    string
}

type Handler func(w *response.Writer, req *request.Request)

func Serve(port uint16, handler Handler) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	closed := &atomic.Bool{}
	closed.Store(false)
	s := &Server{Addr: addr, listener: &listener, closed: closed, handler: handler}

	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	fmt.Println("server closed!")
	s.closed.Store(false)
	return (*s.listener).Close()
}

func (s *Server) listen() {
	for {
		conn, err := (*s.listener).Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	writer := response.NewWriter(conn)
	req, err := request.RequestFromReader(conn)
	if err != nil {
		header := response.GetDefaultHeaders(len(err.Error()))
		err = writer.WriteRespose(400, header, []byte(err.Error()))
		if err != nil {
			log.Fatal("error writing response: ", err)
		}
		return
	}

	s.handler(writer, req)

	err = conn.Close()
	if err != nil {
		log.Fatal("error closing connection: ", err)
	}
}
