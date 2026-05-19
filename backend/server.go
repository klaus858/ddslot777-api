package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ServiceName     = "ddslot777-api"
	APIVersion      = "v1"
	ContractVersion = "admin-mvp-v2"
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
	db          *pgxpool.Pool
	dbErr       string
	users       []user
	deposits    []deposit
	withdrawals []withdrawal
	logs        []auditLog
}

func NewHandler() http.Handler {
	s := newDemoStore()
	s.connectDatabase()

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

func (s *store) connectDatabase() {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		s.dbErr = "DATABASE_URL is not configured; using temporary memory store"
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		s.dbErr = "database_pool_error: " + err.Error()
		log.Print(s.dbErr)
		return
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		s.dbErr = "database_ping_error: " + err.Error()
		log.Print(s.dbErr)
		return
	}

	s.db = pool
	if err := s.migrate(ctx); err != nil {
		pool.Close()
		s.db = nil
		s.dbErr = "database_migration_error: " + err.Error()
		log.Print(s.dbErr)
		return
	}

	s.dbErr = ""
}

func (s *store) migrate(ctx context.Context) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			username TEXT NOT NULL,
			balance NUMERIC(14,2) NOT NULL DEFAULT 0,
			vip TEXT NOT NULL,
			status TEXT NOT NULL,
			last_login TEXT NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS deposits (
			order_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id),
			user_name TEXT NOT NULL,
			amount NUMERIC(14,2) NOT NULL,
			channel TEXT NOT NULL,
			state TEXT NOT NULL,
			time_label TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS withdrawals (
			order_id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL REFERENCES users(id),
			user_name TEXT NOT NULL,
			amount NUMERIC(14,2) NOT NULL,
			channel TEXT NOT NULL,
			address TEXT NOT NULL,
			state TEXT NOT NULL,
			time_label TEXT NOT NULL,
			reason TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			actor TEXT NOT NULL,
			action TEXT NOT NULL,
			target TEXT NOT NULL,
			detail TEXT NOT NULL,
			time_label TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS deposits_created_at_idx ON deposits(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS withdrawals_created_at_idx ON withdrawals(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS audit_logs_created_at_idx ON audit_logs(created_at DESC)`,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(ctx, statement); err != nil {
			return err
		}
	}

	for _, item := range s.users {
		if _, err := s.db.Exec(ctx, `INSERT INTO users (id, name, username, balance, vip, status, last_login)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (id) DO NOTHING`, item.ID, item.Name, item.Username, item.Balance, item.VIP, item.Status, item.LastLogin); err != nil {
			return err
		}
	}
	for _, item := range s.deposits {
		if _, err := s.db.Exec(ctx, `INSERT INTO deposits (order_id, user_id, user_name, amount, channel, state, time_label)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (order_id) DO NOTHING`, item.Order, item.UserID, item.User, item.Amount, item.Channel, item.State, item.Time); err != nil {
			return err
		}
	}
	for _, item := range s.withdrawals {
		if _, err := s.db.Exec(ctx, `INSERT INTO withdrawals (order_id, user_id, user_name, amount, channel, address, state, time_label, reason)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			ON CONFLICT (order_id) DO NOTHING`, item.Order, item.UserID, item.User, item.Amount, item.Channel, item.Address, item.State, item.Time, item.Reason); err != nil {
			return err
		}
	}
	for _, item := range s.logs {
		if _, err := s.db.Exec(ctx, `INSERT INTO audit_logs (id, actor, action, target, detail, time_label)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (id) DO NOTHING`, item.ID, item.Actor, item.Action, item.Target, item.Detail, item.Time); err != nil {
			return err
		}
	}

	return nil
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
	payload := response{
		"ok":              true,
		"service":         ServiceName,
		"apiVersion":      APIVersion,
		"contractVersion": ContractVersion,
		"storage":         s.storageMode(),
		"time":            time.Now().UTC().Format(time.RFC3339),
	}
	if s.dbErr != "" {
		payload["storageWarning"] = s.dbErr
	}
	writeJSON(w, http.StatusOK, payload)
}

func (s *store) contract(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, response{
		"apiVersion":      APIVersion,
		"contractVersion": ContractVersion,
		"storage":         s.storageMode(),
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
	if s.db != nil {
		s.dbSummary(w, r)
		return
	}

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

	writeJSON(w, http.StatusOK, response{
		"todayDeposit":       todayDeposit,
		"pendingDeposits":    pendingDeposits,
		"pendingWithdrawals": countPendingWithdrawals(s.withdrawals),
		"pendingWithdraws":   countPendingWithdrawals(s.withdrawals),
		"newMembers":         len(s.users),
		"totalUsers":         len(s.users),
		"activityClaims":     384,
	})
}

func (s *store) usersList(w http.ResponseWriter, r *http.Request) {
	if s.db != nil {
		s.dbUsersList(w, r)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, http.StatusOK, response{"items": append([]user(nil), s.users...)})
}

func (s *store) updateBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, response{"error": "method_not_allowed"})
		return
	}
	if s.db != nil {
		s.dbUpdateBalance(w, r)
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
	if s.db != nil {
		s.dbDepositsList(w, r)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, http.StatusOK, response{"items": append([]deposit(nil), s.deposits...)})
}

func (s *store) confirmDeposit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, response{"error": "method_not_allowed"})
		return
	}
	if s.db != nil {
		s.dbConfirmDeposit(w, r)
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
	if s.db != nil {
		s.dbWithdrawalsList(w, r)
		return
	}

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
	if s.db != nil {
		s.dbSetWithdrawalState(w, r, state)
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
	if s.db != nil {
		s.dbAuditLogs(w, r)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	writeJSON(w, http.StatusOK, response{"items": append([]auditLog(nil), s.logs...)})
}

func (s *store) walletSession(w http.ResponseWriter, r *http.Request) {
	if s.db != nil {
		s.dbWalletSession(w, r)
		return
	}

	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = "U10021"
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.users {
		if item.ID == userID {
			writeJSON(w, http.StatusOK, response{"ok": true, "user": item})
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
	if s.db != nil {
		s.dbWalletDepositSuccess(w, r)
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

	for i := range s.users {
		if s.users[i].ID == payload.UserID {
			s.users[i].Balance += payload.Amount
			if payload.Channel == "" {
				payload.Channel = "demo-wallet"
			}
			if payload.Order == "" {
				payload.Order = "D" + time.Now().UTC().Format("20060102150405")
			}
			entry := deposit{Order: payload.Order, UserID: payload.UserID, User: s.users[i].Name, Amount: payload.Amount, Channel: payload.Channel, State: "credited", Time: time.Now().Format("15:04")}
			s.deposits = append([]deposit{entry}, s.deposits...)
			s.addLogLocked("wallet", "deposit_success", payload.Order, s.users[i].Name+" "+money(payload.Amount)+" via "+payload.Channel)
			writeJSON(w, http.StatusOK, response{"ok": true, "deposit": entry, "user": s.users[i], "balance": s.users[i].Balance})
			return
		}
	}

	writeJSON(w, http.StatusNotFound, response{"error": "user_not_found"})
}

func (s *store) dbSummary(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	var todayDeposit float64
	var pendingDeposits int
	var pendingWithdrawals int
	var totalUsers int
	if err := s.db.QueryRow(ctx, `SELECT COALESCE(SUM(amount) FILTER (WHERE state = 'credited'), 0)::float8, COUNT(*) FILTER (WHERE state = 'pending') FROM deposits`).Scan(&todayDeposit, &pendingDeposits); err != nil {
		writeServerError(w, err)
		return
	}
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM withdrawals WHERE state = 'pending'`).Scan(&pendingWithdrawals); err != nil {
		writeServerError(w, err)
		return
	}
	if err := s.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&totalUsers); err != nil {
		writeServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{
		"todayDeposit":       todayDeposit,
		"pendingDeposits":    pendingDeposits,
		"pendingWithdrawals": pendingWithdrawals,
		"pendingWithdraws":   pendingWithdrawals,
		"newMembers":         totalUsers,
		"totalUsers":         totalUsers,
		"activityClaims":     384,
	})
}

