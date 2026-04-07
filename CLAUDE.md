# distributed-ticketing

Distributed unique ID generation across multiple app servers backed by a single MySQL ticket server.

## Architecture

- **App server** (`cmd/appserver`) — serves `GET /id`. Uses a `TicketAllocator` to hand out IDs from an in-memory range, refilling from the ticket server when exhausted.
- **Ticket server** (`cmd/ticketserver`) — serves `POST /reserve`. Atomically increments `end_range` in MySQL and returns the new `[start, end)` range.
- **Allocator** (`internal/allocator`) — goroutine-safe, lock-free hot path; only takes a mutex when the local range is exhausted and a refill is needed.
- **DB** (`internal/db`) — wraps MySQL. `IncrementEndRange` uses a transaction with `SELECT ... FOR UPDATE` to atomically claim a range.

## Building

```bash
go build -o bin/appserver ./cmd/appserver
go build -o bin/ticketserver ./cmd/ticketserver
```

## Database setup

```bash
bash scripts/setup_db.sh
export DB_DSN="root:<password>@tcp(127.0.0.1:3306)/ticketing"
```

## Running

```bash
# Terminal 1
DB_DSN="..." ./bin/ticketserver

# Terminal 2+
./bin/appserver
```

## Tests

Unit tests (no external dependencies):
```bash
go test ./internal/...
```

Integration tests (require MySQL via `DB_DSN`):
```bash
DB_DSN="..." go test ./tests/integration/...
```

## Test conventions

- Unit tests live alongside the source file they test (e.g. `internal/allocator/allocator_test.go`).
- Integration tests that require external services (MySQL) live in `tests/integration/`.
