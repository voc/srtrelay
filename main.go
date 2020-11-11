package main

import (
	"flag"
	"log"
	"strconv"

	"github.com/haivision/srtgo"
	"github.com/voc/srtrelay/server"
)

func main() {
	var port = flag.Int("port", 1337, "relay port (default 1337)")
	var latency = flag.Int("latency", 300, "srt latency in ms (default 300)")
	flag.Parse()

	options := make(map[string]string)
	options["blocking"] = "0"
	options["transtype"] = "live"
	options["latency"] = strconv.Itoa(*latency)

	address := "0.0.0.0"
	buffersize := uint(384000) // 1s @ 3Mbits/

	sck := srtgo.NewSrtSocket(address, uint16(*port), options)
	err := sck.Listen(1)
	defer sck.Close()
	if err != nil {
		log.Fatalln("listen failed", err)
	}
	log.Printf("Listening on %s:%d\n", address, *port)

	server := server.NewServer(buffersize)
	for {
		sock, err := sck.Accept()
		if err != nil {
			log.Fatalln("accept failed", err)
		}
		go server.Handle(sock)
	}
}
