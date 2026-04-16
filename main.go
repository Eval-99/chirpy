package main

import (
	"net/http"
)

func main() {
	apiCfg := apiConfig{}
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("POST /api/validate_chirp", ValidateChirpHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.hitsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHitsHandler)

	serverStruct := http.Server{Handler: mux, Addr: ":8080"}
	serverStruct.ListenAndServe()
}
