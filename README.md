# ddslot777-api

Go API service for the DD Slot admin system.

Live API:

- https://ddslot777-api.vercel.app

Contract:

- API version: `v1`
- Admin contract version: `admin-v1`
- Contract endpoint: `GET /api/contract`

Current endpoints:

- `GET /api/health`
- `GET /api/contract`
- `POST /api/auth/login`
- `GET /api/admin/summary`
- `GET /api/admin/members`
- `GET /api/admin/deposits`
- `POST /api/admin/deposits/confirm`

The non-prefixed endpoints are also supported for simple checks:

- `GET /health`
- `GET /contract`

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

Compatibility rules:

- Keep existing endpoint paths and field names stable for `admin-v1`.
- Add new fields instead of renaming or deleting existing fields.
- Add a new contract version before making breaking changes.
- Keep status values stable: member `normal`, `risk_review`; deposit `pending`, `credited`, `risk_review`.

This is still a demo API backed by in-memory data. Database persistence, real authentication, payment callbacks, balance ledger entries, and audit logs should be added before production use.
