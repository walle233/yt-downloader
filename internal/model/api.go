package model

type HealthResponse struct {
	Status      string `json:"status"`
	Environment string `json:"environment"`
}

type VideoProbeRequest struct {
	URL string `json:"url"`
}

type VideoProbeResponse struct {
	VideoID        string   `json:"videoId"`
	Title          string   `json:"title"`
	DurationSec    int      `json:"durationSec"`
	ThumbnailURL   string   `json:"thumbnailUrl"`
	AllowedFormats []string `json:"allowedFormats"`
	Message        string   `json:"message,omitempty"`
}

type CreateDownloadRequest struct {
	URL          string `json:"url"`
	OutputFormat string `json:"outputFormat"`
}

type CreateDownloadResponse struct {
	JobID   string `json:"jobId"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type DownloadStatusResponse struct {
	JobID    string `json:"jobId"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Step     string `json:"step"`
	Message  string `json:"message,omitempty"`
}

type DownloadListItem struct {
	JobID        string `json:"jobId"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	OutputFormat string `json:"outputFormat"`
	FileName     string `json:"fileName,omitempty"`
	ThumbnailURL string `json:"thumbnailUrl,omitempty"`
	CreatedAt    string `json:"createdAt"`
}
