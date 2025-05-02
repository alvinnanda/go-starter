# Starter App

A Go application built with Go Fiber.

## Requirements

- Go 1.23.5
- Go Fiber

## Project Structure

```
/starter-app
├── cmd/                # Main applications
│   └── server/
│       └── main.go
├── config/             # Configuration files
│   └── config.go
├── internal/           # Private application code
│   ├── handler/
│   ├── service/
│   ├── repository/
│   ├── model/
│   └── middleware/
├── pkg/                # Reusable packages
│   └── logger/
│   └── jwt/            # JWT utilities
├── routes/             # Router setup
│   └── routes.go
├── scripts/            # Build/deploy scripts (optional)
├── go.mod
├── go.sum
└── README.md
```

## Getting Started

1. Clone the repository
2. Install dependencies: `go mod download`
3. Run the application: `go run cmd/server/main.go`

## Authentication

The app includes a simple JWT-based authentication system:

- Register: `POST /auth/register` with body `{"name": "...", "email": "...", "password": "..."}`
- Login: `POST /auth/login` with body `{"email": "...", "password": "..."}`
- Get current user: `GET /api/v1/users/me` (requires token)

For protected endpoints, include the JWT token in the Authorization header:
```
Authorization: Bearer your_token_here
```

Default admin credentials:
- Email: admin@example.com
- Password: admin123

## Security Features

The application includes the following security features:

- Password strength validation (minimum 8 chars, uppercase, lowercase, numbers, special chars)
- Rate limiting to prevent brute force attacks
- HTTP security headers (CSP, XSS protection, HSTS, etc.)
- CORS configuration with secure defaults
- JWT token with enhanced security (token revocation, secure storage)
- Content type validation
- Secure password hashing with bcrypt
- Protection against common vulnerabilities (XSS, CSRF, etc.)
- Body size limits to prevent DOS attacks
- Error handling that prevents information leakage

## API Testing Guide

### Base URL
```
http://localhost:8086
```

### Public Endpoints

#### Health Check
```
GET /health
```

Response:
```json
{
  "status": "ok"
}
```

#### Welcome
```
GET /api/v1
```

Response:
```json
{
  "message": "Welcome to the API"
}
```

### Authentication Endpoints

#### Register
```
POST /auth/register
```

Request body:
```json
{
  "name": "Test User",
  "email": "user@example.com",
  "password": "password123"
}
```

Response:
```json
{
  "message": "User registered successfully. Please login to continue.",
  "user": {
    "id": "...",
    "name": "Test User",
    "email": "user@example.com",
    "role": "user"
  }
}
```

#### Login
```
POST /auth/login
```

Request body:
```json
{
  "email": "admin@example.com",
  "password": "admin123"
}
```

Response:
```json
{
  "message": "Login successful",
  "token": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "token_type": "Bearer"
  },
  "user": {
    "id": "...",
    "name": "Admin User",
    "email": "admin@example.com",
    "role": "admin"
  }
}
```

#### Refresh Token
```
POST /auth/refresh
```

Request body:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Response:
```json
{
  "message": "Token refreshed successfully",
  "token": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "token_type": "Bearer"
  }
}
```

#### Logout
```
POST /auth/logout
```

Request body:
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

Response:
```json
{
  "message": "Logged out successfully"
}
```

### Protected Endpoints

For all protected endpoints, you need to include the JWT token in the Authorization header:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

#### Get Current User
```
GET /api/v1/users/me
```

Response:
```json
{
  "user": {
    "id": "...",
    "name": "Admin User",
    "email": "admin@example.com",
    "role": "admin"
  }
}
```

#### Get All Users (Admin only)
```
GET /api/v1/users
```

Response:
```json
{
  "message": "This would return all users"
}
```

#### Get User by ID
```
GET /api/v1/users/:id
```

Response:
```json
{
  "message": "This would return user with ID: :id"
}
```

#### Create User (Admin only)
```
POST /api/v1/users
```

Response:
```json
{
  "message": "This would create a new user"
}
```

#### Update User
```
PUT /api/v1/users/:id
```

Response:
```json
{
  "message": "This would update user with ID: :id"
}
```

#### Delete User (Admin only)
```
DELETE /api/v1/users/:id
```

Response:
```json
{
  "message": "This would delete user with ID: :id"
}
```

## Configuration

The application can be configured via environment variables:

- `PORT`: The port the server runs on (default: 8080)
- `ENV`: The environment (development, production) (default: development)
- `DB_HOST`: Database host (default: localhost)
- `DB_PORT`: Database port (default: 5432)
- `DB_USER`: Database user (default: postgres)
- `DB_PASSWORD`: Database password (default: postgres)
- `DB_NAME`: Database name (default: starter_app)
- `JWT_SECRET`: Secret key for signing JWT tokens (default: your-secret-key-change-in-production)
- `TOKEN_EXPIRY`: Access token expiry time (e.g., "24h", "30m") (default: 24h)
- `REFRESH_EXPIRY`: Refresh token expiry time (e.g., "168h", "30d") (default: 168h or 7 days)

### Database Connection Pool Configuration

- `DB_POOL_MAX_CONNS`: Maximum number of connections in the pool (default: 10)
- `DB_POOL_MIN_CONNS`: Minimum number of connections in the pool (default: 2)
- `DB_POOL_MAX_CONN_LIFETIME`: Maximum lifetime of a connection (default: 1h)
- `DB_POOL_MAX_CONN_IDLE_TIME`: Maximum idle time for a connection (default: 30m)
- `DB_POOL_HEALTH_CHECK_PERIOD`: How often to check connection health (default: 1m)

These settings allow you to fine-tune the database connection pool based on your workload and available resources.

### Cache Configuration

- `CACHE_DRIVER`: Cache implementation to use (options: "memory", "redis") (default: memory)
- `CACHE_TTL`: Default time-to-live for cached items (default: 5m)
- `REDIS_ADDRESS`: Redis server address (default: localhost:6379)
- `REDIS_PASSWORD`: Redis server password (default: empty)
- `REDIS_DB`: Redis database number (default: 0)

## License

[MIT](LICENSE)
