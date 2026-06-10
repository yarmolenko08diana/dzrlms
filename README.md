# CorpLMS — Corporate Learning Management System

A production-ready MVP for managing corporate employee training.  
Built with **Go + Gin + PostgreSQL + Bootstrap**.

---

## Tech Stack & Architecture Decisions

| Layer      | Choice            | Reason |
|------------|-------------------|--------|
| Backend    | Go + Gin          | Fast, minimal, great for MVPs |
| Database   | PostgreSQL + GORM | Relational integrity; GORM reduces boilerplate |
| Auth       | **Session-based** | Simpler for server-rendered HTML; no token refresh complexity |
| Frontend   | HTML + Bootstrap  | No build step, fast to iterate |

**Why sessions over JWT?**  
This app uses server-rendered HTML templates (not an SPA). Sessions are the natural fit — no CORS headaches, instant logout that actually works, and no token storage issues on the client. JWT shines when you have a separate frontend or mobile app consuming an API.

---

## Project Structure

```
lms/
├── main.go                    # Entry point: DB + server setup
├── go.mod
├── schema.sql                 # Reference SQL schema
├── db/
│   └── database.go            # Connect, Migrate, Seed
├── models/
│   └── models.go              # User, Course, Assignment structs
├── middleware/
│   └── auth.go                # RequireAuth, RequireAdmin, InjectUser
├── controllers/
│   ├── auth.go                # Login, Logout
│   ├── dashboard.go           # Admin + Employee dashboards
│   ├── employee.go            # CRUD for employees
│   ├── course.go              # CRUD for courses
│   └── assignment.go          # Assignment management
├── routes/
│   └── routes.go              # All route definitions
└── templates/
    ├── auth/login.html
    ├── admin/
    │   ├── layout.html         # Shared CSS/styles
    │   ├── dashboard.html
    │   ├── employees.html
    │   ├── employee_form.html
    │   ├── employee_profile.html
    │   ├── courses.html
    │   ├── course_form.html
    │   ├── assignments.html
    │   └── assignment_form.html
    └── employee/
        └── dashboard.html
```

---

## Database Design

```
users
  id, name, email, password (bcrypt), role (admin|employee), created_at, updated_at

courses
  id, title, description, duration, created_at, updated_at

assignments
  id, user_id (FK→users), course_id (FK→courses),
  status (not_started|in_progress|completed),
  created_at, updated_at
  UNIQUE(user_id, course_id)
```

**Relationships:**
- `users` → `assignments`: One-to-many (one employee, many assignments)
- `courses` → `assignments`: One-to-many (one course, many assignments)
- `users` ↔ `courses`: Many-to-many through `assignments`

---

## Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Git

---

## Setup Instructions

### 1. Clone / create the project

```bash
git clone <your-repo> lms
cd lms
```

### 2. Create the PostgreSQL database

```bash
psql -U postgres
CREATE DATABASE lms_db;
\q
```

### 3. Install Go dependencies

```bash
go mod tidy
```

### 4. Configure environment (optional)

The app uses these env vars with defaults:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=lms_db
export PORT=8080
```

Or create a `.env` file and source it:
```bash
source .env
```

### 5. Run the application

```bash
go run main.go
```

The app will:
- Connect to PostgreSQL
- Run GORM AutoMigrate (creates tables automatically)
- Seed demo data (admin + 4 employees + 5 courses + assignments)

Open: **http://localhost:8080**

---

## Demo Credentials

| Role     | Email                  | Password     |
|----------|------------------------|--------------|
| Admin    | admin@company.com      | admin123     |
| Employee | alice@company.com      | password123  |
| Employee | bob@company.com        | password123  |
| Employee | carol@company.com      | password123  |
| Employee | david@company.com      | password123  |

---

## Feature Walkthrough

### Admin Panel (`/admin/...`)

| URL                          | Feature                    |
|------------------------------|----------------------------|
| `/admin/dashboard`           | Stats + recent activity    |
| `/admin/employees`           | List all employees         |
| `/admin/employees/new`       | Create employee            |
| `/admin/employees/:id/edit`  | Edit employee              |
| `/admin/employees/:id/profile` | View profile + courses   |
| `/admin/courses`             | List all courses           |
| `/admin/courses/new`         | Create course              |
| `/admin/courses/:id/edit`    | Edit course                |
| `/admin/assignments`         | View all assignments       |
| `/admin/assignments?status=in_progress` | Filter by status |
| `/admin/assignments/new`     | Assign course to employee  |

### Employee Panel (`/employee/...`)

| URL                          | Feature                    |
|------------------------------|----------------------------|
| `/employee/dashboard`        | View assigned courses + update status |

---

## API Reference (for Postman or curl)

Since this is a session-based app, use a cookie jar in Postman.

### 1. Login
```
POST /login
Content-Type: application/x-www-form-urlencoded

email=admin@company.com&password=admin123
```
→ Redirects to `/admin/dashboard`, sets session cookie.

### 2. Create Employee
```
POST /admin/employees
Content-Type: application/x-www-form-urlencoded
Cookie: lms_session=<your-session-cookie>

name=John%20Doe&email=john@company.com&password=secret123
```

### 3. Create Course
```
POST /admin/courses
Content-Type: application/x-www-form-urlencoded
Cookie: lms_session=<your-session-cookie>

title=New%20Course&description=Course%20description&duration=2%20hours
```

### 4. Assign Course to Employee
```
POST /admin/assignments
Content-Type: application/x-www-form-urlencoded
Cookie: lms_session=<your-session-cookie>

user_id=2&course_id=3
```

### 5. Update Assignment Status (Admin)
```
POST /admin/assignments/:id/status
Content-Type: application/x-www-form-urlencoded
Cookie: lms_session=<your-session-cookie>

status=completed
```

### 6. Update Assignment Status (Employee)
```
POST /employee/assignments/:id/status
Content-Type: application/x-www-form-urlencoded
Cookie: lms_session=<employee-session-cookie>

status=in_progress
```

### 7. Delete Employee
```
POST /admin/employees/:id/delete
Cookie: lms_session=<your-session-cookie>
```

### 8. Logout
```
GET /logout
```

### Postman Collection Setup
1. Create a new Collection in Postman
2. Add a Collection variable `baseUrl = http://localhost:8080`
3. Enable **"Automatically follow redirects"** in settings
4. Enable **"Send cookies"** — Postman will manage the session cookie after login

---

## Security Notes

- Passwords hashed with **bcrypt** (cost factor 10)
- Routes protected by session middleware
- Admin routes further protected by role check
- Duplicate assignments are silently ignored (idempotent)
- Delete operations cascade to assignments (no orphaned records)

---

## Production Checklist

Before deploying:
- [ ] Change `lms-secret-key-change-in-production` in `routes/routes.go` to a random 32+ byte secret
- [ ] Set `GIN_MODE=release`
- [ ] Use environment variables for all DB credentials
- [ ] Add HTTPS (use nginx reverse proxy or Caddy)
- [ ] Consider adding CSRF protection (`github.com/utrack/gin-csrf`)
- [ ] Add rate limiting on the login endpoint

---

## Extending the System

**Add email notifications:** Hook into the assignment Create handler, send via SMTP using `net/smtp`.

**Add file uploads (course materials):** Use `c.FormFile()` in Gin + store to disk or S3.

**Add pagination:** Pass `page` query param, use GORM's `.Offset().Limit()`.

**Add API endpoints (JSON):** Add a `/api/v1/` route group that returns JSON instead of HTML — useful for a future mobile app.
