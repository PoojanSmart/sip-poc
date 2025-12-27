package sip

import (
	"log"
	"net"
	"sip-poc/internal/registrar"
	"sync"
	"time"
)

type Transaction struct {
	Caller *net.TCPAddr
	Callee *net.TCPAddr
}

type Server struct {
	addr         string
	registrar    *registrar.Store
	txMu         sync.Mutex
	transactions map[string]*Transaction
}

func NewServer(addr string) *Server {
	return &Server{
		addr:         addr,
		registrar:    registrar.NewSore(),
		transactions: make(map[string]*Transaction),
	}
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
	log.Printf(
		"\n===================== SIP MESSSAGE RECEIVED =====================================\n"+
			"FROM: %s\n"+
			"SIZE: %d bytes\n"+
			"RAW: \n%s\n"+
			"====================================================================================\n",
		remote.String(),
		len(data),
		string(data),
	)

	msg := ParseMessage(data)

	switch msg.Method {
	case "REGISTER":
		user := ExtractUser(msg.Headers["To"])
		ttl := 300 * time.Second

		s.registrar.Save(user, remote, ttl)

		resp := BuildResponse(200, "OK", msg)

		conn.WriteToUDP([]byte(resp), remote)

	case "INVITE":
		callee := ExtractUser(msg.Headers["To"])
		reg, ok := s.registrar.Get(callee)

		if !ok {
			resp := BuildResponse(404, "Not Found", msg)
			conn.WriteToUDP([]byte(resp), remote)
			log.Printf(
				"\n===================== SIP MESSSAGE SENT =====================================\n"+
					"SIZE: %d bytes\n"+
					"RAW: \n%s\n"+
					"====================================================================================\n",
				len(resp),
				string(resp),
			)
			return
		}

		callID := msg.Headers["Call-ID"]

		s.txMu.Lock()
		s.transactions[callID] = &Transaction{
			Caller: (*net.TCPAddr)(remote),
			Callee: (*net.TCPAddr)(reg.Addr),
		}
		s.txMu.Unlock()

		msg = ParseMessage(data)

		msg.Headers["Via"] = buildProxyVia("127.0.0.1", 5060)

		resp := []byte(BuildResponse(msg.StatusCode, msg.Method, msg))

		log.Printf(
			"\n===================== SIP MESSSAGE SENT =====================================\n"+
				"SIZE: %d bytes\n"+
				"RAW: \n%s\n"+
				"====================================================================================\n",
			len(resp),
			string(resp),
		)

		conn.WriteToUDP(resp, reg.Addr)

		/*conn.WriteToUDP(
			[]byte(BuildResponse(100, "TRYING", msg)),
			remote,
		)*/
		/*conn.WriteToUDP(
			[]byte(BuildResponse(180, "Ringing", msg)),
			remote,
		)
		conn.WriteToUDP(
			[]byte(BuildResponse(200, "OK", msg)),
			remote,
		)*/
	case "ACK":
		// nothing
	case "BYE":
		resp := BuildResponse(200, "OK", msg)
		conn.WriteToUDP([]byte(resp), remote)
	case "RESPONSE":
		callID := msg.Headers["Call-ID"]
		s.txMu.Lock()
		transactionsAddr, ok := s.transactions[callID]
		s.txMu.Unlock()

		if ok {
			msg = ParseMessage(data)

			msg.Headers["Via"] = buildProxyVia("127.0.0.1", 5060)

			resp := []byte(BuildResponse(msg.StatusCode, msg.Method, msg))

			log.Printf(
				"\n===================== SIP MESSSAGE SENT =====================================\n"+
					"SIZE: %d bytes\n"+
					"RAW: \n%s\n"+
					"====================================================================================\n",
				len(resp),
				string(resp),
			)

			conn.WriteToUDP(resp, (*net.UDPAddr)(transactionsAddr.Caller))
		}
	case "Trying":
		callID := msg.Headers["Call-ID"]

		s.txMu.Lock()
		transactionsAddr, ok := s.transactions[callID]
		s.txMu.Unlock()

		if !ok {
			resp := BuildResponse(404, "Not Found", msg)
			conn.WriteToUDP([]byte(resp), remote)
			return
		}

		msg = ParseMessage(data)

		msg.Headers["Via"] = buildProxyVia("127.0.0.1", 5060)

		resp := []byte(BuildResponse(msg.StatusCode, msg.Method, msg))

		log.Printf(
			"\n===================== SIP MESSSAGE SENT =====================================\n"+
				"SIZE: %d bytes\n"+
				"RAW: \n%s\n"+
				"====================================================================================\n",
			len(resp),
			string(resp),
		)

		conn.WriteToUDP(resp, (*net.UDPAddr)(transactionsAddr.Caller))
	}
}
