# ddslot777-api

Go API service for the DD Slot admin system.

Live API:

- https://ddslot777-api.vercel.app

Contract:

- API version: `v1`
- Admin contract version: `admin-mvp-v3`
- Contract endpoint: `GET /api/contract`
- Storage: Postgres when `DATABASE_URL` is configured, temporary memory fallback otherwise.

Current endpoints:

- `GET /api/health`
- `GET /api/contract`
- `GET /api/wallet/session`
- `POST /api/wallet/deposit/success`
- `POST /api/auth/login`
- `POST /api/auth/register`
- `GET /api/auth/me`
- `GET /api/admin/summary`
- `GET /api/admin/users`
- `GET /api/admin/members`
- `POST /api/admin/users/balance`
- `GET /api/admin/deposits`
- `POST /api/admin/deposits/confirm`
- `GET /api/admin/withdrawals`
- `POST /api/admin/withdrawals/approve`
- `POST /api/admin/withdrawals/reject`
- `GET /api/admin/audit-logs`

Demo login payload:

```json
{
  "username": "admin",
  "password": "admin123"
}
```

Successful login returns a Bearer token. Protected endpoints require:

```text
Authorization: Bearer demo-admin-token
```

Demo player accounts:

- `player123@ddslot777.com` / `Demo123`
- `5550001001` / `Demo123`

Player login payload:

```json
{
  "identifier": "player123@ddslot777.com",
  "password": "Demo123"
}
```

Player register payload:

```json
{
  "phone": "5551234567",
  "password": "Aa1234"
}
```

Successful player login/register returns a Bearer token such as `demo-player-U10021`.
The frontend should store that token and call `GET /api/auth/me` to sync the current user and balance.
Wallet session and deposit callback endpoints require the same player Bearer token.

Phase 1 admin loop:

- Admin login
- User list
- Player register/login and frontend auth state
- Frontend wallet deposit callback synced to user balance and deposit orders
- Deposit order list and confirm deposit
- Withdrawal order list and approve/reject withdrawal
- Manual user balance update
- Simple operation/audit log

Database setup:

- Create a Postgres database, for example Vercel Postgres, Neon, or Supabase.
- Add the connection string to the `ddslot777-api` Vercel project as `DATABASE_URL`.
- Redeploy the API. On boot, the API creates `users`, `deposits`, `withdrawals`, and `audit_logs` tables automatically.
- `GET /api/health` returns `"storage":"postgres"` when the database is active.

Sync rule:

- `ddslot777-admin` must read lists from this API only.
- After every write action, the admin UI should reload summary, users, deposits, withdrawals, and audit logs from the API.
- If the API is unavailable, the admin UI should show an error instead of local fallback data.

This is still a demo API. Real authentication, payment callbacks, balance ledger entries, and stricter audit controls should be added before production use.
