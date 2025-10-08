package main

import (
	"crypto/sha256"
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

	"github.com/nichol20/http-server/internal/header"
	"github.com/nichol20/http-server/internal/request"
	"github.com/nichol20/http-server/internal/response"
	"github.com/nichol20/http-server/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, func(w *response.Writer, req *request.Request) {
		rt := req.RequestLine.RequestTarget
		method := req.RequestLine.Method
		if method != "GET" {
			serveHTML(w, 200)
			return
		}

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
		case rt == "/video":
			serveVideo(w)
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

func thisFile() string {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("runtime.Caller failed")
	}

	return thisFile
}

func templatesDir() string {
	root := filepath.Join(filepath.Dir(thisFile()), "..", "..")
	templates := filepath.Join(root, "internal", "response", "templates")
	abs, err := filepath.Abs(templates)
	if err != nil {
		log.Fatalf("Abs: %v", err)
	}
	return abs
}

func assetsDir() string {
	root := filepath.Join(filepath.Dir(thisFile()), "..", "..")
	assets := filepath.Join(root, "assets")
	abs, err := filepath.Abs(assets)
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

// echo -e "GET /httpbin/stream/100 HTTP/1.1\r\nHost: localhost:42069\r\nConnection: close\r\n\r\n" | nc localhost 42069
func serveChunkedData(w *response.Writer, path string) {
	hdr := response.GetDefaultHeaders(0)
	hdr.Del("Content-Length")
	hdr.Set("Transfer-Encoding", "chunked")
	hdr.Set("Trailer", "X-Content-SHA256")
	hdr.Set("Trailer", "X-Content-Length")

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
	err = w.WriteHeader(hdr)
	if err != nil {
		log.Fatal("error writing header: ", err)
	}

	buf := make([]byte, 1024)
	body := []byte{}
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
			if _, werr := w.WriteChunkedBody(buf[:n]); werr != nil {
				log.Fatalf("error writing chunk to client: %v", werr)
			}
		}

		if rerr != nil {
			if errors.Is(rerr, io.EOF) {
				if _, doneErr := w.WriteChunkedBodyDone(); doneErr != nil {
					log.Fatalf("error writing final chunk done: %v", doneErr)
				}

				sum := sha256.Sum256(body)
				trailer := header.Header{}
				trailer["X-Content-SHA256"] = fmt.Sprintf("%x", sum)
				trailer["X-Content-Length"] = fmt.Sprintf("%d", len(body))

				if terr := w.WriteTrailer(trailer); terr != nil {
					log.Fatalf("error writing final chunk done: %v", terr)
				}
				return
			}
			log.Fatalf("error reading upstream body: %v", rerr)
		}
	}
}

func serveVideo(w *response.Writer) {
	f, err := os.Open(filepath.Join(assetsDir(), "video.mp4"))
	if err != nil {
		log.Printf("error opening video file: %v", err)
		serveHTML(w, 500)
		return
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		log.Printf("error stating video file: %v", err)
		serveHTML(w, 500)
		return
	}
	size := fi.Size()

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		log.Printf("error seeking video file: %v", err)
		serveHTML(w, 500)
		return
	}

	hdr := response.GetDefaultHeaders(int(size))
	hdr.Replace("Content-Type", "video/mp4")
	hdr.Set("Accept-Ranges", "bytes")

	if err := w.WriteStatusLine(200); err != nil {
		log.Printf("error writing status line for video: %v", err)
		return
	}
	if err := w.WriteHeader(hdr); err != nil {
		log.Printf("error writing headers for video: %v", err)
		return
	}

	buf := make([]byte, 32*1024) // 32KB
	for {
		n, rerr := f.Read(buf)
		if n > 0 {
			if _, werr := w.WriteBody(buf[:n]); werr != nil {
				log.Printf("error writing video chunk: %v", werr)
				return
			}
		}
		if rerr != nil {
			if errors.Is(rerr, io.EOF) {
				log.Println("finished streaming video")
				return
			}
			log.Printf("error reading video file: %v", rerr)
			return
		}
	}
}
