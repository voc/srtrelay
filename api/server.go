package api

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/voc/srtrelay/config"
	"github.com/voc/srtrelay/srt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server serves HTTP API requests
type Server struct {
	conf      config.APIConfig
	srtServer srt.Server
	done      sync.WaitGroup
}

func NewServer(conf config.APIConfig, srtServer srt.Server) *Server {
	prometheus.MustRegister(NewExporter(srtServer))
	log.Println("Registered server metrics")
	return &Server{
		conf:      conf,
		srtServer: srtServer,
	}
}

func (s *Server) Listen(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/streams", s.HandleStreams)
	mux.HandleFunc("/sockets", s.HandleSockets)
	mux.Handle("/metrics", promhttp.Handler())
	serv := &http.Server{
		Addr:           s.conf.Address,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 14,
	}

	s.done.Add(2)
	// http listener
	go func() {
		defer s.done.Done()
		err := serv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
		}
	}()

	// shutdown goroutine
	go func() {
		defer s.done.Done()
		<-ctx.Done()
		ctx2, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		if err := serv.Shutdown(ctx2); err != nil {
			log.Println(err)
		}
	}()

	return nil
}

// Wait blocks until listening sockets have been closed
func (s *Server) Wait() {
	s.done.Wait()
}

func (s *Server) HandleStreams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := s.srtServer.GetStatistics()
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Println(err)
	}
}

func (s *Server) HandleSockets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := s.srtServer.GetSocketStatistics()
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Println(err)
	}
}
