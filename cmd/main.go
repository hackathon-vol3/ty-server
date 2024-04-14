package main

import (
	"fmt"
	"log"
	"net/http"
	"ty-server/internal/server"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
r.Use(func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        origin := r.Header.Get("Origin")
        if origin == "http://localhost:3000" || origin == "http://localhost:3001" {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            w.Header().Set("Access-Control-Allow-Credentials", "true")

            // Handle preflight requests
            if r.Method == "OPTIONS" {
                w.WriteHeader(http.StatusOK)
                return
            }
        }

        next.ServeHTTP(w, r)
    })
})
	r.HandleFunc("/", server.HandleConnections)
	r.HandleFunc("/signup", server.SignupPage)
	r.HandleFunc("/login", server.LoginPage)
	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
