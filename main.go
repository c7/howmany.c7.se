package main

//go:generate ./convert.sh

import (
	"fmt"
	"net/http"
	"os"
)

const defaultPort = "8000"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	http.HandleFunc("/", Hello)

	fmt.Println("Listening on port", port)
	http.ListenAndServe(":"+port, nil)
}

func Hello(w http.ResponseWriter, r *http.Request) {
	name := "World"

	if p := r.URL.Path[1:]; p != "" {
		name = p
	}

	fmt.Fprintf(w, "Hello, %s!", name)
}
