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
	router.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {

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
	router.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
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

		isLoggedIn, err := database.Login(Pool, req.Email, req.Password)
		if err != nil {
			log.Println("Login error:", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if !isLoggedIn {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"logged_in"}`))
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
