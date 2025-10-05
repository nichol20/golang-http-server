package server

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"

	"github.com/nichol20/http-server/internal/header"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

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
	req, err := request.RequestFromReader(conn)
	if err != nil {
		header := response.GetDefaultHeaders(len(err.Error()))
		err = s.writeMessage(conn, 400, header, []byte(err.Error()))
		if err != nil {
			log.Fatal("error writing response: ", err)
		}
		return
	}

	var b bytes.Buffer

	handlerErr := s.handler(&b, req)

	if handlerErr != nil {
		header := response.GetDefaultHeaders(len(handlerErr.Message))
		err := s.writeMessage(conn, handlerErr.StatusCode, header, []byte(handlerErr.Message))
		if err != nil {
			log.Fatal("error writing handler error: ", err)
		}
	} else {
		header := response.GetDefaultHeaders(b.Len())
		err := s.writeMessage(conn, 200, header, b.Bytes())
		if err != nil {
			log.Fatal("error writing response: ", err)
		}
	}

	err = conn.Close()
	if err != nil {
		log.Fatal("error closing connection: ", err)
	}
}

func (s *Server) writeMessage(w io.Writer, statusCode int16, header header.Header, message []byte) error {
	err := response.WriteStatusLine(w, statusCode)
	if err != nil {
		return fmt.Errorf("error writing status line: %w", err)
	}
	err = response.WriteHeader(w, header)
	if err != nil {
		return fmt.Errorf("error writing headers: %w", err)
	}
	_, err = w.Write(message)
	if err != nil {
		return fmt.Errorf("error writing message: %w", err)
	}
	return nil
}
