package main

import (
	"fmt"
	"log"
	"net"

	"github.com/nichol20/http-server/internal/request"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		log.Fatal("Error creating network listener: ", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error accepting connection: ", err)
		}
		fmt.Println("A connection has been established!")

		req, err := request.RequestFromReader(conn)
		if err != nil {
			log.Fatal("Error parsing request: ", err)
		}

		fmt.Println("Request line:")
		fmt.Printf(
			"- Method: %s\n- Target: %s\n- Version: %s\n",
			req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion,
		)
		fmt.Println("Header:")
		for key, value := range req.Header {
			fmt.Printf("- %s: %s\n", key, value)
		}
		fmt.Println("Body:")
		fmt.Println(string(req.Body))

		conn.Close()
		fmt.Println("A connection has been closed!")
	}

}
