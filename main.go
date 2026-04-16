package main

import (
	"net/http"
)

func main() {
	apiCfg := apiConfig{}
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/metrics", apiCfg.hitsHandler)
	mux.HandleFunc("/reset", apiCfg.resetHitsHandler)
	serverStruct := http.Server{Handler: mux, Addr: ":8080"}
	serverStruct.ListenAndServe()
}

func healthHandler(writter http.ResponseWriter, request *http.Request) {
	writter.Header().Set("Content-Type", "text/plain; charset=utf-8")
	writter.WriteHeader(http.StatusOK)
	writter.Write([]byte("OK"))
}
