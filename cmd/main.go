package main

import (
	"fmt"
	"log"
	"net/http"
	"ty-server/internal/server"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		server.HandleConnections(w, r)
	})
	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
