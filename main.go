package main

import (
	"log"
	"net/http"
	"os"

	"github.com/klaus858/ddslot777-api/backend"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("%s listening on :%s", backend.ServiceName, port)
	log.Fatal(http.ListenAndServe(":"+port, backend.NewHandler()))
}