func (s *store) dbUsersList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	rows, err := s.db.Query(ctx, `SELECT id, name, username, balance::float8, vip, status, last_login FROM users ORDER BY id`)
	if err != nil {
		writeServerError(w, err)
		return
	}
	defer rows.Close()

	items := []user{}
	for rows.Next() {
		var item user
		if err := rows.Scan(&item.ID, &item.Name, &item.Username, &item.Balance, &item.VIP, &item.Status, &item.LastLogin); err != nil {
			writeServerError(w, err)
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, response{"items": items})
}

func (s *store) dbUpdateBalance(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UserID  string  `json:"userId"`
		Balance float64 `json:"balance"`
		Reason  string  `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid_json"})
		return
	}

	ctx, cancel := requestContext(r)
	defer cancel()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		writeServerError(w, err)
		return
	}
	defer tx.Rollback(ctx)

	var oldBalance float64
	if err := tx.QueryRow(ctx, `SELECT balance::float8 FROM users WHERE id = $1 FOR UPDATE`, payload.UserID).Scan(&oldBalance); err != nil {
		writeNotFoundOrServerError(w, err, "user_not_found")
		return
	}

	var item user
	err = tx.QueryRow(ctx, `UPDATE users SET balance = $2, updated_at = NOW() WHERE id = $1 RETURNING id, name, username, balance::float8, vip, status, last_login`, payload.UserID, payload.Balance).Scan(&item.ID, &item.Name, &item.Username, &item.Balance, &item.VIP, &item.Status, &item.LastLogin)
	if err != nil {
		writeServerError(w, err)
		return
	}
	if err := insertAuditTx(ctx, tx, "admin", "update_balance", payload.UserID, money(oldBalance)+" -> "+money(payload.Balance)+" "+payload.Reason); err != nil {
		writeServerError(w, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{"ok": true, "user": item})
}

func (s *store) dbDepositsList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	rows, err := s.db.Query(ctx, `SELECT order_id, user_id, user_name, amount::float8, channel, state, time_label FROM deposits ORDER BY created_at DESC, order_id DESC`)
	if err != nil {
		writeServerError(w, err)
		return
	}
	defer rows.Close()

	items := []deposit{}
	for rows.Next() {
		var item deposit
		if err := rows.Scan(&item.Order, &item.UserID, &item.User, &item.Amount, &item.Channel, &item.State, &item.Time); err != nil {
			writeServerError(w, err)
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, response{"items": items})
}

func (s *store) dbConfirmDeposit(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Order string `json:"order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid_json"})
		return
	}

	ctx, cancel := requestContext(r)
	defer cancel()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		writeServerError(w, err)
		return
	}
	defer tx.Rollback(ctx)

	item, err := selectDepositTx(ctx, tx, payload.Order, true)
	if err != nil {
		writeNotFoundOrServerError(w, err, "deposit_not_found")
		return
	}
	if item.State != "credited" {
		if _, err := tx.Exec(ctx, `UPDATE users SET balance = balance + $2, updated_at = NOW() WHERE id = $1`, item.UserID, item.Amount); err != nil {
			writeServerError(w, err)
			return
		}
		if _, err := tx.Exec(ctx, `UPDATE deposits SET state = 'credited', updated_at = NOW() WHERE order_id = $1`, item.Order); err != nil {
			writeServerError(w, err)
			return
		}
		item.State = "credited"
	}
	if err := insertAuditTx(ctx, tx, "admin", "confirm_deposit", item.Order, "Credited "+money(item.Amount)+" to "+item.User); err != nil {
		writeServerError(w, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{"ok": true, "deposit": item})
}

