package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/voc/srtrelay/config"
	"github.com/voc/srtrelay/srt"
)

// Server serves HTTP API requests
type Server struct {
	conf      config.APIConfig
	srtServer srt.Server
}

func NewServer(conf config.APIConfig, srtServer srt.Server) *Server {
	return &Server{
		conf:      conf,
		srtServer: srtServer,
	}
}

func (s *Server) Listen(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/streams", s.HandleStreams)
	serv := &http.Server{
		Addr:           s.conf.Address,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 14,
	}

	go func() {
		log.Fatal(serv.ListenAndServe())
	}()
	go func() {
		<-ctx.Done()
		ctx2, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		serv.Shutdown(ctx2)
	}()

	return nil
}

func (s *Server) HandleStreams(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := s.srtServer.GetStatistics()
	json.NewEncoder(w).Encode(stats)
}
