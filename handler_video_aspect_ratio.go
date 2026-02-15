package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func (cfg *apiConfig) getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-print_format", "json",
		"-show_streams",
		filePath,
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return "", err
	}

	type ffprobeOutput struct {
		Streams []struct {
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			CodecType string `json:"codec_type"`
		} `json:"streams"`
	}

	var probeData ffprobeOutput

	if err := json.Unmarshal(out.Bytes(), &probeData); err != nil {
		return "", err
	}

	if len(probeData.Streams) == 0 {
		return "", fmt.Errorf("no streams found")
	}

	// Find the first video stream
	var width, height int
	for _, stream := range probeData.Streams {
		if stream.CodecType == "video" {
			width = stream.Width
			height = stream.Height
			break
		}
	}

	if width == 0 || height == 0 {
		return "", fmt.Errorf("invalid video dimensions")
	}

	// Simple orientation check
	if width > height {
		return "16:9", nil // landscape
	}

	if height > width {
		return "9:16", nil // portrait
	}

	return "other", nil

}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
