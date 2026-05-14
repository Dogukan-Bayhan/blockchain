package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Run registers wallet HTTP handlers and starts serving requests.
func (ws *WalletServer) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", ws.Index)
	mux.HandleFunc("/wallet", ws.Wallet)
	mux.HandleFunc("/transaction", ws.CreateTransaction)
	mux.HandleFunc("/wallet/amount", ws.WalletAmount)

	server := &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(int(ws.Port())),
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("http: listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http: listen error: %v", err)
		}
	}()

	<-ctx.Done()
	stop()
	log.Printf("server: shutdown requested")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("http: shutdown error: %v", err)
	}
}
