package http

import (
	"context"
	"log"
	"net/http"
	"os"
	"testberry/internal/ports"
)

type Server struct {
	handler *Handler
	addr    string
	logger  ports.Logger
}

func NewServer(service ports.OrderService, addr string, logger ports.Logger) *Server {
	return &Server{
		handler: NewHandler(service, logger),
		addr:    addr,
		logger:  logger,
	}
}

func (s *Server) RunServer(ctx context.Context) error {
	if s.addr == "" {
		s.logger.Error("Addr is not set", "addr", s.addr)
		os.Exit(1)
	}
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("front"))))
	mux.HandleFunc("/order/", s.handler.GetOrder)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "front/index.html")
	})
	server := &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()
	<-ctx.Done()
	return server.Shutdown(ctx)
}
