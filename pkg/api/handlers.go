package api

import (
	"encoding/json"
	"net/http"

	"github.com/rithish279/resilient-lb/pkg/lb"
)

type BackendStatus struct {
	URL         	string `json:"url"`
	Healthy     	bool   `json:"healthy"`
	ActiveConns 	int64  `json:"active_connections"`
	Weight      	int    `json:"weight"`
	CircuitBreaker 	string `json:"circuit_breaker"`
}

type AdminHandler struct {
	backends []*lb.Backend
}

func NewAdminHandler(backends []*lb.Backend) *AdminHandler {
	return &AdminHandler{backends: backends}
}

func (h *AdminHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (h *AdminHandler) handleBackends(w http.ResponseWriter, r *http.Request) {
	statuses := make([]BackendStatus, len(h.backends))
	for i, b := range h.backends {
		statuses[i] = BackendStatus{
			URL: b.URL.String(),
			Healthy: b.Healthy.Load(),
			ActiveConns: b.ActiveConns(),
			Weight: b.Weight,
			CircuitBreaker: b.CircuitBreaker.State(),
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statuses)
}

func (h *AdminHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/backends", h.handleBackends)
}