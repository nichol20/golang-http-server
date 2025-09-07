package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

// use "nc -u -l 42069" to listen for UDP packets
func main() {
	remoteAddr, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal("Error resolving upd address: ", err)
	}

	conn, err := net.DialUDP("udp", nil, remoteAddr)
	if err != nil {
		log.Fatal("Error dialing UDP:", err)
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading user input: ", err)
		}

		_, err = conn.Write([]byte(str[:len(str)-1]))
		if err != nil {
			fmt.Println("Error sending data to UDP connection: ", err)
		}
	}
}
