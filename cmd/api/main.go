package main

import (
	"log"
	"net/http"

	"github.com/ethannself/cloud-drive-b/internal/database"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	Pool := database.New()
	defer Pool.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!\n"))
	})
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Handle user registration logic here
		w.Write([]byte("User registration endpoint\n"))
	})
	http.ListenAndServe(":8080", nil)
}
