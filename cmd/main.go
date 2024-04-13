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
	r.HandleFunc("/", server.HandleConnections)
	r.HandleFunc("/signup", server.SignupPage)
	r.HandleFunc("/login", server.LoginPage)
	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
