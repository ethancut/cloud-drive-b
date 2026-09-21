package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ethannself/cloud-drive-b/internal/database"
)

type contextKey string
type InvalidRegistrationKeyError error

const UserIDContextKey contextKey = "userID"

type RegisterRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	Email           string `json:"email"`
	RegistrationKey string `json:"registration_key"`
}

func Register(dataStore *database.DataStore, req RegisterRequest) (*database.TokenPair, string, error) {
	var err error
	log.Printf("user.Register: got credentials: username: %s, email: %s, registration_key: %s", req.Username, req.Email, req.RegistrationKey)

	if req.RegistrationKey != os.Getenv("REGISTRATION_KEY") {
		log.Println("Register error: invalid registration key")
		return nil, "", fmt.Errorf("invalid registration key")
	}

	err = dataStore.AddUser(req.Username, req.Password, req.Email)
	if err != nil {
		log.Println("Register error:", err)
		return nil, "", err
	}
	return Login(dataStore, req)
}

func Login(dataStore *database.DataStore, req RegisterRequest) (*database.TokenPair, string, error) {
	userID, username, err := dataStore.Login(req.Email, req.Password)
	if err != nil {
		log.Println("Login error:", err)
		return nil, "", err
	}
	if userID == -1 {
		log.Println("Login error: invalid credentials")
		return nil, "", fmt.Errorf("invalid credentials")
	}

	tokens, err := database.GenerateTokenPair(userID)
	if err != nil {
		log.Println("JWT generation error:", err)
		return nil, "", err
	}

	return tokens, username, nil
}
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("JWTMiddleware: checking token for request to", r.URL.Path)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		userID, err := database.ValidateAccessToken(authHeader)
		if err != nil {
			log.Println("Access Token Validation Error:", err)
			http.Error(w, "Unauthorized "+err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(int)
	return userID, ok
}
