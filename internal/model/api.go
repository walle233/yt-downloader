package model

type HealthResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
}

type VideoProbeRequest struct {
	URL string `json:"url"`
}

type VideoProbeResponse struct {
	VideoID      string            `json:"videoId"`
	Title        string            `json:"title"`
	DurationSec  int               `json:"durationSec"`
	ThumbnailURL string            `json:"thumbnailUrl"`
	Profiles     []DownloadProfile `json:"profiles"`
	Message      string            `json:"message,omitempty"`
}

type CreateDownloadRequest struct {
	URL       string `json:"url"`
	ProfileID string `json:"profileId"`
}

type CreateDownloadResponse struct {
	JobID     string `json:"jobId"`
	Status    string `json:"status"`
	ProfileID string `json:"profileId"`
	Message   string `json:"message,omitempty"`
}

type DownloadStatusResponse struct {
	JobID        string `json:"jobId"`
	Status       string `json:"status"`
	Progress     int    `json:"progress"`
	Step         string `json:"step"`
	Message      string `json:"message,omitempty"`
	Title        string `json:"title"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	DurationSec  int    `json:"durationSec,omitempty"`
	ProfileID    string `json:"profileId"`
	MediaKind    string `json:"mediaKind"`
	TargetHeight int    `json:"targetHeight,omitempty"`
	FileName     string `json:"fileName,omitempty"`
	FileSize     int64  `json:"fileSize,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

type DownloadListItem struct {
	JobID        string `json:"jobId"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	ProfileID    string `json:"profileId"`
	MediaKind    string `json:"mediaKind"`
	TargetHeight int    `json:"targetHeight,omitempty"`
	FileName     string `json:"fileName,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

type DownloadResultResponse struct {
	JobID       string `json:"jobId"`
	Status      string `json:"status"`
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	DownloadURL string `json:"downloadUrl"`
}

type DownloadListResponse struct {
	Items []DownloadListItem `json:"items"`
}
