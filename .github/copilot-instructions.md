# Copilot Instructions for UAS_GO

## Project Overview
Academic Achievement Management System (UAS) backend built with Go using Fiber framework, PostgreSQL (relational data), and MongoDB (document storage). Three-tier role system: admin, mahasiswa (students), dosen_wali (academic advisors).

## Architecture

### Database Strategy (Hybrid)
- **PostgreSQL**: User accounts, roles, permissions, students, lecturers, achievement references (verification status)
  - Uses UUID for all IDs
  - Key table: `users` (id, email, password_hash, role_id, is_active)
- **MongoDB**: Raw achievement documents (flexible schema for diverse achievement types)
  - Collection: `achievements` (nested details, attachments, tags, points)
- **Pattern**: Achievement references in PostgreSQL link students to MongoDB achievements; enables verification workflow

### Service Layer Pattern
- `AuthService` (in `/app/service/auth_serv.go`): Handles authentication logic
  - Methods: `Login()` returns `LoginResponse` (user + JWT token)
  - `VerifyToken()` validates JWT tokens for middleware
  - Uses `helper.CheckPassword()` for bcrypt password verification
- Services directly query PostgreSQL via `database.PSQL` (package-level var)
- MongoDB queried via `database.DB` (*mongo.Database)

### HTTP Handler Pattern (Fiber)
- Routes defined in `/route/auth_route.go` → groups under `/api/v1/alumni`
- Handler function receives `*fiber.Ctx` and service instance
- Uses `helper.APIResponse()` for standardized responses: `{status, message, data}`
- Predefined helpers: `BadRequest()`, `Unauthorized()`, `NotFound()`, `InternalError()`

### Authentication Flow
1. POST `/api/v1/alumni/login` with `{email, password}`
2. Service finds user in PostgreSQL, verifies bcrypt hash
3. JWT token generated with claims: `{user_id, email, role}`
4. Response: `{user: {...}, token: "jwt_string"}`
5. Token expiry: 24 hours (configurable via `JWT_SECRET` env var)

## Project Structure
```
app/
  models/          # Structs for requests/responses/db entities
  service/         # Business logic (auth, achievements, etc.)
  repository/      # [Empty] Reserved for data access layer
database/
  connection_postgres.go  # PSQL = *sql.DB global var
  connection_mongo.go     # DB = *mongo.Database global var
  migrate.go              # MigrateTesting() seeds roles, users, permissions
route/
  auth_route.go           # Route handlers
  route.go                # RegisterRoutes() entry point
helper/
  password.go             # HashPassword(), CheckPassword() via bcrypt
  response.go             # APIResponse(), BadRequest(), etc.
config/
  env.go                  # LoadEnv(), GetEnv()
  logger.go               # InitLogger() to logs/app.log
  app.go                  # NewApp() creates Fiber instance
```

## Key Conventions

### Error Handling
- Service methods return `(data, error)`; handlers call `helper.APIResponse()` with status codes
- No panic; return descriptive error messages to client
- Example: Missing user → `Unauthorized(c, "user not found")`

### ID Generation
- Use `github.com/google/uuid` for all new IDs (UUID v4 string)
- Example in migrate.go: `adminRoleID := uuid.New().String()`

### Environment Variables (.env)
Required vars for auth:
- `JWT_SECRET`: Sign/verify tokens (defaults to placeholder—CHANGE IN PRODUCTION)
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASS`, `DB_NAME`: PostgreSQL
- `MONGO_URI`, `MONGO_DB_NAME`: MongoDB
- `APP_PORT`: Server port (default 3000)

### Password Security
- **Always** use `helper.HashPassword()` when storing; returns bcrypt hash
- **Always** use `helper.CheckPassword(plaintext, hash)` when verifying
- Example in migrate.go: `adminHash, _ := helper.HashPassword("123456")`

## Testing & Running

### Build & Run
- Development: `go run main.go`
- Build: `go build -o UAS_GO.exe`
- Format: `go fmt ./...`
- Lint: `go vet ./...`

### Database Setup
- Ensure PostgreSQL running on `localhost:5432`
- Ensure MongoDB running on `localhost:27017`
- First startup: uncomment `database.MigrateTesting(database.PSQL)` in main.go to seed test data
  - Creates: admin, dosen1, mhs1 users (password: "123456")
  - Creates: 3 roles, 6 permissions, role-permission mappings

### API Testing Login Endpoint
```bash
curl -X POST http://localhost:3000/api/v1/alumni/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"123456"}'
```

## Adding New Features

### New Endpoint with Authentication
1. Add handler in `/route/auth_route.go` (or create new route file)
2. Create service method in `/app/service/` (matches business domain)
3. Use `VerifyToken()` in handler to extract claims before protected logic
4. Query data from PSQL or MongoDB as needed
5. Return via `helper.APIResponse()`

### New Service
- Create `{domain}_serv.go` in `/app/service/`
- Constructor: `NewXService() *XService { return &XService{} }`
- Methods use `database.PSQL` or `database.DB` package-level globals
- No dependency injection; services are simple functions on implicit state

### Role-Based Access
- Load user role from JWT claims (`claims.Role`)
- Example: Check if role == "admin" before allowing dangerous operations
- Permission system exists in DB (`role_permissions`, `permissions` tables) but currently unused

## Critical Patterns
- **No repository layer active**: Queries written inline in services; can migrate to repositories later
- **Single-responsibility**: Each service handles one domain; routes dispatch to services
- **Stateless handlers**: Fiber context passed through; no global request state beyond database connections
- **No middleware implemented yet**: Routes in comments show where to add auth checks (e.g., `middleware.AuthMiddleware()`)

---
*Last updated: November 19, 2025*
