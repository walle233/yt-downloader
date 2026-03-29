package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	clerk "github.com/clerk/clerk-sdk-go/v2"
	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"

	"github.com/walle233/yt-downloader/internal/config"
	"github.com/walle233/yt-downloader/internal/model"
	"github.com/walle233/yt-downloader/internal/service"
)

type Server struct {
	config  config.Config
	runtime *service.Runtime
}

func New(cfg config.Config, runtime *service.Runtime) *Server {
	return &Server{config: cfg, runtime: runtime}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/healthz", s.handleHealthz)
	mux.HandleFunc("/api/v1/videos/probe", s.handleProbeVideo)
	mux.Handle("/api/v1/downloads", s.withAuthorization(http.HandlerFunc(s.handleDownloads)))
	mux.Handle("/api/v1/downloads/", s.withAuthorization(http.HandlerFunc(s.handleDownloadStatus)))
	return withCORS(withLogging(mux))
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.runtime.HealthCheck(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "error",
			"error":  err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, model.HealthResponse{
		Status:      "ok",
		Environment: s.config.AppEnv,
	})
}

func (s *Server) handleProbeVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	var req model.VideoProbeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	if !looksLikeYouTubeURL(req.URL) {
		writeError(w, http.StatusBadRequest, "only youtube.com and youtu.be URLs are supported")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	probe, err := s.runtime.Downloader.Probe(ctx, req.URL)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.VideoProbeResponse{
		VideoID:      probe.VideoID,
		Title:        probe.Title,
		DurationSec:  probe.DurationSec,
		ThumbnailURL: probe.ThumbnailURL,
		Profiles:     probe.Profiles,
	})
}

func (s *Server) handleDownloads(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleListDownloads(w, r)
	case http.MethodPost:
		s.handleCreateDownload(w, r)
	default:
		w.Header().Set("Allow", "GET, POST")
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleCreateDownload(w http.ResponseWriter, r *http.Request) {
	clerkUserID, ok := currentClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req model.CreateDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	if !looksLikeYouTubeURL(req.URL) {
		writeError(w, http.StatusBadRequest, "only youtube.com and youtu.be URLs are supported")
		return
	}

	if _, ok := model.FindDownloadProfileSpec(model.DownloadProfileID(req.ProfileID)); !ok {
		writeError(w, http.StatusBadRequest, "invalid download profile")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	download, err := s.runtime.CreateDownload(ctx, clerkUserID, req.URL, model.DownloadProfileID(req.ProfileID))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrProfileUnavailable):
			writeError(w, http.StatusBadRequest, "requested download profile is not available for this video")
			return
		default:
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
	}

	writeJSON(w, http.StatusAccepted, model.CreateDownloadResponse{
		JobID:     download.JobID,
		Status:    download.Status,
		ProfileID: download.ProfileID,
		Message:   "Download queued successfully.",
	})
}

func (s *Server) handleListDownloads(w http.ResponseWriter, r *http.Request) {
	clerkUserID, ok := currentClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	items, err := s.runtime.ListRecentDownloads(ctx, clerkUserID, 20)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := make([]model.DownloadListItem, 0, len(items))
	for _, item := range items {
		response = append(response, model.DownloadListItem{
			JobID:        item.JobID,
			Title:        item.Title,
			Status:       item.Status,
			ProfileID:    item.ProfileID,
			MediaKind:    item.MediaKind,
			TargetHeight: item.TargetHeight,
			FileName:     item.FileName,
			ThumbnailURL: item.ThumbnailURL,
			CreatedAt:    item.CreatedAt.In(s.config.AppLocation).Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, model.DownloadListResponse{
		Items: response,
	})
}

func (s *Server) handleDownloadStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	clerkUserID, ok := currentClerkUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/api/v1/downloads/")
	if jobID == "" || jobID == r.URL.Path {
		writeError(w, http.StatusBadRequest, "missing job id")
		return
	}

	if strings.HasSuffix(jobID, "/result") {
		actualJobID := strings.TrimSuffix(jobID, "/result")
		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		download, resultURL, err := s.runtime.GetResultURL(ctx, clerkUserID, actualJobID)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeJSON(w, http.StatusOK, model.DownloadResultResponse{
			JobID:       download.JobID,
			Status:      download.Status,
			FileName:    download.FileName,
			FileSize:    download.FileSize,
			DownloadURL: resultURL,
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	download, err := s.runtime.GetStatus(ctx, clerkUserID, jobID)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.DownloadStatusResponse{
		JobID:        download.JobID,
		Status:       download.Status,
		Progress:     download.Progress,
		Step:         download.Step,
		Message:      download.ErrorMessage,
		Title:        download.Title,
		ThumbnailURL: download.ThumbnailURL,
		DurationSec:  download.DurationSec,
		ProfileID:    download.ProfileID,
		MediaKind:    download.MediaKind,
		TargetHeight: download.TargetHeight,
		FileName:     download.FileName,
		FileSize:     download.FileSize,
		CreatedAt:    download.CreatedAt.In(s.config.AppLocation).Format(time.RFC3339),
	})
}

func (s *Server) withAuthorization(next http.Handler) http.Handler {
	return clerkhttp.WithHeaderAuthorization(
		clerkhttp.AuthorizationFailureHandler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			writeError(w, http.StatusUnauthorized, "authentication required")
		})),
	)(next)
}

func currentClerkUserID(ctx context.Context) (string, bool) {
	claims, ok := clerk.SessionClaimsFromContext(ctx)
	if !ok || claims == nil || claims.Subject == "" {
		return "", false
	}
	return claims.Subject, true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func writeMethodNotAllowed(w http.ResponseWriter, allowed string) {
	w.Header().Set("Allow", allowed)
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func looksLikeYouTubeURL(url string) bool {
	return strings.Contains(url, "youtube.com/watch?v=") || strings.Contains(url, "youtu.be/")
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[%s] %s %s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
