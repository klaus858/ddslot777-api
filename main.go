package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	serviceName     = "ddslot777-api"
	apiVersion      = "v1"
	contractVersion = "admin-v1"
)

type response map[string]any

type member struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Balance   string `json:"balance"`
	VIP       string `json:"vip"`
	Status    string `json:"status"`
	LastLogin string `json:"lastLogin"`
}

type deposit struct {
	Order   string `json:"order"`
	User    string `json:"user"`
	Amount  string `json:"amount"`
	Channel string `json:"channel"`
	State   string `json:"state"`
	Time    string `json:"time"`
}

var members = []member{
	{ID: "U10021", Name: "Player123", Balance: "$ 860.50", VIP: "VIP 2", Status: "normal", LastLogin: "2026-05-19 12:40"},
	{ID: "U10022", Name: "LuckyJoe", Balance: "$ 2,410.00", VIP: "VIP 4", Status: "normal", LastLogin: "2026-05-19 11:18"},
	{ID: "U10023", Name: "HighRoller", Balance: "$ 9,800.00", VIP: "VIP 6", Status: "risk_review", LastLogin: "2026-05-19 10:06"},
	{ID: "U10024", Name: "PlayWin88", Balance: "$ 320.75", VIP: "VIP 1", Status: "normal", LastLogin: "2026-05-19 09:54"},
}

var deposits = []deposit{
	{Order: "D20260519001", User: "Player123", Amount: "$ 100.00", Channel: "USDT", State: "credited", Time: "12:34"},
	{Order: "D20260519002", User: "LuckyJoe", Amount: "$ 300.00", Channel: "CashApp", State: "pending", Time: "12:22"},
	{Order: "D20260519003", User: "PlayWin88", Amount: "$ 50.00", Channel: "PayPal", State: "credited", Time: "11:48"},
	{Order: "D20260519004", User: "HighRoller", Amount: "$ 1,000.00", Channel: "Bank", State: "risk_review", Time: "11:05"},
}

func main() {
	mux := http.NewServeMux()
	registerRoutes(mux, "")
	registerRoutes(mux, "/api")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("%s listening on :%s", serviceName, port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func registerRoutes(mux *http.ServeMux, prefix string) {
	mux.HandleFunc(prefix+"/health", withCORS(healthHandler))
	mux.HandleFunc(prefix+"/contract", withCORS(contractHandler))
	mux.HandleFunc(prefix+"/auth/login", withCORS(loginHandler))
	mux.HandleFunc(prefix+"/admin/summary", withCORS(requireToken(summaryHandler)))
	mux.HandleFunc(prefix+"/admin/members", withCORS(requireToken(membersHandler)))
	mux.HandleFunc(prefix+"/admin/deposits", withCORS(requireToken(depositsHandler)))
	mux.HandleFunc(prefix+"/admin/deposits/confirm", withCORS(requireToken(confirmDepositHandler)))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{
		"ok":              true,
		"service":         serviceName,
		"apiVersion":      apiVersion,
		"contractVersion": contractVersion,
		"time":            time.Now().UTC().Format(time.RFC3339),
	})
}

func contractHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{
		"apiVersion":      apiVersion,
		"contractVersion": contractVersion,
		"auth": response{
			"type":   "Bearer token",
			"header": "Authorization: Bearer <token>",
		},
		"statusValues": response{
			"member":  []string{"normal", "risk_review"},
			"deposit": []string{"pending", "credited", "risk_review"},
		},
		"endpoints": []response{
			{"method": "GET", "path": "/api/health", "auth": false},
			{"method": "GET", "path": "/api/contract", "auth": false},
			{"method": "POST", "path": "/api/auth/login", "auth": false},
			{"method": "GET", "path": "/api/admin/summary", "auth": true},
			{"method": "GET", "path": "/api/admin/members", "auth": true},
			{"method": "GET", "path": "/api/admin/deposits", "auth": true},
			{"method": "POST", "path": "/api/admin/deposits/confirm", "auth": true},
		},
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, response{"error": "method_not_allowed"})
		return
	}

	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid_json"})
		return
	}

	if payload.Username != "admin" || payload.Password != "admin123" {
		writeJSON(w, http.StatusUnauthorized, response{"error": "invalid_credentials"})
		return
	}

	writeJSON(w, http.StatusOK, response{
		"token": "demo-admin-token",
		"user": response{
			"name": "admin",
			"role": "owner",
		},
	})
}

func summaryHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{
		"todayDeposit":     12480,
		"pendingWithdraws": 7,
		"newMembers":       126,
		"activityClaims":   384,
	})
}

func membersHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{"items": members})
}

func depositsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, response{"items": deposits})
}

func confirmDepositHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, response{"error": "method_not_allowed"})
		return
	}

	var payload struct {
		Order string `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid_json"})
		return
	}

	for i := range deposits {
		if deposits[i].Order == payload.Order {
			deposits[i].State = "credited"
			writeJSON(w, http.StatusOK, response{"ok": true, "deposit": deposits[i]})
			return
		}
	}

	writeJSON(w, http.StatusNotFound, response{"error": "deposit_not_found"})
}

func requireToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token != "demo-admin-token" {
			writeJSON(w, http.StatusUnauthorized, response{"error": "unauthorized"})
			return
		}
		next(w, r)
	}
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
