package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/walle233/yt-downloader/internal/config"
	"github.com/walle233/yt-downloader/internal/downloader"
	"github.com/walle233/yt-downloader/internal/model"
	"github.com/walle233/yt-downloader/internal/queue"
	"github.com/walle233/yt-downloader/internal/storage"
	"github.com/walle233/yt-downloader/internal/store"
)

type Runtime struct {
	Config     config.Config
	Store      *store.Store
	Queue      *queue.RedisQueue
	Downloader *downloader.YTDLP
	Storage    *storage.R2Storage
}

var (
	ErrProfileUnavailable = errors.New("requested download profile is not available")
	ErrFreeLimitReached   = store.ErrFreeLimitReached
)

func NewRuntime(ctx context.Context, cfg config.Config) (*Runtime, error) {
	db, err := store.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	redisQueue := queue.NewRedisQueue(cfg.RedisAddr)

	r2Storage, err := storage.NewR2Storage(ctx, cfg.R2)
	if err != nil {
		db.Close()
		_ = redisQueue.Close()
		return nil, err
	}

	return &Runtime{
		Config:     cfg,
		Store:      db,
		Queue:      redisQueue,
		Downloader: downloader.New(cfg.DownloadRoot, cfg.YTDLPCookiesFile, cfg.YTDLPJSRuntimes, cfg.YTDLPRemoteComponents),
		Storage:    r2Storage,
	}, nil
}

func (r *Runtime) Close() {
	if r.Store != nil {
		r.Store.Close()
	}
	if r.Queue != nil {
		_ = r.Queue.Close()
	}
}

func (r *Runtime) HealthCheck(ctx context.Context) error {
	if err := r.Store.Ping(ctx); err != nil {
		return fmt.Errorf("postgres ping: %w", err)
	}
	if err := r.Queue.Ping(ctx); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}

func (r *Runtime) CreateDownload(ctx context.Context, clerkUserID, url string, profileID model.DownloadProfileID) (model.Download, error) {
	probe, err := r.Downloader.Probe(ctx, url)
	if err != nil {
		return model.Download{}, err
	}

	profile, ok := model.FindAvailableProfile(probe.Profiles, profileID)
	if !ok {
		return model.Download{}, ErrProfileUnavailable
	}

	spec, ok := model.FindDownloadProfileSpec(profile.ID)
	if !ok {
		return model.Download{}, ErrProfileUnavailable
	}

	download, consumedFree, err := r.Store.CreateDownload(ctx, clerkUserID, spec, url, probe)
	if err != nil {
		return model.Download{}, err
	}

	if err := r.Queue.EnqueueDownload(ctx, download.ID); err != nil {
		rollbackErr := r.Store.RollbackCreatedDownload(ctx, clerkUserID, download.ID, consumedFree)
		if rollbackErr != nil {
			fmt.Printf("rollback created download failed: job=%s err=%v\n", download.JobID, rollbackErr)
		}
		return model.Download{}, err
	}

	return download, nil
}

func (r *Runtime) GetStatus(ctx context.Context, clerkUserID, jobID string) (model.Download, error) {
	return r.Store.GetDownloadByJobIDForUser(ctx, clerkUserID, jobID)
}

func (r *Runtime) GetResultURL(ctx context.Context, clerkUserID, jobID string) (model.Download, string, error) {
	download, err := r.Store.GetDownloadByJobIDForUser(ctx, clerkUserID, jobID)
	if err != nil {
		return model.Download{}, "", err
	}
	if download.Status != "completed" || download.R2ObjectKey == "" {
		return model.Download{}, "", fmt.Errorf("download is not ready")
	}

	url, err := r.Storage.PresignGetObject(ctx, download.R2ObjectKey, 15*time.Minute)
	if err != nil {
		return model.Download{}, "", err
	}

	return download, url, nil
}

func (r *Runtime) ListRecentDownloads(ctx context.Context, clerkUserID string, limit int) ([]model.Download, error) {
	return r.Store.ListRecentDownloadsByUser(ctx, clerkUserID, limit)
}

func (r *Runtime) GetBillingSummary(ctx context.Context, clerkUserID string) (model.BillingSummary, error) {
	account, err := r.Store.GetBillingAccount(ctx, clerkUserID)
	if err != nil {
		return model.BillingSummary{}, err
	}

	return account.ToSummary(time.Now()), nil
}

func (r *Runtime) RunWorker(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		id, err := r.Queue.BlockingPopDownload(ctx, r.Config.WorkerPollInterval)
		if err != nil {
			if errors.Is(err, redis.Nil) || errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			return fmt.Errorf("blocking pop download: %w", err)
		}

		if err := r.processJob(ctx, id); err != nil {
			_ = r.Store.MarkFailed(ctx, id, "worker_error", err.Error())
		}
	}
}

func (r *Runtime) processJob(ctx context.Context, id int64) error {
	download, err := r.Store.GetDownloadByID(ctx, id)
	if err != nil {
		return err
	}

	if err := r.Store.UpdateProgress(ctx, id, "downloading", "downloading", 10); err != nil {
		return err
	}

	artifact, err := r.Downloader.Download(ctx, download)
	if err != nil {
		return err
	}
	defer func() {
		if artifact.WorkDir != "" {
			_ = removeWorkDir(artifact.WorkDir)
		}
	}()

	if err := r.Store.UpdateProgress(ctx, id, "uploading", "uploading", 80); err != nil {
		return err
	}

	objectKey := buildObjectKey(download.JobID, artifact.FileName)
	if err := r.Storage.UploadFile(ctx, objectKey, artifact.FilePath); err != nil {
		return err
	}

	expiresAt := time.Now().Add(time.Duration(r.Config.DownloadTTLHours) * time.Hour)
	if err := r.Store.MarkCompleted(ctx, id, artifact, objectKey, expiresAt); err != nil {
		return err
	}

	return nil
}

func buildObjectKey(jobID, fileName string) string {
	now := time.Now()
	return fmt.Sprintf("downloads/%04d/%02d/%s/%s", now.Year(), int(now.Month()), jobID, filepath.Base(fileName))
}

func removeWorkDir(path string) error {
	return osRemoveAll(path)
}

var osRemoveAll = func(path string) error {
	return removeAll(path)
}
