package model

import "time"

type Download struct {
	ID            int64      `json:"id"`
	JobID         string     `json:"jobId"`
	UserID        *int64     `json:"userId,omitempty"`
	ClerkUserID   string     `json:"clerkUserId,omitempty"`
	SourceURL     string     `json:"sourceUrl"`
	SourceVideoID string     `json:"sourceVideoId"`
	SourceSite    string     `json:"sourceSite"`
	Title         string     `json:"title"`
	Status        string     `json:"status"`
	OutputFormat  string     `json:"outputFormat"`
	ProfileID     string     `json:"profileId"`
	MediaKind     string     `json:"mediaKind"`
	TargetHeight  int        `json:"targetHeight,omitempty"`
	Progress      int        `json:"progress"`
	Step          string     `json:"step"`
	ErrorCode     string     `json:"errorCode,omitempty"`
	ErrorMessage  string     `json:"errorMessage,omitempty"`
	DurationSec   int        `json:"durationSec"`
	FileName      string     `json:"fileName,omitempty"`
	FileExt       string     `json:"fileExt,omitempty"`
	FileSize      int64      `json:"fileSize,omitempty"`
	R2ObjectKey   string     `json:"r2ObjectKey,omitempty"`
	ThumbnailURL  string     `json:"thumbnailUrl,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
}

type ProbeResult struct {
	VideoID      string
	Title        string
	DurationSec  int
	ThumbnailURL string
	Profiles     []DownloadProfile
}

type DownloadArtifact struct {
	FilePath   string
	FileName   string
	FileExt    string
	FileSize   int64
	WorkDir    string
	Title      string
	VideoID    string
	Thumbnail  string
	SourceURL  string
	OutputType string
}
