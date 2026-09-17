package health

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Check func(ctx context.Context) error

type Registry struct {
	mu     sync.RWMutex
	checks map[string]Check
}

func NewRegistry() *Registry {
	return &Registry{checks: make(map[string]Check)}
}

func (r *Registry) Add(name string, check Check) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks[name] = check
}

func Live(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (r *Registry) Ready(w http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	r.mu.RLock()
	defer r.mu.RUnlock()

	failures := map[string]string{}
	for name, check := range r.checks {
		if err := check(ctx); err != nil {
			failures[name] = err.Error()
		}
	}

	if len(failures) > 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status":   "not_ready",
			"failures": failures,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
