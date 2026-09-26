package auth

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ethannself/cloud-drive-b/internal/database"
	"github.com/google/uuid"
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

func Register(ts *TokenService, dataStore *database.DataStore, req RegisterRequest) (*TokenPair, string, error) {
	var err error

	if req.RegistrationKey != os.Getenv("REGISTRATION_KEY") {
		log.Println("Register error: invalid registration key")
		return nil, "", fmt.Errorf("invalid registration key")
	}

	err = dataStore.AddUser(req.Username, req.Password, req.Email)
	if err != nil {
		log.Println("Register error:", err)
		return nil, "", err
	}
	return Login(ts, dataStore, req)
}

func Login(ts *TokenService, dataStore *database.DataStore, req RegisterRequest) (*TokenPair, string, error) {
	userID, username, err := dataStore.Login(req.Email, req.Password)
	if err != nil {
		log.Println("Login error:", err)
		return nil, "", err
	}
	if userID == uuid.Nil {
		log.Println("Login error: invalid credentials")
		return nil, "", fmt.Errorf("invalid credentials")
	}

	tokens, err := ts.GenerateTokenPair(userID)
	if err != nil {
		log.Println("JWT generation error:", err)
		return nil, "", err
	}

	return tokens, username, nil
}
func JWTMiddleware(ts *TokenService, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("JWTMiddleware: checking token for request to", r.URL.Path)
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		userID, err := ts.ValidateAccessToken(authHeader)
		if err != nil {
			log.Println("Access Token Validation Error:", err)
			http.Error(w, "Unauthorized "+err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDContextKey).(uuid.UUID)
	return userID, ok
}
