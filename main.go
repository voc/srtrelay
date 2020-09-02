package main

import (
	"log"

	"github.com/haivision/srtgo"
	"github.com/voc/srtrelay/relay"
)

// Socket Options
// const (
// // streamIDSocketOption =
// )

func main() {
	options := make(map[string]string)
	options["blocking"] = "0"
	options["transtype"] = "live"
	options["latency"] = "20"

	address := "0.0.0.0"
	port := uint16(8090)

	sck := srtgo.NewSrtSocket(address, port, options)
	err := sck.Listen(1)
	defer sck.Close()
	if err != nil {
		log.Fatalln("listen failed", err)
	}
	log.Printf("Listening on %s:%d\n", address, port)
	// log.Println("packetSize", sck.PacketSize())

	server := relay.NewServer()

	// Handle SIGTERM signal
	// ch := make(chan os.Signal, 1)
	// signal.Notify(ch, syscall.SIGTERM)
	// go func() {
	// 	<-ch
	// 	cancel()
	// }()

	for {
		sock, err := sck.Accept()
		log.Println("packetsize", sock.PacketSize())
		// stats, _ := sock.Stats()
		if err != nil {
			log.Fatalln("accept failed", err)
		}
		go server.Handle(sock)
	}
}
