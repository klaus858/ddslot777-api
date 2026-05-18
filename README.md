# ddslot777-api

Go API service for the DD Slot admin system.

Current demo endpoints:

- `GET /health`
- `POST /auth/login`
- `GET /admin/summary`
- `GET /admin/members`
- `GET /admin/deposits`
- `POST /admin/deposits/confirm`

Demo login payload:

```json
{
  "username": "admin",
  "password": "admin123"
}
```

This is the first mock API version. Database, real authentication, payment callbacks, and audit logs should be added before production use.
