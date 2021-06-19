package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/voc/srtrelay/config"
	"github.com/voc/srtrelay/srt"
)

// Server serves HTTP API requests
type Server struct {
	conf      config.APIConfig
	srtServer srt.Server
	done      sync.WaitGroup
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
	mux.HandleFunc("/sockets", s.HandleSockets)
	serv := &http.Server{
		Addr:           s.conf.Address,
		Handler:        mux,
		ReadTimeout:    5 * time.Second,
		WriteTimeout:   5 * time.Second,
		MaxHeaderBytes: 1 << 14,
	}

	s.done.Add(1)
	go func() {
		defer s.done.Done()
		err := serv.ListenAndServe()
		if err != nil {
			log.Println(err)
		}
	}()
	s.done.Add(1)
	go func() {
		defer s.done.Done()
		<-ctx.Done()
		ctx2, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		serv.Shutdown(ctx2)
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
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) HandleSockets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	stats := s.srtServer.GetSocketStatistics()
	json.NewEncoder(w).Encode(stats)
}
