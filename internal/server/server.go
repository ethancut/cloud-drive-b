package server

import (
	"log"
	"net/http"

	"github.com/ethannself/cloud-drive-b/internal/api/handler"
	"github.com/ethannself/cloud-drive-b/internal/api/router"
	"github.com/ethannself/cloud-drive-b/internal/storage"
)

func Start() {
	storage.InitDataStore()
	dataStore := storage.GetDataStore()
	defer dataStore.Close()

	router := router.New()

	log.Println("Server is running on port 8080")
	err := http.ListenAndServe(":8080", handler.CorsMiddleware(router))
	if err != nil {
		log.Fatal(err)
	}
}