func (s *store) dbWithdrawalsList(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	rows, err := s.db.Query(ctx, `SELECT order_id, user_id, user_name, amount::float8, channel, address, state, time_label, reason FROM withdrawals ORDER BY created_at DESC, order_id DESC`)
	if err != nil {
		writeServerError(w, err)
		return
	}
	defer rows.Close()

	items := []withdrawal{}
	for rows.Next() {
		var item withdrawal
		if err := rows.Scan(&item.Order, &item.UserID, &item.User, &item.Amount, &item.Channel, &item.Address, &item.State, &item.Time, &item.Reason); err != nil {
			writeServerError(w, err)
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, response{"items": items})
}

func (s *store) dbSetWithdrawalState(w http.ResponseWriter, r *http.Request, state string) {
	var payload struct {
		Order  string `json:"order"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, response{"error": "invalid_json"})
		return
	}

	ctx, cancel := requestContext(r)
	defer cancel()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		writeServerError(w, err)
		return
	}
	defer tx.Rollback(ctx)

	item, err := selectWithdrawalTx(ctx, tx, payload.Order, true)
	if err != nil {
		writeNotFoundOrServerError(w, err, "withdrawal_not_found")
		return
	}
	if item.State != "pending" {
		writeJSON(w, http.StatusConflict, response{"error": "withdrawal_already_processed"})
		return
	}

	err = tx.QueryRow(ctx, `UPDATE withdrawals SET state = $2, reason = $3, updated_at = NOW() WHERE order_id = $1 RETURNING order_id, user_id, user_name, amount::float8, channel, address, state, time_label, reason`, payload.Order, state, payload.Reason).Scan(&item.Order, &item.UserID, &item.User, &item.Amount, &item.Channel, &item.Address, &item.State, &item.Time, &item.Reason)
	if err != nil {
		writeServerError(w, err)
		return
	}
	if err := insertAuditTx(ctx, tx, "admin", state+"_withdrawal", item.Order, item.User+" "+money(item.Amount)+" "+payload.Reason); err != nil {
		writeServerError(w, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{"ok": true, "withdrawal": item})
}

func (s *store) dbAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := requestContext(r)
	defer cancel()

	rows, err := s.db.Query(ctx, `SELECT id, actor, action, target, detail, time_label FROM audit_logs ORDER BY created_at DESC, id DESC LIMIT 80`)
	if err != nil {
		writeServerError(w, err)
		return
	}
	defer rows.Close()

	items := []auditLog{}
	for rows.Next() {
		var item auditLog
		if err := rows.Scan(&item.ID, &item.Actor, &item.Action, &item.Target, &item.Detail, &item.Time); err != nil {
			writeServerError(w, err)
			return
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, response{"items": items})
}

func (s *store) dbWalletSession(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = "U10021"
	}

	ctx, cancel := requestContext(r)
	defer cancel()

	item, err := selectUser(ctx, s.db, userID)
	if err != nil {
		writeNotFoundOrServerError(w, err, "user_not_found")
		return
	}
	writeJSON(w, http.StatusOK, response{"ok": true, "user": item})
}

func (s *store) dbWalletDepositSuccess(w http.ResponseWriter, r *http.Request) {
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
	if payload.Channel == "" {
		payload.Channel = "demo-wallet"
	}
	if payload.Order == "" {
		payload.Order = "D" + time.Now().UTC().Format("20060102150405")
	}

	ctx, cancel := requestContext(r)
	defer cancel()

	tx, err := s.db.Begin(ctx)
	if err != nil {
		writeServerError(w, err)
		return
	}
	defer tx.Rollback(ctx)

	item, err := selectUserTx(ctx, tx, payload.UserID, true)
	if err != nil {
		writeNotFoundOrServerError(w, err, "user_not_found")
		return
	}

	err = tx.QueryRow(ctx, `UPDATE users SET balance = balance + $2, updated_at = NOW() WHERE id = $1 RETURNING id, name, username, balance::float8, vip, status, last_login`, payload.UserID, payload.Amount).Scan(&item.ID, &item.Name, &item.Username, &item.Balance, &item.VIP, &item.Status, &item.LastLogin)
	if err != nil {
		writeServerError(w, err)
		return
	}

	entry := deposit{
		Order:   payload.Order,
		UserID:  payload.UserID,
		User:    item.Name,
		Amount:  payload.Amount,
		Channel: payload.Channel,
		State:   "credited",
		Time:    time.Now().Format("15:04"),
	}
	if _, err := tx.Exec(ctx, `INSERT INTO deposits (order_id, user_id, user_name, amount, channel, state, time_label)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (order_id) DO UPDATE SET state = EXCLUDED.state, amount = EXCLUDED.amount, updated_at = NOW()`, entry.Order, entry.UserID, entry.User, entry.Amount, entry.Channel, entry.State, entry.Time); err != nil {
		writeServerError(w, err)
		return
	}
	if err := insertAuditTx(ctx, tx, "wallet", "deposit_success", payload.Order, item.Name+" "+money(payload.Amount)+" via "+payload.Channel); err != nil {
		writeServerError(w, err)
		return
	}
	if err := tx.Commit(ctx); err != nil {
		writeServerError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, response{"ok": true, "deposit": entry, "user": item, "balance": item.Balance})
}

func selectUser(ctx context.Context, db *pgxpool.Pool, userID string) (user, error) {
	var item user
	err := db.QueryRow(ctx, `SELECT id, name, username, balance::float8, vip, status, last_login FROM users WHERE id = $1`, userID).Scan(&item.ID, &item.Name, &item.Username, &item.Balance, &item.VIP, &item.Status, &item.LastLogin)
	return item, err
}

func selectUserTx(ctx context.Context, tx pgx.Tx, userID string, forUpdate bool) (user, error) {
	var item user
	query := `SELECT id, name, username, balance::float8, vip, status, last_login FROM users WHERE id = $1`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	err := tx.QueryRow(ctx, query, userID).Scan(&item.ID, &item.Name, &item.Username, &item.Balance, &item.VIP, &item.Status, &item.LastLogin)
	return item, err
}

func selectDepositTx(ctx context.Context, tx pgx.Tx, order string, forUpdate bool) (deposit, error) {
	var item deposit
	query := `SELECT order_id, user_id, user_name, amount::float8, channel, state, time_label FROM deposits WHERE order_id = $1`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	err := tx.QueryRow(ctx, query, order).Scan(&item.Order, &item.UserID, &item.User, &item.Amount, &item.Channel, &item.State, &item.Time)
	return item, err
}

func selectWithdrawalTx(ctx context.Context, tx pgx.Tx, order string, forUpdate bool) (withdrawal, error) {
	var item withdrawal
	query := `SELECT order_id, user_id, user_name, amount::float8, channel, address, state, time_label, reason FROM withdrawals WHERE order_id = $1`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	err := tx.QueryRow(ctx, query, order).Scan(&item.Order, &item.UserID, &item.User, &item.Amount, &item.Channel, &item.Address, &item.State, &item.Time, &item.Reason)
	return item, err
}

func insertAuditTx(ctx context.Context, tx pgx.Tx, actor string, action string, target string, detail string) error {
	_, err := tx.Exec(ctx, `INSERT INTO audit_logs (id, actor, action, target, detail, time_label)
		VALUES ($1, $2, $3, $4, $5, $6)`, logID(), actor, action, target, detail, time.Now().Format("2006-01-02 15:04:05"))
	return err
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
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = s.db.Exec(ctx, `INSERT INTO audit_logs (id, actor, action, target, detail, time_label)
			VALUES ($1, $2, $3, $4, $5, $6)`, logID(), actor, action, target, detail, time.Now().Format("2006-01-02 15:04:05"))
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.addLogLocked(actor, action, target, detail)
}

func (s *store) addLogLocked(actor string, action string, target string, detail string) {
	entry := auditLog{
		ID:     logID(),
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

func countPendingWithdrawals(items []withdrawal) int {
	count := 0
	for _, item := range items {
		if item.State == "pending" {
			count++
		}
	}
	return count
}

func requestContext(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), 8*time.Second)
}

func logID() string {
	return fmt.Sprintf("L%d", time.Now().UTC().UnixNano())
}

func money(value float64) string {
	return fmt.Sprintf("$%.2f", value)
}

func (s *store) storageMode() string {
	if s.db != nil {
		return "postgres"
	}
	return "memory"
}

func writeNotFoundOrServerError(w http.ResponseWriter, err error, notFound string) {
	if err == pgx.ErrNoRows {
		writeJSON(w, http.StatusNotFound, response{"error": notFound})
		return
	}
	writeServerError(w, err)
}

func writeServerError(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusInternalServerError, response{"error": "server_error", "detail": err.Error()})
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
