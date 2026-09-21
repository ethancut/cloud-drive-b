package handler

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ethannself/cloud-drive-b/internal/auth"
	"github.com/ethannself/cloud-drive-b/internal/storage"
)

type AuthResponse struct {
	Status       string `json:"status"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Username     string `json:"username,omitempty"`
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello, World!\n"))
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
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

	tokens, username, err := auth.Register(storage.GetDataStore(), req)
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
func LoginHandler(w http.ResponseWriter, r *http.Request) {
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

	tokens, username, err := auth.Login(storage.GetDataStore(), req)
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

func JWTTestHandler(w http.ResponseWriter, r *http.Request) {
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
func UploadHandler(w http.ResponseWriter, r *http.Request) {
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

	storage.UploadFile(userID, fileHeader.Filename, file)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "uploaded"})
}

func ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	files, err := storage.ListFiles(userID)
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
func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "Missing filename", http.StatusBadRequest)
		return
	}
	if err := storage.DeleteFile(userID, filename); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	filename := r.PathValue("filename")
	if filename == "" {
		http.Error(w, "Missing filename", http.StatusBadRequest)
		return
	}
	cleanFilename := filepath.Base(filepath.Clean(filename))

	// reject any filename that tries to traverse directories or is invalid
	if cleanFilename == "." || cleanFilename == ".." || cleanFilename == "/" || cleanFilename != filename {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	// double check that the file is within the user's directory
	baseDir := filepath.Join("uploads", strconv.Itoa(userID))
	targetPath := filepath.Join(baseDir, cleanFilename)
	expectedPrefix := filepath.Clean(baseDir) + string(filepath.Separator)

	if !strings.HasPrefix(targetPath, expectedPrefix) {
		http.Error(w, "Access Denied", http.StatusBadRequest)
		return
	}

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
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": cleanFilename}))
	w.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))

	http.ServeContent(w, r, filename, stat.ModTime(), file)
}
func RefreshTokenhandler(w http.ResponseWriter, r *http.Request) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}
	newTokens, err := auth.RefreshAccessToken(authHeader)
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
