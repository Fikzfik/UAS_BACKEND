# API Documentation

## Overview
This is a **Go/Fiber** backend service that provides REST API endpoints for the **Sistem Pelaporan Prestasi Mahasiswa** (Student Achievement Reporting System). The service uses **PostgreSQL** for structured data and **MongoDB** for specific data storage needs, along with **JWT** for authentication and specific role-based access control.

## Table of Contents
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [API Endpoints](#api-endpoints)
- [Utilities](#utilities)
- [Error Handling](#error-handling)

## Tech Stack
- **Runtime**: Go (Golang) v1.24.0
- **Framework**: Fiber v2
- **Database**: 
  - PostgreSQL (Primary)
  - MongoDB (Secondary)
- **Authentication**: JWT (JSON Web Tokens)
- **Documentation**: Swagger (via `swaggo/swag`)
- **Additional Libraries**:
  - `gofiber/fiber/v2`
  - `gorm` (implied or sqlmock/standard lib)
  - `golang.org/x/crypto`

## Getting Started

### Installation
```bash
go mod tidy
```

### Development
```bash
go run main.go
```
The server will start on `http://localhost:3000` (or the port defined in ENV).
Swagger documentation available at `http://localhost:3000/swagger/`.

### Project Structure
```
.
├── app/
│   ├── models/           # Database models
│   ├── repository/       # Database access layer
│   └── service/          # Business logic
├── config/               # Configuration (Env, Database connection)
├── database/             # Database migration and setup
├── docs/                 # Swagger documentation
├── helper/               # Utilities (Response, etc.)
├── middleware/           # Auth and Permission middleware
├── route/                # Route definitions
├── uploads/              # Static file uploads
├── go.mod
└── main.go               # Entry point
```

## Environment Variables
Create a `.env` file or set the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_PORT` | Port for the application | `3000` |
| `DB_HOST` | PostgreSQL Host | - |
| `DB_PORT` | PostgreSQL Port | `5432` |
| `DB_USER` | Database User | - |
| `DB_PASS` | Database Password | - |
| `DB_NAME` | Database Name | - |
| `MONGO_URI` | MongoDB Connection String | - |
| `JWT_SECRET` | Secret key for JWT | - |

## API Endpoints

### 1. Auth API
**Endpoint**: `POST /api/v1/auth/login`

**Description**: Authenticate user and receive JWT token.

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "secretpassword"
}
```

**Response**:
```json
{
  "status": 200,
  "message": "Login success",
  "data": {
    "token": "eyJhbGciOiJIUzI1..."
  }
}
```

### 2. Achievement API
**Endpoint**: `GET /api/v1/achievements`

**Description**: Get all achievements (requires `achievement:read` permission).

**Response**:
```json
{
  "status": 200,
  "message": "Success retrieve achievements",
  "data": [
    {
      "id": 1,
      "title": "National Hackathon Winner",
      "date": "2025-09-30",
      "status": "verified"
    }
  ]
}
```

**Endpoint**: `POST /api/v1/achievements`

**Description**: Create a new achievement record.

**Request Body**:
```json
{
  "title": "Lomba Coding",
  "description": "Juara 1 Lomba Coding Nasional",
  "date": "2025-10-01",
  "category_id": 1
}
```

**Response**:
```json
{
  "status": 201,
  "message": "Achievement created successfully",
  "data": { ... }
}
```

### 3. Achievement Workflow
- **Submit**: `POST /api/v1/achievements/:id/submit`
- **Verify**: `POST /api/v1/achievements/:id/verify` (Lecturer/Admin)
- **Reject**: `POST /api/v1/achievements/:id/reject` (Lecturer/Admin)

### 4. User Profile
**Endpoint**: `GET /api/v1/auth/profile`

**Description**: Get current user profile information.

## Utilities

### Response Format (`helper/response.go`)
The API uses a standardized JSON response format:

```go
type APIResponse struct {
    Status  int         `json:"status"`
    Message string      `json:"message"`
    Data    interface{} `json:"data"`
}
```

### CORS (`handleCors`)
Managed via Fiber's CORS middleware allowing specified origins and headers.

## Error Handling
All endpoints follow a consistent error handling pattern.

**Common Error Responses**:

- **Bad Request (400)**: Invalid input data.
- **Unauthorized (401)**: Missing or invalid JWT token.
- **Forbidden (403)**: Insufficient permissions.
- **Not Found (404)**: Resource not found.
- **Server Error (500)**: Internal system error.

**Example Error**:
```json
{
  "status": 401,
  "message": "Unauthorized access",
  "data": null
}
```

## Security Considerations
- **JWT Authentication**: All protected routes require a valid Bearer token.
- **Role-Based Access Control (RBAC)**: Middleware checks for specific permissions (e.g., `achievement:verify`).
- **CORS Protection**: Restricted to allowed origins.
- **Input Validation**: Request bodies are validated before processing.

---

**Contact**: [mailto:faridfathonin@email.com](mailto:faridfathonin@email.com)  
**Portfolio**: [https://portofolio-fridfn.vercel.app](https://portofolio-fridfn.vercel.app)
