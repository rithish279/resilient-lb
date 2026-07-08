package api

import (
	"encoding/json"
	"net/http"

	"github.com/rithish279/resilient-lb/pkg/chaos"
)

type ChaosHandler struct {
	engine *chaos.Engine
}

func NewChaosHandler(engine *chaos.Engine) *ChaosHandler {
	return &ChaosHandler{engine: engine}
}

func (h *ChaosHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/chaos/inject", h.handleInject)
	mux.HandleFunc("/chaos/status", h.handleStatus)
	mux.HandleFunc("/chaos/clear", h.handleClear)
}

type injectRequest struct {
	Type 		string 		`json:"type"`
	Target 		string 		`json:"target"`
	Probability	float64		`json:"probability"`
	LatencyMs	int 		`json:"latency_ms"`
	ErrorCode 	int 		`json:"error_code"`
	DurationSec	int 		`json:"duration_sec"`
}

func (h *ChaosHandler) handleInject(w http.ResponseWriter, r *http.Request) {
	if (r.Method != http.MethodPost) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return 
	}

	var req injectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if (req.Probability == 0) {
		req.Probability = 1.0
	}

	cfg := &chaos.Config{
		Type:        chaos.FailureType(req.Type),
		Target:      req.Target,
		Probability: req.Probability,
		LatencyMs:   req.LatencyMs,
		ErrorCode:   req.ErrorCode,
		DurationSec: req.DurationSec,
	}

	h.engine.Inject(cfg)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "injected"})

}

func (h *ChaosHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.engine.Status())
}

func (h *ChaosHandler) handleClear(w http.ResponseWriter, r *http.Request) {
	if (r.Method != http.MethodDelete) {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.engine.Clear()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "cleared"})
}