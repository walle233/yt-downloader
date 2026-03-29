package downloader

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/walle233/yt-downloader/internal/model"
)

type YTDLP struct {
	downloadRoot string
	cookiesFile  string
	jsRuntimes   string
	remoteComponents string
}

func New(downloadRoot, cookiesFile, jsRuntimes, remoteComponents string) *YTDLP {
	return &YTDLP{
		downloadRoot: downloadRoot,
		cookiesFile:  cookiesFile,
		jsRuntimes:   jsRuntimes,
		remoteComponents: remoteComponents,
	}
}

type probeOutput struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Duration  float64       `json:"duration"`
	Thumbnail string        `json:"thumbnail"`
	Formats   []probeFormat `json:"formats"`
}

type probeFormat struct {
	FormatID       string  `json:"format_id"`
	Ext            string  `json:"ext"`
	AudioCodec     string  `json:"acodec"`
	VideoCodec     string  `json:"vcodec"`
	Height         int     `json:"height"`
	FileSize       int64   `json:"filesize"`
	FileSizeApprox int64   `json:"filesize_approx"`
	Bitrate        float64 `json:"tbr"`
}

func (y *YTDLP) Probe(ctx context.Context, url string) (model.ProbeResult, error) {
	baseArgs, cleanup, err := y.baseArgs()
	if err != nil {
		return model.ProbeResult{}, err
	}
	defer cleanup()

	args := append([]string{"--dump-single-json"}, baseArgs...)
	args = append(args, url)
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return model.ProbeResult{}, fmt.Errorf("yt-dlp probe: %w; output=%s", err, strings.TrimSpace(string(output)))
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
		Profiles:     buildProfiles(data.Formats, int(data.Duration)),
	}, nil
}

func (y *YTDLP) Download(ctx context.Context, job model.Download) (model.DownloadArtifact, error) {
	workDir := filepath.Join(y.downloadRoot, job.JobID)
	outputDir := filepath.Join(workDir, "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return model.DownloadArtifact{}, fmt.Errorf("mkdir output dir: %w", err)
	}

	outputTemplate := filepath.Join(outputDir, "%(title)s [%(id)s].%(ext)s")
	spec, ok := model.FindDownloadProfileSpec(model.DownloadProfileID(job.ProfileID))
	if !ok {
		return model.DownloadArtifact{}, fmt.Errorf("unsupported download profile: %s", job.ProfileID)
	}

	baseArgs, cleanup, err := y.baseArgs()
	if err != nil {
		return model.DownloadArtifact{}, err
	}
	defer cleanup()

	args := append(baseArgs, "-o", outputTemplate)
	switch spec.Kind {
	case "audio":
		args = append(args, "-f", "bestaudio")
	case "video":
		formatSelector := fmt.Sprintf(
			"bestvideo[height=%d]+bestaudio/best[height=%d]/bestvideo[height<=%d]+bestaudio/best[height<=%d]",
			spec.TargetHeight,
			spec.TargetHeight,
			spec.TargetHeight,
			spec.TargetHeight,
		)
		args = append(args, "-f", formatSelector)
	default:
		return model.DownloadArtifact{}, fmt.Errorf("unsupported media kind: %s", spec.Kind)
	}
	args = append(args, job.SourceURL)

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return model.DownloadArtifact{}, fmt.Errorf("yt-dlp download: %w; output=%s", err, strings.TrimSpace(string(output)))
	}

	filePath, err := findDownloadedFile(outputDir)
	if err != nil {
		return model.DownloadArtifact{}, err
	}

	switch spec.Kind {
	case "audio":
		mp3Path := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".mp3"
		ffmpeg := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", filePath, "-vn", "-acodec", "libmp3lame", "-b:a", "192k", mp3Path)
		if output, err := ffmpeg.CombinedOutput(); err != nil {
			return model.DownloadArtifact{}, fmt.Errorf("ffmpeg mp3 transcode: %w; output=%s", err, strings.TrimSpace(string(output)))
		}
		filePath = mp3Path
	case "video":
		mp4Path := strings.TrimSuffix(filePath, filepath.Ext(filePath)) + ".mp4"
		ffmpeg := exec.CommandContext(ctx, "ffmpeg", "-y", "-i", filePath, "-c:v", "libx264", "-c:a", "aac", "-movflags", "+faststart", mp4Path)
		if output, err := ffmpeg.CombinedOutput(); err != nil {
			return model.DownloadArtifact{}, fmt.Errorf("ffmpeg transcode: %w; output=%s", err, strings.TrimSpace(string(output)))
		}
		filePath = mp4Path
	}

	if err := removeInputArtifact(filePath, outputDir); err != nil {
		return model.DownloadArtifact{}, err
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
		OutputType: spec.Container,
	}, nil
}

