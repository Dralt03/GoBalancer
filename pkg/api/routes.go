package api

import "net/http"

func Routes(h *Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", h.HealthCheck)
	mux.HandleFunc("/backends", h.GetBackends)
	mux.HandleFunc("/backends/", h.BackendByAddress)

	var handler http.Handler = mux
	handler = LoggingMiddleware(handler)
	handler = RecoveryMiddleware(handler)
	return handler
}
