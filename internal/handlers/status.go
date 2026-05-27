package handlers

import (
	"net/http"

	"github.com/go-chi/render"
)

type StatusResponse struct {
	Status string `json:"status"`
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, StatusResponse{Status: "ok"})
}
