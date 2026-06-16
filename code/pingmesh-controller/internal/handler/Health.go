package handler

import (
	"net/http"
	"time"

	"pingmesh-controller/internal/common"
)

func Health(w http.ResponseWriter, r *http.Request) {
	common.WriteJSON(w, map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}
