package probes

import "net/http"

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

func Run(eC chan error) {
	eC <- server.ListenAndServe()
}

func liveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func readiness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
