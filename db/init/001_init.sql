create table if not exists users (
  id bigserial primary key,
  email text not null unique,
  password_hash text not null,
  role text not null default 'user',
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create table if not exists downloads (
  id bigserial primary key,
  user_id bigint references users(id) on delete set null,
  clerk_user_id text,
  source_url text not null,
  source_video_id text,
  source_site text not null default 'youtube',
  title text,
  status text not null,
  output_format text not null,
  profile_id text,
  media_kind text,
  target_height integer,
  progress integer not null default 0,
  step text not null default 'queued',
  error_code text,
  error_message text,
  duration_sec integer,
  file_name text,
  file_ext text,
  file_size bigint,
  r2_object_key text,
  thumbnail_url text,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  expires_at timestamptz
);

create index if not exists idx_downloads_status on downloads(status);
create index if not exists idx_downloads_user_id on downloads(user_id);
create index if not exists idx_downloads_clerk_user_id on downloads(clerk_user_id);
create index if not exists idx_downloads_clerk_user_id_created_at on downloads(clerk_user_id, created_at desc);
create index if not exists idx_downloads_source_video_id on downloads(source_video_id);

create table if not exists video_meta_cache (
  source_video_id text primary key,
  title text,
  duration_sec integer,
  thumbnail_url text,
  raw_meta_json jsonb not null default '{}'::jsonb,
  fetched_at timestamptz not null default now()
);

create table if not exists audit_logs (
  id bigserial primary key,
  user_id bigint references users(id) on delete set null,
  action text not null,
  target_id text,
  ip inet,
  user_agent text,
  created_at timestamptz not null default now()
);
