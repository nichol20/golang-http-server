package main

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nichol20/http-server/internal/request"
	"github.com/nichol20/http-server/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w io.Writer, req *request.Request) *server.HandlerError {
		switch req.RequestLine.RequestTarget {
		case "/bad-request":
			return &server.HandlerError{
				StatusCode: 400,
				Message:    "Hello from Bad Request Endpoint",
			}
		case "/server-error":
			return &server.HandlerError{
				StatusCode: 500,
				Message:    "Hello from Server Error Endpoint",
			}
		default:
			_, err := w.Write([]byte("Hello, 世界"))
			if err != nil {
				log.Fatal("error writing response message: ", err)
			}
			return nil
		}
	})

	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
