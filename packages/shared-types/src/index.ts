export type OutputFormat = "source" | "mp4";

export interface ProbeResponse {
  videoId: string;
  title: string;
  durationSec: number;
  thumbnailUrl: string;
  allowedFormats: OutputFormat[];
  message?: string;
}

export interface CreateDownloadResponse {
  jobId: string;
  status: string;
  message?: string;
}

export interface DownloadStatusResponse {
  jobId: string;
  status: string;
  progress: number;
  step: string;
  message?: string;
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
  outputFormat: string;
  fileName?: string;
  thumbnailUrl?: string;
  createdAt: string;
}

export interface DownloadListResponse {
  items: DownloadListItem[];
}
