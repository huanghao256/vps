package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/vps-inspector/vps-inspector/internal/agent"
	"github.com/vps-inspector/vps-inspector/internal/config"
	"github.com/vps-inspector/vps-inspector/internal/firewall"
	"github.com/vps-inspector/vps-inspector/internal/status"
)

const maxJSONBodyBytes = 1 << 20

type server struct {
	inspector   *agent.Inspector
	store       *agent.RunStore
	statusSvc   status.Service
	firewallSvc firewall.Service
	cfg         config.Config
	logger      *slog.Logger
}

// NewServer builds the HTTP server with API routes, middleware, and static web assets.
func NewServer(cfg config.Config, inspector *agent.Inspector, logger *slog.Logger) *http.Server {
	api := &server{
		inspector:   inspector,
		store:       agent.NewRunStore(50),
		statusSvc:   status.NewService(),
		firewallSvc: firewall.NewService(),
		cfg:         cfg,
		logger:      logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", api.health)
	mux.HandleFunc("GET /api/v1/checks", api.withAuth(api.listChecks))
	mux.HandleFunc("GET /api/v1/runs", api.withAuth(api.listRuns))
	mux.HandleFunc("POST /api/v1/runs", api.withAuth(api.createRun))
	mux.HandleFunc("GET /api/v1/runs/", api.withAuth(api.getRun))
	mux.HandleFunc("GET /api/v1/status", api.withAuth(api.systemStatus))
	mux.HandleFunc("GET /api/v1/firewall", api.withAuth(api.firewallStatus))
	mux.HandleFunc("POST /api/v1/firewall/enable", api.withAuth(api.enableFirewall))
	mux.HandleFunc("POST /api/v1/firewall/disable", api.withAuth(api.disableFirewall))
	mux.HandleFunc("POST /api/v1/firewall/rules", api.withAuth(api.addFirewallRule))
	mux.HandleFunc("DELETE /api/v1/firewall/rules", api.withAuth(api.deleteFirewallRule))
	mux.HandleFunc("/", api.serveWeb)

	handler := api.requestLog(securityHeaders(mux))
	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"auth":   s.cfg.AuthToken != "",
	})
}

func (s *server) listChecks(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"checks": s.inspector.ListChecks()})
}

func (s *server) listRuns(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"runs": s.store.List()})
}

func (s *server) createRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CheckIDs []string `json:"checkIds"`
	}
	if r.Body != nil {
		defer r.Body.Close()
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBodyBytes))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				req.CheckIDs = nil
			} else {
				writeError(w, http.StatusBadRequest, "invalid JSON body")
				return
			}
		}
	}

	run, err := s.inspector.Run(r.Context(), req.CheckIDs)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.store.Add(run)
	writeJSON(w, http.StatusCreated, run)
}

func (s *server) getRun(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/runs/")
	if id == "" {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}
	run, ok := s.store.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (s *server) systemStatus(w http.ResponseWriter, r *http.Request) {
	snapshot, err := s.statusSvc.Snapshot(r.Context())
	if err != nil {
		writeError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func (s *server) firewallStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.firewallSvc.Snapshot(r.Context()))
}

func (s *server) enableFirewall(w http.ResponseWriter, r *http.Request) {
	if err := s.firewallSvc.Enable(r.Context()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.firewallSvc.Snapshot(r.Context()))
}

func (s *server) disableFirewall(w http.ResponseWriter, r *http.Request) {
	if err := s.firewallSvc.Disable(r.Context()); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.firewallSvc.Snapshot(r.Context()))
}

func (s *server) addFirewallRule(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeFirewallRule(w, r)
	if !ok {
		return
	}
	if err := s.firewallSvc.AddRule(r.Context(), req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, s.firewallSvc.Snapshot(r.Context()))
}

func (s *server) deleteFirewallRule(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeFirewallRule(w, r)
	if !ok {
		return
	}
	if err := s.firewallSvc.DeleteRule(r.Context(), req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, s.firewallSvc.Snapshot(r.Context()))
}

func decodeFirewallRule(w http.ResponseWriter, r *http.Request) (firewall.PortRuleRequest, bool) {
	var req firewall.PortRuleRequest
	if r.Body == nil {
		writeError(w, http.StatusBadRequest, "missing JSON body")
		return firewall.PortRuleRequest{}, false
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxJSONBodyBytes))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return firewall.PortRuleRequest{}, false
	}
	return req, true
}

func (s *server) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.cfg.AuthToken == "" {
			next(w, r)
			return
		}
		if !validBearerToken(r.Header.Get("Authorization"), s.cfg.AuthToken) {
			writeError(w, http.StatusUnauthorized, "missing or invalid bearer token")
			return
		}
		next(w, r)
	}
}

func validBearerToken(header, token string) bool {
	expected := "Bearer " + token
	return subtle.ConstantTimeCompare([]byte(header), []byte(expected)) == 1
}

func (s *server) requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		s.logger.Info("request", "method", r.Method, "path", r.URL.Path, "remote", r.RemoteAddr)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{"error": message})
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}
