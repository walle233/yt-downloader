package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/walle233/yt-downloader/internal/model"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("new pgx pool: %w", err)
	}

	return &Store{pool: pool}, nil
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) CreateDownload(ctx context.Context, sourceURL, outputFormat string, probe model.ProbeResult) (model.Download, error) {
	row := s.pool.QueryRow(ctx, `
		insert into downloads (
			source_url, source_video_id, source_site, title, status, output_format,
			progress, step, duration_sec, thumbnail_url
		)
		values ($1, $2, 'youtube', $3, 'queued', $4, 0, 'queued', $5, $6)
		returning id, source_url, coalesce(source_video_id, ''), source_site, coalesce(title, ''),
			status, output_format, progress, step, coalesce(error_code, ''), coalesce(error_message, ''),
			coalesce(duration_sec, 0), coalesce(file_name, ''), coalesce(file_ext, ''), coalesce(file_size, 0),
			coalesce(r2_object_key, ''), coalesce(thumbnail_url, ''), created_at, updated_at, expires_at
	`, sourceURL, probe.VideoID, probe.Title, outputFormat, probe.DurationSec, probe.ThumbnailURL)

	return scanDownload(row)
}

func (s *Store) GetDownloadByJobID(ctx context.Context, jobID string) (model.Download, error) {
	id, err := parseJobID(jobID)
	if err != nil {
		return model.Download{}, err
	}

	row := s.pool.QueryRow(ctx, `
		select id, source_url, coalesce(source_video_id, ''), source_site, coalesce(title, ''),
			status, output_format, progress, step, coalesce(error_code, ''), coalesce(error_message, ''),
			coalesce(duration_sec, 0), coalesce(file_name, ''), coalesce(file_ext, ''), coalesce(file_size, 0),
			coalesce(r2_object_key, ''), coalesce(thumbnail_url, ''), created_at, updated_at, expires_at
		from downloads
		where id = $1
	`, id)

	return scanDownload(row)
}

func (s *Store) GetDownloadByID(ctx context.Context, id int64) (model.Download, error) {
	row := s.pool.QueryRow(ctx, `
		select id, source_url, coalesce(source_video_id, ''), source_site, coalesce(title, ''),
			status, output_format, progress, step, coalesce(error_code, ''), coalesce(error_message, ''),
			coalesce(duration_sec, 0), coalesce(file_name, ''), coalesce(file_ext, ''), coalesce(file_size, 0),
			coalesce(r2_object_key, ''), coalesce(thumbnail_url, ''), created_at, updated_at, expires_at
		from downloads
		where id = $1
	`, id)

	return scanDownload(row)
}

func (s *Store) UpdateProgress(ctx context.Context, id int64, status, step string, progress int) error {
	_, err := s.pool.Exec(ctx, `
		update downloads
		set status = $2, step = $3, progress = $4, updated_at = now()
		where id = $1
	`, id, status, step, progress)
	return err
}

func (s *Store) MarkFailed(ctx context.Context, id int64, code, message string) error {
	_, err := s.pool.Exec(ctx, `
		update downloads
		set status = 'failed', step = 'failed', error_code = $2, error_message = $3, updated_at = now()
		where id = $1
	`, id, code, message)
	return err
}

func (s *Store) MarkCompleted(ctx context.Context, id int64, artifact model.DownloadArtifact, objectKey string, expiresAt time.Time) error {
	_, err := s.pool.Exec(ctx, `
		update downloads
		set status = 'completed',
			step = 'done',
			progress = 100,
			file_name = $2,
			file_ext = $3,
			file_size = $4,
			r2_object_key = $5,
			expires_at = $6,
			updated_at = now()
		where id = $1
	`, id, artifact.FileName, artifact.FileExt, artifact.FileSize, objectKey, expiresAt)
	return err
}

func (s *Store) ListRecentDownloads(ctx context.Context, limit int) ([]model.Download, error) {
	rows, err := s.pool.Query(ctx, `
		select id, source_url, coalesce(source_video_id, ''), source_site, coalesce(title, ''),
			status, output_format, progress, step, coalesce(error_code, ''), coalesce(error_message, ''),
			coalesce(duration_sec, 0), coalesce(file_name, ''), coalesce(file_ext, ''), coalesce(file_size, 0),
			coalesce(r2_object_key, ''), coalesce(thumbnail_url, ''), created_at, updated_at, expires_at
		from downloads
		order by created_at desc
		limit $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Download, 0, limit)
	for rows.Next() {
		download, err := scanDownload(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, download)
	}

	return items, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDownload(row rowScanner) (model.Download, error) {
	var download model.Download
	err := row.Scan(
		&download.ID,
		&download.SourceURL,
		&download.SourceVideoID,
		&download.SourceSite,
		&download.Title,
		&download.Status,
		&download.OutputFormat,
		&download.Progress,
		&download.Step,
		&download.ErrorCode,
		&download.ErrorMessage,
		&download.DurationSec,
		&download.FileName,
		&download.FileExt,
		&download.FileSize,
		&download.R2ObjectKey,
		&download.ThumbnailURL,
		&download.CreatedAt,
		&download.UpdatedAt,
		&download.ExpiresAt,
	)
	if err != nil {
		return model.Download{}, err
	}

	download.JobID = formatJobID(download.ID)
	return download, nil
}

func formatJobID(id int64) string {
	return fmt.Sprintf("job_%d", id)
}

func parseJobID(jobID string) (int64, error) {
	if !strings.HasPrefix(jobID, "job_") {
		return 0, fmt.Errorf("invalid job id: %s", jobID)
	}

	var id int64
	_, err := fmt.Sscanf(jobID, "job_%d", &id)
	if err != nil {
		return 0, fmt.Errorf("parse job id: %w", err)
	}
	return id, nil
}
