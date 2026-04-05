# Distributed Ticketing

A distributed unique ID generation system. Each app server maintains an in-memory atomic counter backed by a reserved range from a central ticket server, which persists state to MySQL.

See [docs/design.md](docs/design.md) for the full design.

## Architecture

```
Client → App Server (atomic counter, range [start, end))
                ↓ range exhausted
         Ticket Server (reserves next range)
                ↓
          MySQL (tickets table, end_range)
```

- **App Server** — serves unique `int64` IDs from an in-memory range. Fast path is a single atomic increment; slow path calls the ticket server to reserve a new range.
- **Ticket Server** — atomically advances `end_range` in MySQL and returns the new range to the app server.
- **MySQL** — single leader instance; holds one row tracking the global `end_range`.

## Prerequisites

- Go 1.22+
- MySQL (local: `brew install mysql`)

## Setup

**1. Start MySQL**
```bash
brew services start mysql
```

**2. Create the database and table**
```bash
./scripts/setup_db.sh
```

**3. Configure environment**
```bash
cp .env.example .env
# Edit .env and set your MySQL password
```

**4. Export env vars**
```bash
export $(cat .env | xargs)
```

## Running

Build both binaries:
```bash
go build -o bin/ ./cmd/...
```

Start the ticket server first, then the app server:
```bash
./bin/ticketserver  # listens on :8081
./bin/appserver     # listens on :8080
```

Request a unique ID:
```bash
curl http://localhost:8080/id
```

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `DB_DSN` | MySQL DSN for the ticket server | `root:password@tcp(127.0.0.1:3306)/ticketing` |

In production, supply `DB_DSN` from your secret manager instead of a `.env` file.

## Project Structure

```
cmd/
  appserver/       # app server binary
  ticketserver/    # ticket server binary
internal/
  allocator/       # TicketAllocator: atomic fast path + mutex-guarded refill
  ticketserver/    # RangeReserver: reserves ID ranges from MySQL
  db/              # MySQL store
  types/           # shared types (RangeResponse)
scripts/
  setup_db.sh      # creates database and tickets table
docs/
  design.md        # system design
```
