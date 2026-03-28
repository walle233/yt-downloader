package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/walle233/yt-downloader/internal/model"
)

type YTDLP struct {
	downloadRoot string
}

func New(downloadRoot string) *YTDLP {
	return &YTDLP{downloadRoot: downloadRoot}
}

type probeOutput struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Duration  float64 `json:"duration"`
	Thumbnail string  `json:"thumbnail"`
}

func (y *YTDLP) Probe(ctx context.Context, url string) (model.ProbeResult, error) {
	cmd := exec.CommandContext(ctx, "yt-dlp", "--dump-single-json", "--no-playlist", url)
	output, err := cmd.Output()
	if err != nil {
		return model.ProbeResult{}, fmt.Errorf("yt-dlp probe: %w", err)
	}

	var data probeOutput
	if err := json.Unmarshal(output, &data); err != nil {
		return model.ProbeResult{}, fmt.Errorf("decode probe output: %w", err)
	}

	return model.ProbeResult{
		VideoID:      data.ID,
		Title:        data.Title,
		DurationSec:  int(data.Duration),
		ThumbnailURL: data.Thumbnail,
	}, nil
}

func (y *YTDLP) Download(ctx context.Context, job model.Download) (model.DownloadArtifact, error) {
	workDir := filepath.Join(y.downloadRoot, job.JobID)
	outputDir := filepath.Join(workDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return model.DownloadArtifact{}, fmt.Errorf("mkdir output dir: %w", err)
	}

	outputTemplate := filepath.Join(outputDir, "%(title)s [%(id)s].%(ext)s")
	cmd := exec.CommandContext(ctx, "yt-dlp", "--no-playlist", "-o", outputTemplate, job.SourceURL)
	if output, err := cmd.CombinedOutput(); err != nil {
		return model.DownloadArtifact{}, fmt.Errorf("yt-dlp download: %w; output=%s", err, strings.TrimSpace(string(output)))
	}

	filePath, err := findDownloadedFile(outputDir)
	if err != nil {
		return model.DownloadArtifact{}, err
	}

	if job.OutputFormat == "mp4" && filepath.Ext(filePath) != ".mp4" {
		mp4Path := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".mp4"
		ffmpeg := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", filePath, "-c:v", "libx264", "-c:a", "aac", "-movflags", "+faststart", mp4Path)
		if output, err := ffmpeg.CombinedOutput(); err != nil {
			return model.DownloadArtifact{}, fmt.Errorf("ffmpeg transcode: %w; output=%s", err, strings.TrimSpace(string(output)))
		}

		if err := os.Remove(filePath); err == nil {
			filePath = mp4Path
		} else {
			filePath = mp4Path
		}
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return model.DownloadArtifact{}, fmt.Errorf("stat artifact: %w", err)
	}

	return model.DownloadArtifact{
		FilePath:   filePath,
		FileName:   filepath.Base(filePath),
		FileExt:    strings.TrimPrefix(filepath.Ext(filePath), "."),
		FileSize:   info.Size(),
		WorkDir:    workDir,
		Title:      job.Title,
		VideoID:    job.SourceVideoID,
		Thumbnail:  job.ThumbnailURL,
		SourceURL:  job.SourceURL,
		OutputType: job.OutputFormat,
	}, nil
}

func findDownloadedFile(outputDir string) (string, error) {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return "", fmt.Errorf("read output dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasSuffix(name, ".part") || strings.HasSuffix(name, ".ytdl") {
			continue
		}

		return filepath.Join(outputDir, name), nil
	}

	return "", fmt.Errorf("no downloaded file found")
}
