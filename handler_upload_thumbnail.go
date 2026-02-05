package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here
	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get thumbnail from form", err)
		return
	}
	defer file.Close()

	mediaType := header.Header.Get("Content-Type")
	if mediaType != "image/png" && mediaType != "image/jpeg" {
		respondWithError(w, http.StatusBadRequest, "Thumbnail must be a PNG or JPEG image", nil)
		return
	}

	metaData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get video metadata", err)
		return
	}

	if metaData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "You are not the uploader of this video", nil)
		return
	}

	var fileExtension string
	switch mediaType {
	case "image/png":
		fileExtension = ".png"
	case "image/jpeg":
		fileExtension = ".jpg"
	default:
		respondWithError(w, http.StatusBadRequest, "Thumbnail must be PNG or JPEG", nil)
		return
	}

	filename := videoID.String() + fileExtension

	filePath := filepath.Join(cfg.assetsRoot, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create thumbnail file", err)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save thumbnail file", err)
		return
	}

	thumbnailURL := fmt.Sprintf(
		"http://localhost:%s/assets/%s",
		cfg.port,
		filename,
	)

	metaData.ThumbnailURL = &thumbnailURL

	err = cfg.db.UpdateVideo(metaData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video metadata", err)
		return
	}

	respondWithJSON(w, http.StatusOK, metaData)
}
