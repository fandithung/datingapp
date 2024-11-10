# Dating App Service

A dating app backend service built with Go that provides user matching and premium subscription features.

## Features

- User authentication (signup/login) with JWT
- Profile matching system
- Daily interaction limits (10 per day for non-premium users)
- Premium subscription features

## Project Structure
```
.
├── cmd
│   └── api          # Application entrypoint
├── internal
│   ├── config       # Configuration management
│   ├── handler      # HTTP handlers
│   ├── middleware   # HTTP middleware
│   ├── repository   # Database operations
│   └── service      # Business logic
├── migrations       # Database migrations
├── docker-compose.yml   # Docker compose configuration
└── Makefile        # Build and development commands
```

## Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15
- Make

## Getting Started

1. Clone the repository:
```bash
git clone {{repo_url}}
cd dating-app
```

2. Install development tools:
```bash
make install-tools
```

3. Start the database and other services:
```bash
make docker-up
```
4. Run database migrations:
```bash
make migrate-up
```
5. Run the application with hot reloading:
```bash
make run-dev
```

## Available commands
- `make install-tools`: Install development tools
- `make run`: Run the service
- `make run-dev`: Run the service with hot reload
- `make build`: Build the service
- `make test`: Run tests
- `make migrate-up`: Apply database migrations
- `make migrate-down`: Rollback database migrations
- `make seed`: Seed the database with sample data
- `make docker-up`: Start all services
- `make docker-down`: Stop all services
- `make lint`: Run linter

## API Endpoints

### Public Endpoints
- `POST /api/v1/signup`: Create new user account
- `POST /api/v1/login`: Authenticate user and get JWT token

### Protected Endpoints (requires JWT)
- `GET /api/v1/profiles`: Get candidate profiles
- `POST /api/v1/profiles/:id/response`: Respond to a profile (like/pass)
- `GET /api/v1/features`: List available premium features
- `GET /api/v1/features/my`: Get user's active features
- `POST /api/v1/features/:id/subscribe`: Subscribe to a premium feature

## Linter
We use [golangci-lint](https://golangci-lint.run/usage/install/) to lint the code.
