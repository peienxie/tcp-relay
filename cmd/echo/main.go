package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

var (
	PORT = 8089
)

type client struct {
	chMessage chan []byte
	reader    *bufio.Reader
	writer    *bufio.Writer
}

func handleConnection(conn net.Conn) {
	client := client{
		chMessage: make(chan []byte),
		writer:    bufio.NewWriter(conn),
		reader:    bufio.NewReader(conn),
	}
	go client.read()
	go client.write()
}

func (c *client) read() {
	buffer := make([]byte, 4096)
	for {
		n, err := c.reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Println("read error:", err)
			}
			break
		}
		log.Println("readed", string(buffer[:n]))
		c.chMessage <- buffer[:n]
	}
}

func (c *client) write() {
	for data := range c.chMessage {
		reply := []byte("echo " + string(data))
		_, err := c.writer.Write(reply)
		if err != nil {
			log.Println("write error:", err)
			break
		}
		log.Println("written", string(reply))
		c.writer.Flush()
	}
}

func main() {
	conn, err := net.Dial("tcp4", fmt.Sprintf(":%d", PORT))
	if err != nil {
		log.Fatal(err)
	}
	handleConnection(conn)
	done := make(chan bool, 1)
	<-done
}
