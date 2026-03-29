export type DownloadProfileId = "video-360p" | "video-480p" | "video-720p" | "video-1080p" | "audio-mp3";

export type MediaKind = "video" | "audio";

export interface DownloadProfile {
  id: DownloadProfileId;
  kind: MediaKind;
  label: string;
  container: "mp4" | "mp3";
  targetHeight?: number;
  available: boolean;
  estimatedSizeBytes?: number;
}

export interface ProbeResponse {
  videoId: string;
  title: string;
  durationSec: number;
  thumbnailUrl: string;
  profiles: DownloadProfile[];
  message?: string;
}

export interface CreateDownloadResponse {
  jobId: string;
  status: string;
  profileId: DownloadProfileId;
  message?: string;
}

export interface DownloadStatusResponse {
  jobId: string;
  status: string;
  progress: number;
  step: string;
  message?: string;
  title: string;
  thumbnailUrl?: string;
  durationSec?: number;
  profileId: DownloadProfileId;
  mediaKind: MediaKind;
  targetHeight?: number;
  fileName?: string;
  fileSize?: number;
  createdAt: string;
}

export interface DownloadResultResponse {
  jobId: string;
  status: string;
  fileName: string;
  fileSize: number;
  downloadUrl: string;
}

export interface DownloadListItem {
  jobId: string;
  title: string;
  status: string;
  profileId: DownloadProfileId;
  mediaKind: MediaKind;
  targetHeight?: number;
  fileName?: string;
  thumbnailUrl?: string;
  createdAt: string;
}

export interface DownloadListResponse {
  items: DownloadListItem[];
}
