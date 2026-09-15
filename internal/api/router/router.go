package router

import (
	"net/http"

	"github.com/ethannself/cloud-drive-b/internal/api/handler"
	"github.com/ethannself/cloud-drive-b/internal/auth"
	"github.com/rs/cors"
)

func New() http.Handler {
	mux := http.NewServeMux()

	// mux.HandleFunc("/", handler.DefaultHandler)
	mux.HandleFunc("/api/register", handler.RegisterHandler)
	mux.HandleFunc("/api/login", handler.LoginHandler)
	mux.Handle("/api/test-jwt", auth.JWTMiddleware(http.HandlerFunc(handler.JWTTestHandler)))
	mux.Handle("/api/files/upload", auth.JWTMiddleware(http.HandlerFunc(handler.UploadHandler)))
	mux.Handle("/api/files/list", auth.JWTMiddleware(http.HandlerFunc(handler.ListFilesHandler)))
	mux.Handle("/api/files/delete/{filename}", auth.JWTMiddleware(http.HandlerFunc(handler.DeleteFileHandler)))
	mux.Handle("/api/files/download/{filename}", auth.JWTMiddleware(http.HandlerFunc(handler.DownloadFileHandler)))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://cloud.ethann.stream", "http://localhost:4321"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})
	router := c.Handler(mux)
	return router
}
