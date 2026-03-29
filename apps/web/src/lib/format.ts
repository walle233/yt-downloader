import type { DownloadProfile, DownloadProfileId, MediaKind } from "@ytvd/shared-types";

const PROFILE_ORDER: DownloadProfileId[] = ["video-1080p", "video-720p", "video-480p", "video-360p", "audio-mp3"];

export function sortProfiles(profiles: DownloadProfile[]): DownloadProfile[] {
  return [...profiles].sort((left, right) => PROFILE_ORDER.indexOf(left.id) - PROFILE_ORDER.indexOf(right.id));
}

export function groupProfiles(profiles: DownloadProfile[]): Record<MediaKind, DownloadProfile[]> {
  return {
    video: sortProfiles(profiles.filter((profile) => profile.kind === "video")),
    audio: sortProfiles(profiles.filter((profile) => profile.kind === "audio")),
  };
}

export function profileHeadline(profileId: DownloadProfileId, targetHeight?: number): string {
  switch (profileId) {
    case "video-1080p":
    case "video-720p":
    case "video-480p":
    case "video-360p":
      return `${targetHeight ?? profileId.replace("video-", "")} MP4`;
    case "audio-mp3":
      return "MP3 Audio";
    default:
      return profileId;
  }
}

export function profileSubline(mediaKind: MediaKind, targetHeight?: number): string {
  if (mediaKind === "audio") {
    return "Audio only";
  }
  return `${targetHeight ?? ""}p preset`;
}

export function formatBytes(value?: number): string {
  if (!value || value <= 0) {
    return "Size generated after processing";
  }

  const units = ["B", "KB", "MB", "GB"];
  let size = value;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit += 1;
  }
  return `${size.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
}

export function formatDuration(seconds?: number): string {
  if (!seconds || seconds <= 0) {
    return "Unknown";
  }
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const remaining = seconds % 60;

  if (hours > 0) {
    return `${hours}:${String(minutes).padStart(2, "0")}:${String(remaining).padStart(2, "0")}`;
  }
  return `${minutes}:${String(remaining).padStart(2, "0")}`;
}

export function statusLabel(status: string): string {
  switch (status) {
    case "queued":
      return "Queued";
    case "downloading":
      return "Downloading";
    case "uploading":
      return "Packaging";
    case "completed":
      return "Ready";
    case "failed":
      return "Failed";
    default:
      return status;
  }
}
