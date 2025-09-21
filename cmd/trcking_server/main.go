package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Tracking Server is running!")
	})

	log.Println("Tracking Server starting on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		log.Fatal("Failed to start Tracking server:", err)
	}
}