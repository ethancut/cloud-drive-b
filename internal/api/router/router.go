package router

import (
	"net/http"

	"github.com/ethannself/cloud-drive-b/internal/api/handler"
	"github.com/ethannself/cloud-drive-b/internal/auth"
	"github.com/rs/cors"
)

func New(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	// mux.HandleFunc("/", handler.DefaultHandler)
	mux.HandleFunc("/api/register", h.RegisterHandler)
	mux.HandleFunc("/api/login", h.LoginHandler)
	mux.Handle("/api/test-jwt", auth.JWTMiddleware(h.TokenService, http.HandlerFunc(h.JWTTestHandler)))
	mux.Handle("/api/files/upload", auth.JWTMiddleware(h.TokenService, http.HandlerFunc(h.UploadHandler)))
	mux.Handle("/api/files/list", auth.JWTMiddleware(h.TokenService, http.HandlerFunc(h.ListFilesHandler)))
	mux.Handle("/api/files/delete/{id}", auth.JWTMiddleware(h.TokenService, http.HandlerFunc(h.DeleteFileHandler)))
	mux.Handle("/api/files/download/{id}", auth.JWTMiddleware(h.TokenService, http.HandlerFunc(h.DownloadFileHandler)))
	mux.Handle("/api/auth/refresh", http.HandlerFunc(h.RefreshTokenhandler))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://cloud.ethann.stream", "http://localhost:4321"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	router := c.Handler(mux)
	return router
}
