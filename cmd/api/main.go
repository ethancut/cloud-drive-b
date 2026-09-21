package main

import (
	"log"
	"os"

	"github.com/ethannself/cloud-drive-b/internal/api/handler"
	"github.com/ethannself/cloud-drive-b/internal/auth"
	"github.com/ethannself/cloud-drive-b/internal/server"
	"github.com/joho/godotenv"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	tokenService, err := auth.NewTokenService(secret, auth.AccessTokenExpiry, auth.RefreshTokenExpiry)
	if err != nil {
		log.Fatal("Failed to create token service:", err)
	}
	h := handler.NewHandler(tokenService)

	server.Start(h)

}
