package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/nichol20/http-server/internal/request"
)

type Server struct {
	Addr     string
	listener net.Listener
	closed   *atomic.Bool
}

func Serve(port uint16) (*Server, error) {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	closed := &atomic.Bool{}
	closed.Store(false)
	s := &Server{Addr: addr, listener: listener, closed: closed}

	go s.listen()

	return s, nil
}

func (s *Server) Close() error {
	fmt.Println("server closed!")
	s.closed.Store(false)
	return s.listener.Close()
}

func (s *Server) listen() {
	for !s.closed.Load() {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Fatal("Error accepting connection: ", err)
		}

		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	_, err := request.RequestFromReader(conn)
	if err != nil {
		log.Fatal("Error parsing request: ", err)
	}

	res := "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello World!"
	conn.Write([]byte(res))

	err = conn.Close()
	if err != nil {
		log.Fatal("error closing connection: ", err)
	}
}
