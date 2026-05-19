package backend

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	ServiceName     = "ddslot777-api"
	APIVersion      = "v1"
	ContractVersion = "admin-mvp-v1"
	adminToken       = "demo-admin-token"
)

type response map[string]any

type user struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Username  string  `json:"username"`
	Balance   float64 `json:"balance"`
	VIP       string  `json:"vip"`
	Status    string  `json:"status"`
	LastLogin string  `json:"lastLogin"`
}

type deposit struct {
	Order   string  `json:"order"`
	UserID  string  `json:"userId"`
	User    string  `json:"user"`
	Amount  float64 `json:"amount"`
	Channel string  `json:"channel"`
	State   string  `json:"state"`
	Time    string  `json:"time"`
}

type withdrawal struct {
	Order   string  `json:"order"`
	UserID  string  `json:"userId"`
	User    string  `json:"user"`
	Amount  float64 `json:"amount"`
	Channel string  `json:"channel"`
	Address string  `json:"address"`
	State   string  `json:"state"`
	Time    string  `json:"time"`
	Reason  string  `json:"reason,omitempty"`
}

type auditLog struct {
	ID     string `json:"id"`
	Actor  string `json:"actor"`
	Action string `json:"action"`
	Target string `json:"target"`
	Detail string `json:"detail"`
	Time   string `json:"time"`
}

type store struct {
	mu          sync.Mutex
	users       []user
	deposits    []deposit
	withdrawals []withdrawal
	logs        []auditLog
}

func NewHandler() http.Handler {
	s := newDemoStore()
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.route)
	return mux
}

func newDemoStore() *store {
	return &store{
		users: []user{
			{ID: "U10021", Name: "Player123", Username: "Player123", Balance: 860.50, VIP: "VIP 2", Status: "normal", LastLogin: "2026-05-19 12:40"},
			{ID: "U10022", Name: "LuckyJoe", Username: "LuckyJoe", Balance: 2410.00, VIP: "VIP 4", Status: "normal", LastLogin: "2026-05-19 11:18"},
			{ID: "U10023", Name: "HighRoller", Username: "HighRoller", Balance: 9800.00, VIP: "VIP 6", Status: "risk_review", LastLogin: "2026-05-19 10:06"},
			{ID: "U10024", Name: "PlayWin88", Username: "PlayWin88", Balance: 320.75, VIP: "VIP 1", Status: "normal", LastLogin: "2026-05-19 09:54"},
		},
		deposits: []deposit{
			{Order: "D20260519001", UserID: "U10021", User: "Player123", Amount: 100.00, Channel: "USDT", State: "credited", Time: "12:34"},
			{Order: "D20260519002", UserID: "U10022", User: "LuckyJoe", Amount: 300.00, Channel: "CashApp", State: "pending", Time: "12:22"},
			{Order: "D20260519003", UserID: "U10024", User: "PlayWin88", Amount: 50.00, Channel: "PayPal", State: "credited", Time: "11:48"},
			{Order: "D20260519004", UserID: "U10023", User: "HighRoller", Amount: 1000.00, Channel: "Bank", State: "risk_review", Time: "11:05"},
		},
		withdrawals: []withdrawal{
			{Order: "W20260519001", UserID: "U10022", User: "LuckyJoe", Amount: 120.00, Channel: "USDT", Address: "TQx9...8pA2", State: "pending", Time: "12:52"},
			{Order: "W20260519002", UserID: "U10023", User: "HighRoller", Amount: 800.00, Channel: "Bank", Address: "**** 2841", State: "pending", Time: "12:08"},
			{Order: "W20260519003", UserID: "U10021", User: "Player123", Amount: 45.00, Channel: "CashApp", Address: "$player123", State: "approved", Time: "10:31"},
		},
		logs: []auditLog{
			{ID: "L10003", Actor: "system", Action: "api_started", Target: "ddslot777-api", Detail: "Demo API initialized", Time: "2026-05-19 12:00:00"},
			{ID: "L10002", Actor: "admin", Action: "confirm_deposit", Target: "D20260519001", Detail: "Credited $100.00 to Player123", Time: "2026-05-19 12:35:00"},
			{ID: "L10001", Actor: "admin", Action: "login", Target: "admin", Detail: "Admin login succeeded", Time: "2026-05-19 12:30:00"},
		},
	}
}

