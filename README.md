# GoLedger

A fintech wallet API written in Go — and a playground for serious test automation.

> 🚧 **Work in progress.** The project is being built step by step; this README grows with it.

## What it will do

- User registration and JWT authentication
- Multi-currency accounts, deposits and transaction history
- Money transfers with idempotency and safe concurrent processing (double-entry ledger)
- Currency exchange via an external rates API
- Limits and simple anti-fraud rules
- Virtual cards (masked numbers, Luhn check)

## Testing strategy

The main goal of the project is test coverage at every level:

| Level | Tools |
|---|---|
| Unit | Go `testing`, testify |
| Integration | testcontainers-go (real PostgreSQL in Docker) |
| API | Go API test suite with Allure reports, Postman collection |
| Fuzz | Go native fuzzing |
| Load | k6 |
| Test quality | mutation testing |
| Mobile | mobile app + UI autotests |
| CI | GitHub Actions, Allure report on GitHub Pages |

## Run locally

Requirements: Go 1.26+, Docker.

```bash
cp .env.example .env          # then set JWT_SECRET: openssl rand -hex 32
make db-up                    # PostgreSQL in Docker
make run                      # API on http://localhost:8080 (migrations run on start)
```

Other commands: `make test`, `make check` (everything CI runs), `make db-down`.

## API

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | — | Liveness check |
| `POST` | `/users` | — | Register: `{"email", "password"}` → `201` user |
| `POST` | `/auth/login` | — | Login: `{"email", "password"}` → `200` `{"access_token", "token_type", "expires_at"}` |
| `GET` | `/me` | Bearer | Current user |

Errors always look like `{"error": {"code": "...", "message": "...", "field": "..."}}`.
Codes: `invalid_json`, `validation_error`, `email_taken`, `invalid_credentials`, `unauthorized`, `internal_error`.

## Stack

Go · PostgreSQL · Docker Compose · OpenAPI
