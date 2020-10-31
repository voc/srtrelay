package main

import (
	"log"

	"github.com/haivision/srtgo"
	"github.com/voc/srtrelay/server"
)

func main() {
	options := make(map[string]string)
	options["blocking"] = "0"
	options["transtype"] = "live"
	options["latency"] = "300"

	address := "0.0.0.0"
	port := uint16(8090)
	buffersize := uint(384000) // 1s @ 3Mbits/

	sck := srtgo.NewSrtSocket(address, port, options)
	err := sck.Listen(1)
	defer sck.Close()
	if err != nil {
		log.Fatalln("listen failed", err)
	}
	log.Printf("Listening on %s:%d\n", address, port)

	server := server.NewServer(buffersize)
	for {
		sock, err := sck.Accept()
		if err != nil {
			log.Fatalln("accept failed", err)
		}
		go server.Handle(sock)
	}
}
