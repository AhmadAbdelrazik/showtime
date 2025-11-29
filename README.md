# **Showtime 🎬**

*A Go REST API for managing movie theaters, halls, movies, and show schedules.*

Showtime is a backend service built with **Go**, **Gin**, and **PostgreSQL**.
It provides authentication, theater and hall management, movie management, and conflict-free movie show scheduling.
The project follows a modular architecture (`cmd/`, `internal/`, `pkg/`), uses database migrations, and includes Swagger documentation.

---

## 🚀 **Features**

* **User Authentication** (Signup, Login, Logout, JWT-based)
* **Theater Management**
* **Hall Management** (per theater)
* **Movie Management**
* **Show Scheduling**

  * With **automatic show-time conflict detection**
* **PostgreSQL integration** with migration system
* **Swagger auto-generated API docs**
* **Modular folder structure** (`controllers`, `models`, `pkg`, etc.)

---

## 📂 **Project Structure**

```
.
├── cmd
│   └── api
│       ├── docs/               # Swagger docs
│       └── main.go             # App entry point
├── internal
│   ├── controllers/            # Handlers, routes, middleware
│   ├── httputil/               # Error helpers
│   └── models/                 # Business logic & DB models
├── pkg
│   ├── cache/                  # In-memory caching helpers
│   └── validator/              # Input validation
├── migrations/                 # SQL migration files
├── Makefile                    # Dev utilities
├── go.mod
└── go.sum
```

---

## 📚 **API Endpoints**

### **Auth**

```
POST   /api/signup
POST   /api/login
GET    /api/logout
GET    /api/user-info        (auth required)
```

---

### **Theaters**

```
GET    /api/theaters
GET    /api/theaters/:id
POST   /api/theaters         (auth required)
PATCH  /api/theaters/:id     (auth required)
DELETE /api/theaters/:id     (auth required)
```

---

### **Halls**

```
GET    /api/theaters/:id/halls/:code
POST   /api/theaters/:id/halls          (auth required)
PATCH  /api/theaters/:id/halls/:code    (auth required)
DELETE /api/theaters/:id/halls/:code    (auth required)
```

---

### **Movies**

```
GET    /api/movies
GET    /api/movies/:id
POST   /api/movies             (auth required)
PATCH  /api/movies/:id         (auth required)
DELETE /api/movies/:id         (auth required)
```

---

### **Shows**

*(Under Development)*

---

## 🧰 **Makefile Commands**

### **Run Migrations**

```bash
make db-up      # Apply migrations
make db-down    # Roll back migrations
```

### **Create Migration**

```bash
make db-migration name=add_new_table
```

### **Run Tests**

Runs migrations on test DB → runs all tests → clears DB.

```bash
make app-test
```

### **PostgreSQL CLI**

```bash
make psql
```

### **Show Current DSN**

```bash
make info
```

### **Generate Swagger Docs**

```bash
make swagger
```

---

## 🛠️ **Running Locally**

### **1. Install dependencies**

```bash
go mod tidy
```

### **2. Start database (Docker recommended)**

Example:

```bash
docker run --name showtime-db -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=showtime \
  -p 5432:5432 -d postgres:15
```

### **3. Apply migrations**

```bash
make db-up
```

### **4. Run the API**

```bash
go run cmd/api/main.go
```

---

## 📄 **Environment Variables**

Create a `.env` file:

```
DB_DSN=postgres://postgres:password@localhost:5432/showtime?sslmode=disable
DB_DSN_TEST=postgres://postgres:password@localhost:5432/showtime_test?sslmode=disable

DB_DATABASE=showtime
```

---

## 📘 **Swagger Documentation**

After generating with:

```bash
make swagger
```

Swagger UI will be served automatically when you run the API.

---

## 🧩 **Tech Stack**

* **Go**
* **Gin**
* **PostgreSQL**
* **go-migrate**
* **Swagger (swaggo)**
* **Docker**
* Clean architecture (controllers → services → models)

---

## 🗺️ **Future Enhancements**

* Payment integration (Paymob)
* Show endpoints (create/update showtimes)
* Role-based permission system
