package handler

import (
	"net/http"

	"github.com/klaus858/ddslot777-api/backend"
)

var apiHandler = backend.NewHandler()

func Handler(w http.ResponseWriter, r *http.Request) {
	apiHandler.ServeHTTP(w, r)
}
