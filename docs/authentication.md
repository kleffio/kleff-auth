# Authentication System Documentation

## Table of Contents
- [Overview](#overview)
- [Core Concepts](#core-concepts)
- [Architecture](#architecture)
- [Authentication Flows](#authentication-flows)
- [API Endpoints](#api-endpoints)
- [Configuration](#configuration)
- [Security](#security)
- [Development](#development)
- [Deployment](#deployment)

## Overview
Kleff Auth is a secure, multi-tenant authentication service built with Go. It provides user authentication, session management, and token-based security using JWT (JSON Web Tokens).

## Core Concepts

### 1. Users
- Represent individual user accounts
- Can authenticate via email/username and password
- Belong to a specific tenant
- Support custom attributes

### 2. Tenants
- Represent separate organizations or workspaces
- Provide isolation between different groups of users
- Each tenant has  its own set of users and sessions

### 3. Sessions
- Track user authentication state
- Support multiple concurrent sessions per user
- Include security context (IP, User-Agent)
- Support refresh token rotation

## Architecture

The system follows a clean/hexagonal architecture with the following layers:

### 1. Domain Layer
- Core business entities and logic
- Defines interfaces (ports) for external systems

### 2. Application Layer
- Implements core business logic
- Defines use cases and workflows
- Depends on domain interfaces

### 3. Adapters Layer
- Implements interfaces defined in the domain layer
- Handles external concerns like:
  - HTTP API
  - Database access
  - Cryptography

## Authentication Flows

### 1. User Registration
1. Client sends registration request with email/username and password
2. System validates input and checks for existing users
3. Password is hashed and stored
4. New user account is created
5. Access and refresh tokens are generated

### 2. User Login
1. Client sends login request with credentials
2. System verifies credentials
3. New session is created
4. Access and refresh tokens are issued

### 3. Token Refresh
1. Client presents expired access token and valid refresh token
2. System validates the refresh token
3. New access and refresh tokens are issued
4. Old refresh token is invalidated

## API Endpoints

### Authentication
- `POST /auth/signup` - Register a new user
- `POST /auth/signin` - Authenticate and get tokens
- `POST /auth/refresh` - Refresh access token
- `POST /auth/logout` - Invalidate current session
- `POST /auth/logout-all` - Invalidate all user sessions
- `GET /auth/me` - Get current user info

## Configuration

### Environment Variables
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `JWT_PRIVATE_KEY`: Path to JWT private key
- `JWT_PUBLIC_KEY`: Path to JWT public key
- `ACCESS_TOKEN_TTL`: Access token TTL (default: 15m)
- `REFRESH_TOKEN_TTL`: Refresh token TTL (default: 7d)

## Security

### Password Security
- Uses bcrypt for password hashing
- Enforces minimum password requirements
- Protects against timing attacks

### Session Security
- Implements refresh token rotation
- Tracks device information
- Supports session revocation

### JWT Security
- Asymmetric key signing (RS256)
- Short-lived access tokens
- Secure token storage in HTTP-only cookies

## Development

### Prerequisites
- Go 1.20+
- PostgreSQL 13+
- Make

### Setup
1. Clone the repository
2. Copy `.env.example` to `.env` and configure
3. Run `make deps` to install dependencies
4. Run `make test` to run tests
5. Run `make run` to start the server

### Testing
- Unit tests: `make test`
- Integration tests: `make test-integration`
- All tests: `make test-all`

## Deployment

### Docker
```bash
docker-compose up -d
```

### Kubernetes
```bash
kubectl apply -f k8s/
```
