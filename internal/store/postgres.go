package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/walle233/yt-downloader/internal/model"
)

type Store struct {
	pool *pgxpool.Pool
}

var ErrFreeLimitReached = errors.New("free download limit reached")

func New(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("new pgx pool: %w", err)
	}

	store := &Store{pool: pool}
	if err := store.ensureSchema(ctx); err != nil {
		store.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) CreateDownload(ctx context.Context, clerkUserID string, spec model.DownloadProfileSpec, sourceURL string, probe model.ProbeResult) (model.Download, bool, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return model.Download{}, false, err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	account, err := ensureBillingAccountForUpdate(ctx, tx, clerkUserID)
	if err != nil {
		return model.Download{}, false, err
	}

	now := time.Now()
	consumeFree := !account.IsProActive(now)
	if consumeFree && account.FreeDownloadsUsed >= account.FreeDownloadsLimit {
		return model.Download{}, false, ErrFreeLimitReached
	}

	row := tx.QueryRow(ctx, `
		insert into downloads (
			clerk_user_id, source_url, source_video_id, source_site, title, status, output_format,
			profile_id, media_kind, target_height, progress, step, duration_sec, thumbnail_url
		)
		values ($1, $2, $3, 'youtube', $4, 'queued', $5, $6, $7, $8, 0, 'queued', $9, $10)
		returning id, coalesce(clerk_user_id, ''), source_url, coalesce(source_video_id, ''), source_site, coalesce(title, ''),
			status, output_format, coalesce(profile_id, ''), coalesce(media_kind, ''), coalesce(target_height, 0),
			progress, step, coalesce(error_code, ''), coalesce(error_message, ''),
			coalesce(duration_sec, 0), coalesce(file_name, ''), coalesce(file_ext, ''), coalesce(file_size, 0),
			coalesce(r2_object_key, ''), coalesce(thumbnail_url, ''), created_at, updated_at, expires_at
	`,
		clerkUserID,
		sourceURL,
		probe.VideoID,
		probe.Title,
		spec.Container,
		spec.ID,
		spec.Kind,
		nullableHeight(spec.TargetHeight),
		probe.DurationSec,
		probe.ThumbnailURL,
	)

	download, err := scanDownload(row)
	if err != nil {
		return model.Download{}, false, err
	}

	if consumeFree {
		if _, err := tx.Exec(ctx, `
			update billing_accounts
			set free_downloads_used = free_downloads_used + 1,
				updated_at = now()
			where clerk_user_id = $1
		`, clerkUserID); err != nil {
			return model.Download{}, false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return model.Download{}, false, err
	}

	return download, consumeFree, nil
}

func (s *Store) RollbackCreatedDownload(ctx context.Context, clerkUserID string, downloadID int64, restoreFreeUsage bool) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	result, err := tx.Exec(ctx, `
		delete from downloads
		where id = $1 and clerk_user_id = $2 and status = 'queued'
	`, downloadID, clerkUserID)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("rollback created download: queued download not found")
	}

	if restoreFreeUsage {
		if _, err := tx.Exec(ctx, `
			update billing_accounts
			set free_downloads_used = greatest(free_downloads_used - 1, 0),
				updated_at = now()
			where clerk_user_id = $1
		`, clerkUserID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *Store) GetBillingAccount(ctx context.Context, clerkUserID string) (model.BillingAccount, error) {
	if err := s.ensureBillingAccount(ctx, clerkUserID); err != nil {
		return model.BillingAccount{}, err
	}

	row := s.pool.QueryRow(ctx, `
		select clerk_user_id,
			plan_code,
			subscription_status,
			coalesce(billing_interval, ''),
			coalesce(stripe_customer_id, ''),
			coalesce(stripe_subscription_id, ''),
			current_period_end,
			cancel_at_period_end,
			free_downloads_limit,
			free_downloads_used,
			created_at,
			updated_at
		from billing_accounts
		where clerk_user_id = $1
	`, clerkUserID)

	return scanBillingAccount(row)
}

func (s *Store) GetDownloadByJobID(ctx context.Context, jobID string) (model.Download, error) {
	id, err := parseJobID(jobID)
	if err != nil {
		return model.Download{}, err
	}

	row := s.pool.QueryRow(ctx, `
		select id, coalesce(clerk_user_id, ''), source_url, coalesce(source_video_id, ''), source_site, coalesce(title, ''),
			status, output_format, coalesce(profile_id, ''), coalesce(media_kind, ''), coalesce(target_height, 0),
			progress, step, coalesce(error_code, ''), coalesce(error_message, ''),
			coalesce(duration_sec, 0), coalesce(file_name, ''), coalesce(file_ext, ''), coalesce(file_size, 0),
			coalesce(r2_object_key, ''), coalesce(thumbnail_url, ''), created_at, updated_at, expires_at
		from downloads
		where id = $1
	`, id)

	return scanDownload(row)
}

func (s *Store) GetDownloadByID(ctx context.Context, id int64) (model.Download, error) {
	row := s.pool.QueryRow(ctx, `
		select id, coalesce(clerk_user_id, ''), source_url, coalesce(source_video_id, ''), source_site, coalesce(title, ''),
			status, output_format, coalesce(profile_id, ''), coalesce(media_kind, ''), coalesce(target_height, 0),
			progress, step, coalesce(error_code, ''), coalesce(error_message, ''),
			coalesce(duration_sec, 0), coalesce(file_name, ''), coalesce(file_ext, ''), coalesce(file_size, 0),
			coalesce(r2_object_key, ''), coalesce(thumbnail_url, ''), created_at, updated_at, expires_at
		from downloads
		where id = $1
	`, id)

	return scanDownload(row)
}

func (s *Store) GetDownloadByJobIDForUser(ctx context.Context, clerkUserID, jobID string) (model.Download, error) {
	id, err := parseJobID(jobID)
	if err != nil {
		return model.Download{}, err
	}

	row := s.pool.QueryRow(ctx, `
		select id, coalesce(clerk_user_id, ''), source_url, coalesce(source_video_id, ''), source_site, coalesce(title, ''),
			status, output_format, coalesce(profile_id, ''), coalesce(media_kind, ''), coalesce(target_height, 0),
			progress, step, coalesce(error_code, ''), coalesce(error_message, ''),
			coalesce(duration_sec, 0), coalesce(file_name, ''), coalesce(file_ext, ''), coalesce(file_size, 0),
			coalesce(r2_object_key, ''), coalesce(thumbnail_url, ''), created_at, updated_at, expires_at
		from downloads
		where id = $1 and clerk_user_id = $2
	`, id, clerkUserID)

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

func (s *Store) ListRecentDownloadsByUser(ctx context.Context, clerkUserID string, limit int) ([]model.Download, error) {
	rows, err := s.pool.Query(ctx, `
		select id, coalesce(clerk_user_id, ''), source_url, coalesce(source_video_id, ''), source_site, coalesce(title, ''),
			status, output_format, coalesce(profile_id, ''), coalesce(media_kind, ''), coalesce(target_height, 0),
			progress, step, coalesce(error_code, ''), coalesce(error_message, ''),
			coalesce(duration_sec, 0), coalesce(file_name, ''), coalesce(file_ext, ''), coalesce(file_size, 0),
			coalesce(r2_object_key, ''), coalesce(thumbnail_url, ''), created_at, updated_at, expires_at
		from downloads
		where clerk_user_id = $1
		order by created_at desc
		limit $2
	`, clerkUserID, limit)
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

func (s *Store) ensureSchema(ctx context.Context) error {
	statements := []string{
		`create table if not exists billing_accounts (
			clerk_user_id text primary key,
			plan_code text not null default 'free',
			subscription_status text not null default 'inactive',
			billing_interval text,
			stripe_customer_id text,
			stripe_subscription_id text,
			current_period_end timestamptz,
			cancel_at_period_end boolean not null default false,
			free_downloads_limit integer not null default 3,
			free_downloads_used integer not null default 0,
			created_at timestamptz not null default now(),
			updated_at timestamptz not null default now()
		)`,
		`alter table downloads add column if not exists clerk_user_id text`,
		`alter table downloads add column if not exists profile_id text`,
		`alter table downloads add column if not exists media_kind text`,
		`alter table downloads add column if not exists target_height integer`,
		`alter table billing_accounts add column if not exists plan_code text not null default 'free'`,
		`alter table billing_accounts add column if not exists subscription_status text not null default 'inactive'`,
		`alter table billing_accounts add column if not exists billing_interval text`,
		`alter table billing_accounts add column if not exists stripe_customer_id text`,
		`alter table billing_accounts add column if not exists stripe_subscription_id text`,
		`alter table billing_accounts add column if not exists current_period_end timestamptz`,
		`alter table billing_accounts add column if not exists cancel_at_period_end boolean not null default false`,
		`alter table billing_accounts add column if not exists free_downloads_limit integer not null default 3`,
		`alter table billing_accounts add column if not exists free_downloads_used integer not null default 0`,
		`create index if not exists idx_downloads_clerk_user_id on downloads(clerk_user_id)`,
		`create index if not exists idx_downloads_clerk_user_id_created_at on downloads(clerk_user_id, created_at desc)`,
		`create index if not exists idx_billing_accounts_plan_code on billing_accounts(plan_code)`,
	}

	for _, statement := range statements {
		if _, err := s.pool.Exec(ctx, statement); err != nil {
			return fmt.Errorf("ensure schema: %w", err)
		}
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDownload(row rowScanner) (model.Download, error) {
	var download model.Download
	err := row.Scan(
		&download.ID,
		&download.ClerkUserID,
		&download.SourceURL,
		&download.SourceVideoID,
		&download.SourceSite,
		&download.Title,
		&download.Status,
		&download.OutputFormat,
		&download.ProfileID,
		&download.MediaKind,
		&download.TargetHeight,
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

func nullableHeight(height int) any {
	if height <= 0 {
		return nil
	}
	return height
}

func (s *Store) ensureBillingAccount(ctx context.Context, clerkUserID string) error {
	_, err := s.pool.Exec(ctx, `
		insert into billing_accounts (clerk_user_id)
		values ($1)
		on conflict (clerk_user_id) do nothing
	`, clerkUserID)
	return err
}

func ensureBillingAccountForUpdate(ctx context.Context, tx pgx.Tx, clerkUserID string) (model.BillingAccount, error) {
	if _, err := tx.Exec(ctx, `
		insert into billing_accounts (clerk_user_id)
		values ($1)
		on conflict (clerk_user_id) do nothing
	`, clerkUserID); err != nil {
		return model.BillingAccount{}, err
	}

	row := tx.QueryRow(ctx, `
		select clerk_user_id,
			plan_code,
			subscription_status,
			coalesce(billing_interval, ''),
			coalesce(stripe_customer_id, ''),
			coalesce(stripe_subscription_id, ''),
			current_period_end,
			cancel_at_period_end,
			free_downloads_limit,
			free_downloads_used,
			created_at,
			updated_at
		from billing_accounts
		where clerk_user_id = $1
		for update
	`, clerkUserID)

	return scanBillingAccount(row)
}

func scanBillingAccount(row rowScanner) (model.BillingAccount, error) {
	var account model.BillingAccount
	err := row.Scan(
		&account.ClerkUserID,
		&account.PlanCode,
		&account.SubscriptionStatus,
		&account.BillingInterval,
		&account.StripeCustomerID,
		&account.StripeSubscriptionID,
		&account.CurrentPeriodEnd,
		&account.CancelAtPeriodEnd,
		&account.FreeDownloadsLimit,
		&account.FreeDownloadsUsed,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		return model.BillingAccount{}, err
	}
	return account, nil
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
