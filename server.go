package tcprelay

import (
	"bufio"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"tcprelay/relaytarget"
	"time"
)

type tcpRelayServer struct {
	addr   string
	target relaytarget.TcpRelayTarget
	tlsCfg *tls.Config
}

func NewTcpRelayServer(port int, target relaytarget.TcpRelayTarget, tlsCfg *tls.Config) *tcpRelayServer {
	return &tcpRelayServer{
		addr:   fmt.Sprintf(":%d", port),
		target: target,
		tlsCfg: tlsCfg,
	}
}

func (s *tcpRelayServer) listener() (l net.Listener, err error) {
	if s.tlsCfg != nil {
		l, err = tls.Listen("tcp", s.addr, s.tlsCfg)
		log.Printf("using TLS connection\n")
	} else {
		l, err = net.Listen("tcp4", s.addr)
	}
	if err != nil {
		return nil, err
	}
	log.Printf("middle server is listening on address %s\n", s.addr)
	return l, nil
}

func (s *tcpRelayServer) Listen() {
	l, err := s.listener()
	if err != nil {
		log.Fatal(err)
	}
	err = s.target.Prepare()
	if err != nil {
		log.Printf("target server is not ready: %+v", err)
		return
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go handleConnection(client, s.target)
	}
}

func handleConnection(client net.Conn, target relaytarget.TcpRelayTarget) {
	defer func() {
		log.Printf("connections terminated\n\n")
		client.Close()
	}()
	log.Printf("client connected from %s\n", client.RemoteAddr().String())

	err := target.Dial()
	if err != nil {
		log.Printf("can't dial target server: %+v\n", err)
		return
	}
	log.Printf("successfully dial target server %s\n\n", target.Conn().RemoteAddr().String())

	for {
		var data []byte
		data, err = copy(target.Conn(), client)
		if err != nil {
			log.Printf("error when send data by client: %+v\n", err)
			return
		}
		log.Printf("%s ==========> %s\n", client.RemoteAddr().String(), target.Conn().RemoteAddr().String())
		log.Printf("transmitted packet length: %d\n%s\n\n", len(data), hex.EncodeToString(data))

		data, err = copy(client, target.Conn())
		if err != nil {
			log.Printf("error when send data back to client: %+v\n", err)
			return
		}

		log.Printf("%s <========== %s\n", client.RemoteAddr().String(), target.Conn().RemoteAddr().String())
		log.Printf("received packet length: %d\n%s\n\n", len(data), hex.EncodeToString(data))
	}
}

func copy(dst net.Conn, src net.Conn) (writtenData []byte, err error) {
	src.SetReadDeadline(time.Now().Add(time.Second * 30))

	r := bufio.NewReader(src)
	w := bufio.NewWriter(dst)
	buf := make([]byte, 1024)

	buf[0], err = r.ReadByte()
	if err != nil {
		return writtenData, fmt.Errorf("read first byte error: %+v", err)
	}
	err = w.WriteByte(buf[0])
	if err != nil {
		return writtenData, fmt.Errorf("write first byte error: %+v", err)
	}
	writtenData = append(writtenData, buf[0])

	for r.Buffered() > 0 {
		nr, er := r.Read(buf[:])
		if nr > 0 {
			nw, ew := w.Write(buf[:nr])
			if ew != nil {
				err = fmt.Errorf("write data error: %+v", ew)
				break
			}
			writtenData = append(writtenData, buf[:nw]...)
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = fmt.Errorf("read data error: %+v", er)
			}
			break
		}
	}
	w.Flush()
	return writtenData, err
}
