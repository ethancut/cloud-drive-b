package router

import (
	"net/http"

	"github.com/ethannself/cloud-drive-b/internal/api/handler"
	"github.com/rs/cors"
)

func New() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler.DefaultHandler)
	mux.HandleFunc("/api/register", handler.RegisterHandler)
	mux.HandleFunc("/api/login", handler.LoginHandler)

	router := cors.Default().Handler(mux)
	return router
}
