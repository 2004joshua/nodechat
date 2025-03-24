package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	hub := NewHub()
	go hub.Run()

	r := mux.NewRouter()
	r.HandleFunc("/ws", hub.HandleWS)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("ui/build")))

	log.Printf("NodeChat listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
