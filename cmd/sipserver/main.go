package main

import (
	"log"
	"sip-poc/internal/sip"
)

func main() {
	server := sip.NewServer(":5060")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
