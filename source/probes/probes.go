package probes

import (
	"log/slog"
	"net/http"
	"os"
)

var server http.Server

func init() {
	r := http.NewServeMux()
	r.HandleFunc("/healthz", liveness)
	r.HandleFunc("/readyz", readiness)

	server = http.Server{
		Addr:    ":8181",
		Handler: r,
	}
}

func Run(logger *slog.Logger) {
	if err := server.ListenAndServe(); err != nil {
		logger.Error("The probes HTTP server has crashed...", "error", err)
		os.Exit(3)
	}
}

func liveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readiness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