func (y *YTDLP) baseArgs() ([]string, func(), error) {
	args := []string{"--no-playlist"}
	cleanup := func() {}

	if y.jsRuntimes != "" {
		args = append(args, "--js-runtimes", y.jsRuntimes)
	}
	if y.remoteComponents != "" {
		args = append(args, "--remote-components", y.remoteComponents)
	}
	if y.cookiesFile != "" {
		cookiesPath, cookiesCleanup, err := prepareCookiesFile(y.cookiesFile)
		if err != nil {
			return nil, nil, fmt.Errorf("prepare yt-dlp cookies file: %w", err)
		}
		if cookiesPath != "" {
			cleanup = cookiesCleanup
			args = append(args, "--cookies", cookiesPath)
		}
	}
	return args, cleanup, nil
}

func prepareCookiesFile(sourcePath string) (string, func(), error) {
	info, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", func() {}, nil
		}
		return "", nil, err
	}
	if info.IsDir() {
		return "", nil, fmt.Errorf("cookies path is a directory: %s", sourcePath)
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return "", nil, err
	}
	defer source.Close()

	tempFile, err := os.CreateTemp("", "yt-dlp-cookies-*.txt")
	if err != nil {
		return "", nil, err
	}
	if _, err := io.Copy(tempFile, source); err != nil {
		tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return "", nil, err
	}
	if err := tempFile.Chmod(0o600); err != nil {
		tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return "", nil, err
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempFile.Name())
		return "", nil, err
	}

	return tempFile.Name(), func() {
		_ = os.Remove(tempFile.Name())
	}, nil
}

func buildProfiles(formats []probeFormat, durationSec int) []model.DownloadProfile {
	audioSize := bestAudioSize(formats)
	videoHeights := availableVideoHeights(formats)

	profiles := make([]model.DownloadProfile, 0, len(model.BuiltInDownloadProfiles()))
	for _, spec := range model.BuiltInDownloadProfiles() {
		profile := model.DownloadProfile{
			ID:           spec.ID,
			Kind:         spec.Kind,
			Label:        spec.Label,
			Container:    spec.Container,
			TargetHeight: spec.TargetHeight,
		}

		switch spec.Kind {
		case "audio":
			profile.Available = audioSize > 0 || hasAudio(formats)
			profile.EstimatedSizeBytes = audioSize
		case "video":
			videoSize, ok := bestVideoSizeForHeight(formats, spec.TargetHeight)
			profile.Available = ok && slices.Contains(videoHeights, spec.TargetHeight)
			if profile.Available {
				profile.EstimatedSizeBytes = videoSize + audioSize
			}
		}

		if profile.Kind == "audio" && profile.EstimatedSizeBytes == 0 && durationSec > 0 {
			profile.EstimatedSizeBytes = int64(durationSec) * 192000 / 8
		}
		profiles = append(profiles, profile)
	}

	return profiles
}

func availableVideoHeights(formats []probeFormat) []int {
	heightSet := make(map[int]struct{})
	for _, format := range formats {
		if !isVideoFormat(format) || format.Height == 0 {
			continue
		}
		heightSet[format.Height] = struct{}{}
	}

	heights := make([]int, 0, len(heightSet))
	for height := range heightSet {
		heights = append(heights, height)
	}
	slices.Sort(heights)
	return heights
}

func hasAudio(formats []probeFormat) bool {
	for _, format := range formats {
		if isAudioFormat(format) {
			return true
		}
	}
	return false
}

func bestAudioSize(formats []probeFormat) int64 {
	var best int64
	for _, format := range formats {
		if !isAudioFormat(format) {
			continue
		}
		size := usableSize(format)
		if size > best {
			best = size
		}
	}
	return best
}

func bestVideoSizeForHeight(formats []probeFormat, targetHeight int) (int64, bool) {
	var best probeFormat
	found := false

	for _, format := range formats {
		if !isVideoFormat(format) || format.Height != targetHeight {
			continue
		}
		if !found || usableSize(format) > usableSize(best) || format.Bitrate > best.Bitrate {
			best = format
			found = true
		}
	}

	if !found {
		return 0, false
	}
	return usableSize(best), true
}

func usableSize(format probeFormat) int64 {
	if format.FileSize > 0 {
		return format.FileSize
	}
	return format.FileSizeApprox
}

func isAudioFormat(format probeFormat) bool {
	return format.AudioCodec != "" && format.AudioCodec != "none"
}

func isVideoFormat(format probeFormat) bool {
	return format.VideoCodec != "" && format.VideoCodec != "none"
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

func removeInputArtifact(finalPath, outputDir string) error {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return fmt.Errorf("read output dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(outputDir, entry.Name())
		if path == finalPath {
			continue
		}
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("cleanup artifact %s: %w", path, err)
		}
	}

	return nil
}
