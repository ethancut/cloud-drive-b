package main

import (
	"log"

	"github.com/ethannself/cloud-drive-b/internal/server"
	"github.com/joho/godotenv"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}
	server.Start()

}
