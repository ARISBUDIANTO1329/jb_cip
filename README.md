# Creator Intelligence Platform (CIP) — Backend

Backend untuk Creator Intelligence Platform.

## Requirements

- Go 1.22+
- Docker & Docker Compose
- PostgreSQL 16+
- Redis 7+

## Struktur Project

```
backend/
├── api/            # HTTP layer (routes, handlers, middleware, DTO)
├── cmd/            # Entry point (api, migrate, seed)
├── config/         # Configuration loader
├── internal/       # Business logic (domain, service, repository, usecase)
├── pkg/            # Reusable packages
├── migrations/     # Database migrations
├── seeds/          # Seed data
├── scripts/        # Helper scripts
├── tests/          # Unit & integration tests
├── docker/         # Dockerfile & docker-compose.yml
├── .env.example    # Environment variables template
├── Makefile        # Common commands
└── go.mod          # Go module
```

## Setup

### 1. Install Go

```bash
# Go sudah terinstall di /usr/local/go
export PATH=$PATH:/usr/local/go/bin
```

### 2. Copy Environment File

```bash
cp .env.example .env
# Edit .env sesuai kebutuhan
```

### 3. Start PostgreSQL & Redis

```bash
make docker-up
```

### 4. Run Migrations

```bash
# Build migration tool
make build

# Run migrations
make migrate-up

# Check version
make migrate-version

# Rollback
make migrate-down
```

### 5. Run Backend

```bash
make dev
```

Server akan berjalan di `http://localhost:8082`.

## Health Check

```bash
curl http://localhost:8082/health
curl http://localhost:8082/api/v1/health
```

## Database

### Schema

| Schema | Tables |
|---|---|
| core | users, workspaces, workspace_members, roles, permissions, role_permissions, workspace_settings |

### Migration Commands

| Command | Description |
|---|---|
| `make migrate-up` | Run all pending migrations |
| `make migrate-down` | Rollback all migrations |
| `make migrate-version` | Show current migration version |
| `make migrate-drop` | Drop all tables |

### Migration Files

| File | Description |
|---|---|
| `000001_init_core_schema.up.sql` | Create core tables |
| `000001_init_core_schema.down.sql` | Drop core tables |
| `000002_seed_data.up.sql` | Insert seed data |
| `000002_seed_data.down.sql` | Remove seed data |

### Seed Data

| Data | Count |
|---|---|
| Super Admin | 1 (admin@cip.local) |
| Permissions | 27 |
| Default Roles | Created per workspace |

## API Endpoints

| Method | Endpoint | Description |
|---|---|---|
| GET | `/health` | Health check |
| GET | `/api/v1/health` | Health check with version |

## Makefile Commands

| Command | Description |
|---|---|
| `make dev` | Run development server |
| `make build` | Build binaries |
| `make test` | Run tests |
| `make docker-up` | Start PostgreSQL & Redis |
| `make docker-down` | Stop PostgreSQL & Redis |
| `make migrate-up` | Run migrations |
| `make migrate-down` | Rollback migrations |
| `make migrate-version` | Show migration version |
| `make seed` | Run seed data |
| `make clean` | Clean build artifacts |
