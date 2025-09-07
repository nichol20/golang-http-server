package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		ch := getLinesChannel(conn)
		for str := range ch {
			fmt.Printf("read: %s\n", str)
		}

		conn.Close()
		fmt.Println("A connection has been closed!")
	}

}

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string, 1)

	go func() {
		currentLine := ""
		buf := make([]byte, 8)

		for {
			n, err := f.Read(buf)
			if err != nil && err != io.EOF {
				log.Fatal("Error reading file: ", err)
			}

			// nothing more to read
			if n == 0 {
				if len(currentLine) > 0 {
					ch <- currentLine
				}
				close(ch)
				break
			}

			parts := strings.Split(string(buf[:n]), "\n")
			currentLine += parts[0]

			for len(parts) > 1 {
				parts = parts[1:]
				ch <- currentLine
				currentLine = ""
				currentLine += parts[0]
			}
		}
	}()

	return ch
}
