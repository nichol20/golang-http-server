package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	"github.com/nichol20/http-server/internal/request"
	"github.com/nichol20/http-server/internal/response"
	"github.com/nichol20/http-server/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		rt := req.RequestLine.RequestTarget
		switch {
		case rt == "/bad-request":
			serveHTML(w, 400)
			return
		case rt == "/server-error":
			serveHTML(w, 500)
			return
		case strings.HasPrefix(rt, "/httpbin/"):
			httpbinPath := strings.TrimPrefix(rt, "/httpbin/")
			serveChunkedData(w, httpbinPath)
			return
		default:
			serveHTML(w, 200)
			return
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

func serveHTML(w *response.Writer, statusCode int16) {
	fileName := fmt.Sprintf("%d.html", statusCode)
	tplDir := templatesDir()
	body, err := os.ReadFile(filepath.Join(tplDir, fileName))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal("error reading file: ", err)
	}

	header := response.GetDefaultHeaders(len(body))
	header.Replace("Content-Type", "text/html")

	err = w.WriteRespose(statusCode, header, body)
	if err != nil {
		log.Fatal("error writing response message: ", err)
	}
}

func serveChunkedData(w *response.Writer, path string) {
	header := response.GetDefaultHeaders(0)
	header.Del("Content-Length")
	header.Set("Transfer-Encoding", "chunked")

	resp, err := http.Get(fmt.Sprintf("https://httpbin.org/%s", path))
	if err != nil {
		log.Fatal("error getting response from httpbin: ", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Fatalf("upstream returned non-2xx: %d %s", resp.StatusCode, resp.Status)
	}

	log.Println("starting to serve chunked data...")

	err = w.WriteStatusLine(200)
	if err != nil {
		log.Fatal("error writing status line: ", err)
	}
	err = w.WriteHeader(header)
	if err != nil {
		log.Fatal("error writing header: ", err)
	}

	buf := make([]byte, 1024)
	for {
		n, rerr := resp.Body.Read(buf)
		log.Println(n)
		if n > 0 {
			if _, werr := w.WriteChunkedBody(buf[:n]); werr != nil {
				log.Fatalf("error writing chunk to client: %v", werr)
			}
		}

		if rerr != nil {
			if errors.Is(rerr, io.EOF) {
				if _, doneErr := w.WriteChunkedBodyDone(); doneErr != nil {
					log.Fatalf("error writing final chunk done: %v", doneErr)
				}
				return
			}
			log.Fatalf("error reading upstream body: %v", rerr)
		}
	}
}
