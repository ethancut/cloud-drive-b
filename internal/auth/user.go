package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ethannself/cloud-drive-b/internal/database"
	"github.com/golang-jwt/jwt/v5"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
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
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
