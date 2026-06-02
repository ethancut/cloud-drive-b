package auth

import (
	"fmt"
	"log"

	"github.com/ethannself/cloud-drive-b/internal/database"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}

func Register(dataStore *database.DataStore, req RegisterRequest) error {
	var err error
	log.Println("user.Register: got credentials:", req.Username, req.Password, req.Email)

	err = dataStore.AddUser(req.Username, req.Password, req.Email)
	if err != nil {
		log.Println("Register error:", err)
		return err
	}
	return nil
}

func Login(dataStore *database.DataStore, req RegisterRequest) (string, error) {
	userID, err := dataStore.Login(req.Email, req.Password)
	if err != nil {
		log.Println("Login error:", err)
		return "", err
	}
	if userID == -1 {
		log.Println("Login error: invalid credentials")
		return "", fmt.Errorf("invalid credentials")
	}

	token, err := database.GenerateJWT(userID)
	if err != nil {
		log.Println("JWT generation error:", err)
		return "", err
	}

	return token, nil
}
