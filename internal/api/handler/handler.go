package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"strconv"

	"github.com/ethannself/cloud-drive-b/internal/auth"
	"github.com/ethannself/cloud-drive-b/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AuthResponse struct {
	Status       string `json:"status"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username,omitempty"`
}
type Handler struct {
	TokenService *auth.TokenService
}

func NewHandler(ts *auth.TokenService) *Handler {
	return &Handler{TokenService: ts}
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!\n"))
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.Password == "" || req.Email == "" || req.RegistrationKey == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	tokens, username, err := auth.Register(h.TokenService, storage.GetDataStore(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := AuthResponse{
		Status:       "ok",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		Username:     username,
	}
	json.NewEncoder(w).Encode(response)
}
func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}
	fmt.Println("Attempting to login with credentials: ", req.Email, req.Password)

	tokens, username, err := auth.Login(h.TokenService, storage.GetDataStore(), req)
	if err != nil {
		http.Error(w, "Invalid Credentials", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Status:       "logged_in",
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		Username:     username,
	})
}

func (h *Handler) JWTTestHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "token_valid",
		"userID": userID,
	})
}
func (h *Handler) UploadHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	id, err := storage.UploadFile(r.Context(), userID, fileHeader.Filename, file)
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "uploaded", "id": id.String()})
}

func (h *Handler) ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	files, err := storage.ListFiles(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to list files", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"files":  files,
	})
}
func (h *Handler) DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	fileID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid file id", http.StatusBadRequest)
		return
	}
	if err := storage.DeleteFile(r.Context(), userID, fileID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	idStr := r.PathValue("id")

	if idStr == "" {
		http.Error(w, "Missing file id", http.StatusBadRequest)
		return
	}
	// reject any id that tries to traverse directories or is invalid
	if idStr == "." || idStr == ".." || idStr == "/" {
		http.Error(w, "Invalid file id", http.StatusBadRequest)
		return
	}
	log.Println("Requested file ID:", idStr)
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid file id", http.StatusBadRequest)
		return
	}

	db := storage.GetDataStore()

	metadata, err := db.GetFileMetadata(r.Context(), id, userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to get file metadata", http.StatusInternalServerError)
		return
	}

	targetPath := metadata.FilePath

	file, err := os.Open(targetPath)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		http.Error(w, "could not stat file", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Security-Policy", "default-src 'none'; sandbox")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": metadata.OriginalFilename}))
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	http.ServeContent(w, r, metadata.OriginalFilename, stat.ModTime(), file)
}

func (h *Handler) RefreshTokenhandler(w http.ResponseWriter, r *http.Request) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}
	newTokens, _, err := h.TokenService.RotateRefreshToken(authHeader)
	if err != nil {
		http.Error(w, "Failed to refresh token: "+err.Error(), http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(AuthResponse{
		Status:       "token_refreshed",
		AccessToken:  newTokens.AccessToken,
		RefreshToken: newTokens.RefreshToken,
	})
}
