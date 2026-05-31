package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ethannself/cloud-drive-b/internal/database"
	"github.com/joho/godotenv"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
}
type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	Pool := database.New()
	defer Pool.Close()
	router := http.NewServeMux()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!\n"))
	})
	router.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req RegisterRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Bad JSON", http.StatusBadRequest)
			return
		}

		fmt.Println(req.Username, req.Password, req.Email)
		err = database.AddUser(Pool, req.Username, req.Password, req.Email)
		if err != nil {
			log.Println("Register error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	router.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req RegisterRequest
		err = json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Bad JSON", http.StatusBadRequest)
			return
		}

		userID, err := database.Login(Pool, req.Email, req.Password)
		fmt.Println("err: ", err)
		if err != nil {
			log.Println("Login error:", err)
			if err == database.AccountNotFoundError || err == database.InvalidPasswordError {
				fmt.Println("sending 401")
				sendError(w, http.StatusUnauthorized, "Invalid email or password")
				return
			} else {
				fmt.Println("sending 500")
				sendError(w, http.StatusInternalServerError, "Internal server error")
				return
			}
		}

		if userID == -1 {
			sendError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		token, err := database.GenerateJWT(userID)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "Failed to create token")
			return
		}
		response := map[string]string{
			"status": "logged_in",
			"token":  token,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	})

	http.ListenAndServe(":8080", corsMiddleware(router))
}
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func sendError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{
		Message: message,
	})
}
