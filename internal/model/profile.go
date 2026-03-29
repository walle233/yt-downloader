package model

type DownloadProfileID string

const (
	ProfileVideo360p  DownloadProfileID = "video-360p"
	ProfileVideo480p  DownloadProfileID = "video-480p"
	ProfileVideo720p  DownloadProfileID = "video-720p"
	ProfileVideo1080p DownloadProfileID = "video-1080p"
	ProfileAudioMP3   DownloadProfileID = "audio-mp3"
)

type DownloadProfileSpec struct {
	ID           DownloadProfileID
	Kind         string
	Label        string
	Container    string
	TargetHeight int
}

type DownloadProfile struct {
	ID                 DownloadProfileID `json:"id"`
	Kind               string            `json:"kind"`
	Label              string            `json:"label"`
	Container          string            `json:"container"`
	TargetHeight       int               `json:"targetHeight,omitempty"`
	Available          bool              `json:"available"`
	EstimatedSizeBytes int64             `json:"estimatedSizeBytes,omitempty"`
}

var builtInDownloadProfiles = []DownloadProfileSpec{
	{ID: ProfileVideo360p, Kind: "video", Label: "360p MP4", Container: "mp4", TargetHeight: 360},
	{ID: ProfileVideo480p, Kind: "video", Label: "480p MP4", Container: "mp4", TargetHeight: 480},
	{ID: ProfileVideo720p, Kind: "video", Label: "720p MP4", Container: "mp4", TargetHeight: 720},
	{ID: ProfileVideo1080p, Kind: "video", Label: "1080p MP4", Container: "mp4", TargetHeight: 1080},
	{ID: ProfileAudioMP3, Kind: "audio", Label: "MP3 Audio", Container: "mp3"},
}

func BuiltInDownloadProfiles() []DownloadProfileSpec {
	profiles := make([]DownloadProfileSpec, len(builtInDownloadProfiles))
	copy(profiles, builtInDownloadProfiles)
	return profiles
}

func FindDownloadProfileSpec(id DownloadProfileID) (DownloadProfileSpec, bool) {
	for _, profile := range builtInDownloadProfiles {
		if profile.ID == id {
			return profile, true
		}
	}
	return DownloadProfileSpec{}, false
}

func FindAvailableProfile(profiles []DownloadProfile, id DownloadProfileID) (DownloadProfile, bool) {
	for _, profile := range profiles {
		if profile.ID == id && profile.Available {
			return profile, true
		}
	}
	return DownloadProfile{}, false
}
