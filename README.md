# ddslot777-api

Go API service for the DD Slot admin system.

Live API:

- https://ddslot777-api.vercel.app

Contract:

- API version: `v1`
- Admin contract version: `admin-mvp-v1`
- Contract endpoint: `GET /api/contract`

Current endpoints:

- `GET /api/health`
- `GET /api/contract`
- `POST /api/auth/login`
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

Phase 1 admin loop:

- Admin login
- User list
- Deposit order list and confirm deposit
- Withdrawal order list and approve/reject withdrawal
- Manual user balance update
- Simple operation/audit log

Sync rule:

- `ddslot777-admin` must read lists from this API only.
- After every write action, the admin UI should reload summary, users, deposits, withdrawals, and audit logs from the API.
- If the API is unavailable, the admin UI should show an error instead of local fallback data.

This is still a demo API backed by in-memory data. Database persistence, real authentication, payment callbacks, balance ledger entries, and durable audit logs should be added before production use.
