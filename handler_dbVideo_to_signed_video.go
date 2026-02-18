package main

import (
	"errors"
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	parts := strings.SplitN(*video.VideoURL, ",", 2)
	if len(parts) != 2 {
		return video, errors.New("invalid video_url format")
	}
	bucket := parts[0]
	key := parts[1]

	presignURL, err := generatePresignedURL(
		cfg.s3Client,
		bucket,
		key,
		15*time.Minute,
	)

	if err != nil {
		return video, err
	}
	video.VideoURL = &presignURL

	return video, nil
}
