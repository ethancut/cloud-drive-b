package router

import (
	"net/http"

	"github.com/ethannself/cloud-drive-b/internal/api/handler"
)

func New() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handler.DefaultHandler)
	mux.HandleFunc("/api/register", handler.RegisterHandler)
	mux.HandleFunc("/api/login", handler.LoginHandler)
	return mux
}
