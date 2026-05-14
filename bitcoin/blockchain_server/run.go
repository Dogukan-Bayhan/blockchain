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

// Run starts the P2P service and the HTTP API.
func (bcs *BlockchainServer) Run() {
	bcs.StartNetwork()

	mux := http.NewServeMux()
	mux.HandleFunc("/", bcs.GetChain)
	mux.HandleFunc("/transactions", bcs.Transactions)
	mux.HandleFunc("/mine", bcs.Mine)
	mux.HandleFunc("/mine/start", bcs.StartMine)
	mux.HandleFunc("/amount", bcs.Amount)

	server := &http.Server{
		Addr:    "0.0.0.0:" + strconv.Itoa(int(bcs.Port())),
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

	bcs.StopNetwork()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("http: shutdown error: %v", err)
	}
}
