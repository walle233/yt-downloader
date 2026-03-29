package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/walle233/yt-downloader/internal/config"
	"github.com/walle233/yt-downloader/internal/httpserver"
	"github.com/walle233/yt-downloader/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.ClerkSecretKey != "" {
		clerk.SetKey(cfg.ClerkSecretKey)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runtime, err := service.NewRuntime(ctx, cfg)
	if err != nil {
		log.Fatalf("create runtime: %v", err)
	}
	defer runtime.Close()

	server := httpserver.New(cfg, runtime)
	addr := ":" + cfg.APIPort

	log.Printf("api listening on %s", addr)
	log.Printf("download root: %s", cfg.DownloadRoot)

	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatalf("listen and serve: %v", err)
	}
}
