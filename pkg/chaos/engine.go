package chaos

import (
	"math/rand"
	"sync"
	"time"
)

type FailureType string

const (
	FailureLatency    FailureType = "latency"
	FailureError      FailureType = "error"
	FailureDrop       FailureType = "drop"
	FailureKillSwitch FailureType = "killswitch"
)

type Config struct {
	Type        FailureType
	Target      string // backend URL
	Probability float64
	LatencyMs   int
	ErrorCode   int
	DurationSec int
}

type Engine struct {
	mu 		sync.RWMutex
	active	bool
	config	*Config
}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) Inject(cfg *Config) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.active = true
	e.config = cfg

	if (cfg.DurationSec > 0) {
		go func() {
			time.Sleep(time.Duration(cfg.DurationSec) * time.Second)
			e.Clear()
		}()
	}
}

func (e *Engine) Clear() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.active = false
	e.config = nil
}

//ShouldInject returns the active config if chaos should apply to this
func (e *Engine) ShouldInject(backendURL string) *Config {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if (!e.active || e.config == nil) {
		return nil
	}

	// Empty target equals apply to all backends
	if (e.config.Target != "" && e.config.Target != backendURL) {
		return nil
	}

	// Probability check
	if (rand.Float64() > e.config.Probability) {
		return nil
	}

	return e.config
}

func (e *Engine) Status() map[string]interface{} {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if (!e.active || e.config == nil) {
		return map[string]interface{}{"active": false}
	}

	return map[string]interface{}{
		"active": 	true,
		"type": 	e.config.Type,
		"target":	e.config.Target,
		"probability":	e.config.Probability,
	}
}