func (s *store) route(w http.ResponseWriter, r *http.Request) {
	withCORS(w, r)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	if path == "" {
		path = "/"
	}

	switch path {
	case "/", "/api", "/health", "/api/health":
		s.health(w)
	case "/contract", "/api/contract":
		s.contract(w)
	case "/auth/login", "/api/auth/login":
		s.login(w, r)
	case "/admin/summary", "/api/admin/summary":
		s.requireToken(w, r, s.summary)
	case "/admin/members", "/api/admin/members", "/admin/users", "/api/admin/users":
		s.requireToken(w, r, s.usersList)
	case "/admin/users/balance", "/api/admin/users/balance":
		s.requireToken(w, r, s.updateBalance)
	case "/admin/deposits", "/api/admin/deposits":
		s.requireToken(w, r, s.depositsList)
	case "/admin/deposits/confirm", "/api/admin/deposits/confirm":
		s.requireToken(w, r, s.confirmDeposit)
	case "/admin/withdrawals", "/api/admin/withdrawals":
		s.requireToken(w, r, s.withdrawalsList)
	case "/admin/withdrawals/approve", "/api/admin/withdrawals/approve":
		s.requireToken(w, r, s.approveWithdrawal)
	case "/admin/withdrawals/reject", "/api/admin/withdrawals/reject":
		s.requireToken(w, r, s.rejectWithdrawal)
	case "/admin/audit-logs", "/api/admin/audit-logs":
		s.requireToken(w, r, s.auditLogs)
	case "/wallet/session", "/api/wallet/session":
		s.walletSession(w, r)
	case "/wallet/deposit/success", "/api/wallet/deposit/success":
		s.walletDepositSuccess(w, r)
	default:
		writeJSON(w, http.StatusNotFound, response{"error": "not_found", "path": r.URL.Path})
	}
}

func (s *store) health(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, response{
		"ok":              true,
		"service":         ServiceName,
		"apiVersion":      APIVersion,
		"contractVersion": ContractVersion,
		"time":            time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *store) contract(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, response{
		"apiVersion":      APIVersion,
		"contractVersion": ContractVersion,
		"statusValues": response{
			"user":       []string{"normal", "risk_review", "disabled"},
			"deposit":    []string{"pending", "credited", "risk_review"},
			"withdrawal": []string{"pending", "approved", "rejected"},
		},
		"endpoints": []response{
			{"method": "GET", "path": "/api/wallet/session", "auth": false},
			{"method": "POST", "path": "/api/wallet/deposit/success", "auth": false},
			{"method": "POST", "path": "/api/auth/login", "auth": false},
			{"method": "GET", "path": "/api/admin/summary", "auth": true},
			{"method": "GET", "path": "/api/admin/users", "auth": true},
			{"method": "POST", "path": "/api/admin/users/balance", "auth": true},
			{"method": "GET", "path": "/api/admin/deposits", "auth": true},
			{"method": "POST", "path": "/api/admin/deposits/confirm", "auth": true},
			{"method": "GET", "path": "/api/admin/withdrawals", "auth": true},
			{"method": "POST", "path": "/api/admin/withdrawals/approve", "auth": true},
			{"method": "POST", "path": "/api/admin/withdrawals/reject", "auth": true},
			{"method": "GET", "path": "/api/admin/audit-logs", "auth": true},
		},
	})
}

func (s *store) login(w http.ResponseWriter, r *http.Request) {
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

	s.addLog("admin", "login", "admin", "Admin login succeeded")
	writeJSON(w, http.StatusOK, response{"token": adminToken, "user": response{"name": "admin", "role": "owner"}})
}

func (s *store) summary(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var todayDeposit float64
	pendingDeposits := 0
	for _, item := range s.deposits {
		if item.State == "credited" {
			todayDeposit += item.Amount
		}
		if item.State == "pending" {
			pendingDeposits++
		}
	}

	pendingWithdrawals := 0
	for _, item := range s.withdrawals {
		if item.State == "pending" {
			pendingWithdrawals++
		}
	}

	writeJSON(w, http.StatusOK, response{
		"todayDeposit":       todayDeposit,
		"pendingDeposits":    pendingDeposits,
		"pendingWithdrawals": pendingWithdrawals,
		"pendingWithdraws":   pendingWithdrawals,
		"newMembers":         len(s.users),
		"totalUsers":         len(s.users),
		"activityClaims":     384,
	})
}

func (s *store) usersList(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, http.StatusOK, response{"items": append([]user(nil), s.users...)})
}

