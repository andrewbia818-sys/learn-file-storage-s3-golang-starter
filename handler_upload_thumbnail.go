package main

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "io"
    "mime"
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

    const maxMemory = 10 << 20
    r.ParseMultipartForm(maxMemory)

    file, header, err := r.FormFile("thumbnail")
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
        return
    }
    defer file.Close()

    // Parse media type
    contentType := header.Header.Get("Content-Type")
    mediaType, _, err := mime.ParseMediaType(contentType)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid Content-Type header", err)
        return
    }

    // Only allow JPEG and PNG
    if mediaType != "image/jpeg" && mediaType != "image/png" {
        respondWithError(w, http.StatusBadRequest, "Only JPEG and PNG thumbnails are allowed", nil)
        return
    }

    // Read image bytes
    imageData, err := io.ReadAll(file)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Unable to read image data", err)
        return
    }

    // Fetch video metadata
    video, err := cfg.db.GetVideo(videoID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Video not found", err)
        return
    }

    // Ensure user owns the video
    if video.UserID != userID {
        respondWithError(w, http.StatusUnauthorized, "You do not own this video", nil)
        return
    }

    // Determine file extension
    var fileExt string
    if mediaType == "image/jpeg" {
        fileExt = "jpg"
    } else {
        fileExt = "png"
    }

    // Generate 32 random bytes
    randomBytes := make([]byte, 32)
    _, err = rand.Read(randomBytes)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Unable to generate random filename", err)
        return
    }

    // Convert to base64 URL-safe string
    randomName := base64.RawURLEncoding.EncodeToString(randomBytes)

    // Build file path: /assets/<randomName>.<ext>
    filename := fmt.Sprintf("%s.%s", randomName, fileExt)
    filePath := filepath.Join(cfg.assetsRoot, filename)

    // Create file on disk
    outFile, err := os.Create(filePath)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Unable to create file on disk", err)
        return
    }
    defer outFile.Close()

    // Write image bytes to disk
    _, err = outFile.Write(imageData)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Unable to write file to disk", err)
        return
    }

    // Update thumbnail_url
    thumbURL := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, filename)
    video.ThumbnailURL = &thumbURL

    // Update DB record
    err = cfg.db.UpdateVideo(video)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Unable to update video", err)
        return
    }

    respondWithJSON(w, http.StatusOK, video)
}


/*package main

import (
    "fmt"
    "io"
    "mime"
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

    const maxMemory = 10 << 20
    r.ParseMultipartForm(maxMemory)

    file, header, err := r.FormFile("thumbnail")
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
        return
    }
    defer file.Close()

    // Parse media type using mime.ParseMediaType
    contentType := header.Header.Get("Content-Type")
    mediaType, _, err := mime.ParseMediaType(contentType)
    if err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid Content-Type header", err)
        return
    }

    // Only allow image/jpeg and image/png
    if mediaType != "image/jpeg" && mediaType != "image/png" {
        respondWithError(w, http.StatusBadRequest, "Only JPEG and PNG thumbnails are allowed", nil)
        return
    }

    // Read image bytes
    imageData, err := io.ReadAll(file)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Unable to read image data", err)
        return
    }

    // Fetch video metadata
    video, err := cfg.db.GetVideo(videoID)
    if err != nil {
        respondWithError(w, http.StatusNotFound, "Video not found", err)
        return
    }

    // Ensure user owns the video
    if video.UserID != userID {
        respondWithError(w, http.StatusUnauthorized, "You do not own this video", nil)
        return
    }

    // Determine file extension
    var fileExt string
    if mediaType == "image/jpeg" {
        fileExt = "jpg"
    } else {
        fileExt = "png"
    }

    // Build file path: /assets/<videoID>.<ext>
    filename := fmt.Sprintf("%s.%s", videoID.String(), fileExt)
    filePath := filepath.Join(cfg.assetsRoot, filename)

    // Create file on disk
    outFile, err := os.Create(filePath)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Unable to create file on disk", err)
        return
    }
    defer outFile.Close()

    // Write image bytes to disk
    _, err = outFile.Write(imageData)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Unable to write file to disk", err)
        return
    }

    // Update thumbnail_url to point to the asset
    thumbURL := fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, filename)
    video.ThumbnailURL = &thumbURL

    // Update DB record
    err = cfg.db.UpdateVideo(video)
    if err != nil {
        respondWithError(w, http.StatusInternalServerError, "Unable to update video", err)
        return
    }

    respondWithJSON(w, http.StatusOK, video)
}
*/