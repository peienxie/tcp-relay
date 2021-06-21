package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
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
		// reply := []byte("echo " + string(data))
		reply := data
		_, err := c.writer.Write(reply)
		if err != nil {
			log.Println("write error:", err)
			break
		}
		log.Println("written", string(reply))
		c.writer.Flush()
	}
}

var (
	remoteIPAddr *string
	remotePortNum *int
	localPortNum *int
)

var (
	errInvalidCommand = fmt.Errorf("unknwon command, should be either listen or connect")
	errInvliadPortNumber = fmt.Errorf("invalid port number provided")
	errInvalidIPAddress = fmt.Errorf("invalid ip address provided")
)

func main() {
	listenArgs := flag.NewFlagSet("listen", flag.ExitOnError)
	connectArgs := flag.NewFlagSet("connect", flag.ExitOnError)
	
	localPortNum = listenArgs.Int("port", 0, "local host port number would be listening on")

	remoteIPAddr = connectArgs.String("ip", "", "remote host IP address")
	remotePortNum = connectArgs.Int("port", 0, "remote host port number")
	
    if len(os.Args) < 2 {
		log.Fatal(errInvalidCommand)
    }
	switch os.Args[1] {
    case "listen":
        listenArgs.Parse(os.Args[2:])
    case "connect":
        connectArgs.Parse(os.Args[2:])
    default:
		log.Fatal(errInvalidCommand)
    }
	
	if listenArgs.Parsed() {
		if *localPortNum == 0 {
			log.Fatal(errInvliadPortNumber)
		}
		addr := fmt.Sprintf(":%d", *localPortNum)
		l, err := net.Listen("tcp4", addr)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("server is listening on address %s\n", addr)
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatal(err)
			}
			go handleConnection(conn)
		}
	}
	
	if connectArgs.Parsed() {
		if *remoteIPAddr == "" {
			connectArgs.PrintDefaults()
			log.Fatalf("%s:%+v\n", errInvalidIPAddress, *remoteIPAddr)
		}
		if *remotePortNum == 0 {
			connectArgs.PrintDefaults()
			log.Fatalf("%s:%+v\n", errInvliadPortNumber, *remoteIPAddr)
		}
		addr := fmt.Sprintf("%s:%d", *remoteIPAddr, *remotePortNum)
		conn, err := net.Dial("tcp4", addr)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("successfully connect to %s\n", addr)
		handleConnection(conn)
		for {
		}
	}
}
