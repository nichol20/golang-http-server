package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/nichol20/http-server/internal/request"
	"github.com/nichol20/http-server/internal/response"
	"github.com/nichol20/http-server/internal/server"
)

const port = 42069

func templatesDir() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("runtime.Caller failed")
	}
	root := filepath.Join(filepath.Dir(thisFile), "..", "..")
	templates := filepath.Join(root, "internal", "response", "templates")
	abs, err := filepath.Abs(templates)
	if err != nil {
		log.Fatalf("Abs: %v", err)
	}
	return abs
}

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		var statusCode int16

		switch req.RequestLine.RequestTarget {
		case "/bad-request":
			statusCode = 400
		case "/server-error":
			statusCode = 500
		default:
			statusCode = 200
		}

		fileName := fmt.Sprintf("%d.html", statusCode)
		tplDir := templatesDir()
		body, err := os.ReadFile(filepath.Join(tplDir, fileName))
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Fatal("error reading file: ", err)
		}

		log.Println(tplDir)
		log.Println(err)

		header := response.GetDefaultHeaders(len(body))
		header.Replace("Content-Type", "text/html")

		err = w.WriteRespose(statusCode, header, body)
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
