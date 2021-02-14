package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {

	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}
	toBeSent := make(chan string)
	receivedMessages := make(chan string)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			text, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err.Error())
				break
			}
			toBeSent <- text
		}
	}()

	go func() {
		byteSlice := make([]byte, 1024)
		for {
			readBytes, err := bufio.NewReader(conn).Read(byteSlice)
			if err != nil {
				log.Fatal(err.Error())
			}
			receivedMessages <- string(byteSlice[:readBytes])
		}
	}()

	for {
		select {
		case receivedMsg := <-receivedMessages:
			go func() {
				fmt.Println(string(receivedMsg))
			}()
		case toSend := <-toBeSent:
			go func() {
				conn.Write([]byte(toSend))
			}()
		}
	}

}
