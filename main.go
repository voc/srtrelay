package main

import (
	"context"
	"flag"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"net/netip"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/voc/srtrelay/api"
	"github.com/voc/srtrelay/config"
	"github.com/voc/srtrelay/relay"
	"github.com/voc/srtrelay/srt"
)

func parseFlags(conf *config.Config) {
	flags := flag.NewFlagSet("srtrelay", flag.ExitOnError)

	// flag just for usage
	flags.String("config", "config.toml", "path to config file")

	// actual flag parsing, use config as default values
	addresses := flags.String("addresses", strings.Join(conf.App.Addresses, ","), "relay bind addresses, separated by commas")
	flags.UintVar(&conf.App.LatencyMs, "latency", conf.App.LatencyMs, "srt protocol latency in ms")
	flags.UintVar(&conf.App.Buffersize, "buffersize", conf.App.Buffersize,
		`relay buffer size in bytes, determines maximum delay of a client`)
	flags.StringVar(&conf.Profile.Address, "pprof", "", "enable profiling server on given address")
	_ = flags.Parse(os.Args[1:])
	conf.App.Addresses = strings.Split(*addresses, ",")
}

func main() {
	// first flags parse to get config path
	configFlags := flag.NewFlagSet("srtrelay", flag.ContinueOnError)
	configFlags.SetOutput(io.Discard)
	configPath := configFlags.String("config", "config.toml", "")
	err := configFlags.Parse(os.Args[1:])
	if err == flag.ErrHelp {
		// will print usage and exit
		parseFlags(&config.Config{})
	}

	// parse config
	conf, err := config.Parse([]string{*configPath, "/etc/srtrelay/config.toml"})
	if err != nil {
		log.Fatal(err)
	}
	parseFlags(conf)

	if conf.Profile.Address != "" {
		log.Println("Enabling profiling on", conf.Profile.Address)
		if err := enablePprof(conf.Profile.Address); err != nil {
			log.Println("failed to enable profiling:", err)
		}
	}

	var addresses []netip.AddrPort
	for _, addr := range conf.App.Addresses {
		addrs, err := config.ParseAddress(strings.TrimSpace(addr))
		if err != nil {
			log.Fatalf("invalid address %s: %v", addr, err)
		}
		addresses = append(addresses, addrs...)
	}

	auth, err := config.GetAuthenticator(conf.Auth)
	if err != nil {
		log.Fatal(err)
	}

	serverConfig := srt.Config{
		Server: srt.ServerConfig{
			Addresses:     addresses,
			PublicAddress: conf.App.PublicAddress,
			LatencyMs:     conf.App.LatencyMs,
			LossMaxTTL:    conf.App.LossMaxTTL,
			SyncClients:   conf.App.SyncClients,
			PacketSize:    conf.App.PacketSize,
			Auth:          auth,
		},
		Relay: relay.RelayConfig{
			BufferSize: conf.App.Buffersize,
			PacketSize: uint(conf.App.PacketSize),
		},
	}

	// setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	handleSignal(ctx, cancel)

	// create server
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

	<-ctx.Done()

	// Wait for graceful shutdown
	shutdownDone := make(chan struct{})
	go func() {
		srtServer.Wait()
		if apiServer != nil {
			apiServer.Wait()
		}
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
	case <-time.After(time.Second * 2):
		slog.Warn("Graceful shutdown timed out, forcing exit")
	}
}

func enablePprof(addr string) error {
	conn, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	srv := http.Server{
		Handler: mux,
	}
	go func() {
		err := srv.Serve(conn)
		if err != nil && err != http.ErrServerClosed {
			log.Println(err)
		}
	}()
	return nil
}

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
