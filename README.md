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

## Stack

Go · PostgreSQL · Docker Compose · OpenAPI
