# Kleff Auth
Open-source, Go-based authentication service (Firebase-like) using **Hexagonal Architecture**. Ships as a single modular monolith with clean ports/adapters, multi-tenant user management, JWT access tokens (Ed25519), rotating refresh tokens, and an extensible attributes system.


## ✨ Features (MVP)
- Email/Username + Password sign-up/sign-in
- Short-lived **JWT access** + rotating **opaque refresh** tokens
- **Multi-tenant** isolation (e.g., `portfolio`, `kleff`)
- User custom attributes (`users.attrs` JSONB) with optional JSON Schema validation
- `/me`, `/token`, `/jwks.json`, `/healthz`
- Hexagonal architecture: domain → application/use-cases → adapters (HTTP/Postgres/Keys)


## 🧭 Roadmap (near-term)
- Email verification + password reset
- Rate limiting (Redis)
- Google/GitHub OAuth
- Admin UI (React + Tailwind)


## 📦 Quick start
```bash
make dev # run with docker compose (db/redis/auth soon)
# or
go run ./cmd/authd # run the server (no DB features yet)
```


### 🗂️ Project layout (hexagonal)
```
cmd/authd/                      # composition root
internal/
    domain/                     # entities, value objects, domain errors
    application/auth/           # use-cases + ports
    adapters/
        http/                   # handlers → use-cases
        postgres/               # repositories (implements ports)
        encryption/                 # token signer (Ed25519), JWKS
    runtime/                    # config, logger, health
    migrations/                 # SQL
docs/                           # architecture docs & diagrams
```

### 🪪 License
(To be determined)