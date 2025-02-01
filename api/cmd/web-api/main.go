package main

import (
	"log"
	"net/http"
	"os"

	"log/slog"

	"github.com/pix303/localemgmt-go/api/pkg/router"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	r, err := router.NewRouter()
	if err != nil {
		log.Fatalf("Error on startup router: %v", err)
	}

	slog.Info("Starting server on port 8080")
	slog.Error("Error: %s", "error", http.ListenAndServe(":8080", r.Router).Error())
}
