package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nichol20/http-server/internal/request"
	"github.com/nichol20/http-server/internal/response"
	"github.com/nichol20/http-server/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		var msg string
		var statusCode int16

		switch req.RequestLine.RequestTarget {
		case "/bad-request":
			msg = "Hello from Bad Request Endpoint"
			statusCode = 400
		case "/server-error":
			msg = "Hello from Server Error Endpoint"
			statusCode = 500
		default:
			msg = "Hello, 世界"
			statusCode = 200
		}

		err := w.WriteRespose(statusCode, response.GetDefaultHeaders(len(msg)), []byte(msg))
		if err != nil {
			log.Fatal("error writing response message: ", err)
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