func (s *store) updateBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, response{"error": "method_not_allowed"})
		return
	}

	var payload struct {
		UserID  string  `json:"userId"`
		Balance float64 `json:"balance"`
		Reason  string  `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid_json"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.users {
		if s.users[i].ID == payload.UserID {
			oldBalance := s.users[i].Balance
			s.users[i].Balance = payload.Balance
			s.addLogLocked("admin", "update_balance", payload.UserID, money(oldBalance)+" -> "+money(payload.Balance)+" "+payload.Reason)
			writeJSON(w, http.StatusOK, response{"ok": true, "user": s.users[i]})
			return
		}
	}
	writeJSON(w, http.StatusNotFound, response{"error": "user_not_found"})
}

func (s *store) depositsList(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, http.StatusOK, response{"items": append([]deposit(nil), s.deposits...)})
}

func (s *store) confirmDeposit(w http.ResponseWriter, r *http.Request) {
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

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.deposits {
		if s.deposits[i].Order == payload.Order {
			if s.deposits[i].State != "credited" {
				for userIndex := range s.users {
					if s.users[userIndex].ID == s.deposits[i].UserID {
						s.users[userIndex].Balance += s.deposits[i].Amount
						break
					}
				}
			}
			s.deposits[i].State = "credited"
			s.addLogLocked("admin", "confirm_deposit", payload.Order, "Credited "+money(s.deposits[i].Amount)+" to "+s.deposits[i].User)
			writeJSON(w, http.StatusOK, response{"ok": true, "deposit": s.deposits[i]})
			return
		}
	}
	writeJSON(w, http.StatusNotFound, response{"error": "deposit_not_found"})
}

func (s *store) withdrawalsList(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, http.StatusOK, response{"items": append([]withdrawal(nil), s.withdrawals...)})
}

func (s *store) approveWithdrawal(w http.ResponseWriter, r *http.Request) {
	s.setWithdrawalState(w, r, "approved")
}

func (s *store) rejectWithdrawal(w http.ResponseWriter, r *http.Request) {
	s.setWithdrawalState(w, r, "rejected")
}

func (s *store) setWithdrawalState(w http.ResponseWriter, r *http.Request, state string) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, response{"error": "method_not_allowed"})
		return
	}

	var payload struct {
		Order  string `json:"order"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid_json"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.withdrawals {
		if s.withdrawals[i].Order == payload.Order {
			if s.withdrawals[i].State != "pending" {
				writeJSON(w, http.StatusConflict, response{"error": "withdrawal_already_processed"})
				return
			}
			s.withdrawals[i].State = state
			s.withdrawals[i].Reason = payload.Reason
			s.addLogLocked("admin", state+"_withdrawal", payload.Order, s.withdrawals[i].User+" "+money(s.withdrawals[i].Amount)+" "+payload.Reason)
			writeJSON(w, http.StatusOK, response{"ok": true, "withdrawal": s.withdrawals[i]})
			return
		}
	}
	writeJSON(w, http.StatusNotFound, response{"error": "withdrawal_not_found"})
}

func (s *store) auditLogs(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, http.StatusOK, response{"items": append([]auditLog(nil), s.logs...)})
}

func (s *store) walletSession(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = "U10021"
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.users {
		if item.ID == userID {
			writeJSON(w, http.StatusOK, response{
				"ok":   true,
				"user": item,
			})
			return
		}
	}

	writeJSON(w, http.StatusNotFound, response{"error": "user_not_found"})
}

func (s *store) walletDepositSuccess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, response{"error": "method_not_allowed"})
		return
	}

	var payload struct {
		UserID  string  `json:"userId"`
		Amount  float64 `json:"amount"`
		Channel string  `json:"channel"`
		Order   string  `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid_json"})
		return
	}
	if payload.UserID == "" {
		payload.UserID = "U10021"
	}
	if payload.Amount <= 0 {
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid_amount"})
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var creditedUser *user
	for i := range s.users {
		if s.users[i].ID == payload.UserID {
			s.users[i].Balance += payload.Amount
			creditedUser = &s.users[i]
			break
		}
	}
	if creditedUser == nil {
		writeJSON(w, http.StatusNotFound, response{"error": "user_not_found"})
		return
	}

	if payload.Channel == "" {
		payload.Channel = "demo-wallet"
	}
	if payload.Order == "" {
		payload.Order = "D" + time.Now().UTC().Format("20060102150405")
	}

	entry := deposit{
		Order:   payload.Order,
		UserID:  payload.UserID,
		User:    creditedUser.Name,
		Amount:  payload.Amount,
		Channel: payload.Channel,
		State:   "credited",
		Time:    time.Now().Format("15:04"),
	}
	s.deposits = append([]deposit{entry}, s.deposits...)
	s.addLogLocked("wallet", "deposit_success", payload.Order, creditedUser.Name+" "+money(payload.Amount)+" via "+payload.Channel)

	writeJSON(w, http.StatusOK, response{
		"ok":      true,
		"deposit": entry,
		"user":    *creditedUser,
		"balance": creditedUser.Balance,
	})
}

func (s *store) requireToken(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if token != adminToken {
		writeJSON(w, http.StatusUnauthorized, response{"error": "unauthorized"})
		return
	}
	next(w, r)
}

func (s *store) addLog(actor string, action string, target string, detail string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.addLogLocked(actor, action, target, detail)
}

func (s *store) addLogLocked(actor string, action string, target string, detail string) {
	entry := auditLog{
		ID:     "L" + time.Now().UTC().Format("20060102150405"),
		Actor:  actor,
		Action: action,
		Target: target,
		Detail: detail,
		Time:   time.Now().Format("2006-01-02 15:04:05"),
	}
	s.logs = append([]auditLog{entry}, s.logs...)
	if len(s.logs) > 50 {
		s.logs = s.logs[:50]
	}
}

func money(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}

func withCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = "*"
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
