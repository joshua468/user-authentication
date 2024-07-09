package handler

import (
	"log"
	"net/http"

	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables from .env file in config folder
	if err := godotenv.Load("./config/.env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

// Handler is the exported function for Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func main() {
	http.HandleFunc("/", Handler)
	log.Fatal(http.ListenAndServe(":3000", nil))
}
