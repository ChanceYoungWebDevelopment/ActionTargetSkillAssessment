package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ChanceYoungWebDevelopment/ActionTargetSkillAssessment/internal/monitor"
)

type Options struct {
	Addr         string
	PushInterval time.Duration
	StaticDir    string
}

type Server struct {
	opts Options
	mgr  *monitor.Manager

	mu   sync.Mutex
	subs map[chan []byte]struct{}

	httpSrv *http.Server
}

func NewServer(opts Options, mgr *monitor.Manager) *Server {
	return &Server{
		opts: opts,
		mgr:  mgr,
		subs: make(map[chan []byte]struct{}),
	}
}

func (s *Server) Serve() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/api/snapshot", s.handleSnapshot)
	mux.HandleFunc("/api/stream", s.handleStream)
	mux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir(s.opts.StaticDir))))

	s.httpSrv = &http.Server{Addr: s.opts.Addr, Handler: mux}

	go s.broadcastLoop()
	return s.httpSrv.ListenAndServe()
}

func (s *Server) Shutdown(shutdownCtx interface{ Done() <-chan struct{} }) error {
	if s.httpSrv == nil { return nil }
	return s.httpSrv.Close()
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r); return
	}
	path := filepath.Join(s.opts.StaticDir, "index.html")
	http.ServeFile(w, r, path)
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.mgr.SnapshotAll())
}

func (s *Server) handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "stream unsupported", http.StatusInternalServerError)
		return
	}

	ch := make(chan []byte, 8)
	s.mu.Lock(); s.subs[ch] = struct{}{}; s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.subs, ch); s.mu.Unlock() }()

	// initial hello
	fmt.Fprintf(w, ": hi\n\n")
	flusher.Flush()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg := <-ch:
			w.Write([]byte("event: update\n"))
			w.Write([]byte("data: "))
			w.Write(msg)
			w.Write([]byte("\n\n"))
			flusher.Flush()
		}
	}
}

func (s *Server) broadcastLoop() {
	t := s.opts.PushInterval
	if t <= 0 { t = time.Second }
	ticker := time.NewTicker(t)
	defer ticker.Stop()
	for range ticker.C {
		payload, _ := json.Marshal(s.mgr.SnapshotAll())
		s.mu.Lock()
		for ch := range s.subs {
			select { case ch <- payload: default: /* drop slow */ }
		}
		s.mu.Unlock()
	}
	_ = os.Stderr // keep import sanity if unused
}
