package server

import "net"

type Server struct {
}

func Serve(port int) (*Server, error)

func (s *Server) Close() error

func (s *Server) listen()

func (s *Server) handle(conn net.Conn)
