package main

import (
	"net/http"
	"os"

	"log/slog"

	"github.com/pix303/localemgmt-go/api/pkg/router"
)

func main() {
	return
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	r, err := router.NewRouter()
	if err != nil {
		slog.Error("Error on startup router", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Starting server on port 8080")
	slog.Error("Error: %s", "error", http.ListenAndServe(":8080", r.Router).Error())
}
