package sip

import (
	"log"
	"net"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) Start() error {
	udpAddr, err := net.ResolveUDPAddr("udp", s.addr)
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}

	log.Println("SIP Server Listening on", s.addr)

	buf := make([]byte, 65535)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("Read Error: ", err)
			continue
		}

		go s.handlePacket(conn, remoteAddr, buf[:n])
	}
}

func (s *Server) handlePacket(
	conn *net.UDPConn,
	remote *net.UDPAddr,
	data []byte,
) {
	msg := ParseMessage(data)

	switch msg.Method {
	case "REGISTER":
		resp := BuildResponse(200, "OK", msg)
		conn.WriteToUDP([]byte(resp), remote)

	case "INVITE":
		conn.WriteToUDP(
			[]byte(BuildResponse(100, "Trying", msg)),
			remote,
		)
		conn.WriteToUDP(
			[]byte(BuildResponse(180, "Ringing", msg)),
			remote,
		)
		conn.WriteToUDP(
			[]byte(BuildResponse(200, "OK", msg)),
			remote,
		)
	case "ACK":
		// nothing
	case "BYE":
		resp := BuildResponse(200, "OK", msg)
		conn.WriteToUDP([]byte(resp), remote)
	}
}
