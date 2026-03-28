package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/walle233/yt-downloader/internal/config"
	"github.com/walle233/yt-downloader/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := os.MkdirAll(filepath.Clean(cfg.DownloadRoot), 0o755); err != nil {
		log.Fatalf("create download root: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runtime, err := service.NewRuntime(ctx, cfg)
	if err != nil {
		log.Fatalf("create runtime: %v", err)
	}
	defer runtime.Close()

	log.Printf("worker started")
	log.Printf("redis addr: %s", cfg.RedisAddr)
	log.Printf("download root: %s", cfg.DownloadRoot)
	log.Printf("worker concurrency: %d", cfg.WorkerConcurrency)
	log.Printf("poll interval: %s", cfg.WorkerPollInterval)

	if err := runtime.RunWorker(ctx); err != nil && err != context.Canceled {
		log.Fatalf("run worker: %v", err)
	}
}
