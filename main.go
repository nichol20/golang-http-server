package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	relativePath := "./messages.txt"
	absPath, _ := filepath.Abs(relativePath)

	file, err := os.Open(absPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	ch := getLinesChannel(file)

	for str := range ch {
		fmt.Printf("read: %s\n", str)
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
