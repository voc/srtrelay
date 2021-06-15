package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/haivision/srtgo"
	"github.com/voc/srtrelay/api"
	"github.com/voc/srtrelay/config"
	"github.com/voc/srtrelay/relay"
	"github.com/voc/srtrelay/srt"
)

func handleSignal(ctx context.Context, cancel context.CancelFunc) {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case s := <-c:
				log.Println("caught signal", s)
				if s == syscall.SIGHUP {
					continue
				}
				cancel()
			}
		}
	}()
}

func main() {
	// allow specifying config path
	configFlags := flag.NewFlagSet("config", flag.ContinueOnError)
	configFlags.SetOutput(ioutil.Discard)
	configPath := configFlags.String("config", "config.toml", "")
	configFlags.Parse(os.Args[1:])

	// parse config
	conf, err := config.Parse([]string{*configPath, "/etc/srtrelay/config.toml"})
	if err != nil {
		log.Fatal(err)
	}

	// flag just for usage
	flag.String("config", "config.toml", "path to config file")

	// actual flags, use config as default and storage
	var addresses string
	flag.StringVar(&addresses, "addresses", strings.Join(conf.App.Addresses, ","), "relay bind addresses, seperated by commata")
	flag.UintVar(&conf.App.Latency, "latency", conf.App.Latency, "srt protocol latency in ms")
	flag.UintVar(&conf.App.Buffersize, "buffersize", conf.App.Buffersize,
		`relay buffer size in bytes, determines maximum delay of a client`)
	flag.Parse()

	conf.App.Addresses = strings.Split(addresses, ",")

	auth, err := config.GetAuthenticator(conf.Auth)
	if err != nil {
		log.Println(err)
	}

	serverConfig := srt.Config{
		Server: srt.ServerConfig{
			Addresses:   conf.App.Addresses,
			Latency:     conf.App.Latency,
			LossMaxTTL:  conf.App.LossMaxTTL,
			SyncClients: conf.App.SyncClients,
			Auth:        auth,
		},
		Relay: relay.RelayConfig{
			Buffersize: conf.App.Buffersize, // 1s @ 3Mbits/
		},
	}

	// setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	handleSignal(ctx, cancel)

	// create server
	srtgo.InitSRT()
	srtServer := srt.NewServer(&serverConfig)
	err = srtServer.Listen(ctx)
	if err != nil {
		log.Fatal(err)
	}

	var apiServer *api.Server
	if conf.API.Enabled {
		apiServer = api.NewServer(conf.API, srtServer)
		err := apiServer.Listen(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("API listening on %s\n", conf.API.Address)
	}

	// Wait for graceful shutdown
	srtServer.Wait()
	if apiServer != nil {
		apiServer.Wait()
	}
	srtgo.CleanupSRT()
}
