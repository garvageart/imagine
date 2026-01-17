# Infrastructure & Building Guide

This guide covers how to build, run, and develop **Imagine**.

**Preferred Method**: Docker Compose (simplest, consistent environment).
**Manual Method**: For specialized development or environments where Docker is not available.

---

## 1. Quick Start (Docker)

This is the recommended way to run the application. It handles the database, redis, and all dependencies automatically.

### Prerequisites
- **Docker Engine** & **Docker Compose**
- **Git**

### Usage

1.  **Clone the repository**:
```bash
git clone https://github.com/garvageart/imagine.git
cd imagine
```

2.  **Environment Setup**:
Copy the example configuration:
```bash
cp .env.example .env
```
*Modify `.env` to change ports or database credentials.*

3.  **Start Services**:
```bash
docker compose up --build -d
```
This brings up:
- **`postgres`**: Database (Port 5432)
- **`redis`**: Job Queue Broker (Port 6379)
- **`api`**: Go Backend (Port 7770)
- **`viz`**: SvelteKit Frontend (Port 7777 - *Note: The API also serves the frontend in production builds, but the dev container runs `pnpm dev` for hot-reloading.*)

4.  **Access**:
    - Frontend: `http://localhost:7777`
    - API: `http://localhost:7770`

### Database Note
The Docker setup uses `docker/initdb/01-create-superuser.sh` to automatically create the Postgres user and database defined in your `.env` file on the first run.

---

## 2. Manual Development (Non-Docker)

**⚠️ Windows Users**: The project `Makefile` is written for Bash. Use **WSL2** (Windows Subsystem for Linux) or **Git Bash**

### Prerequisites
- **Go**: v1.25+ (Required for workspace support)
- **Node.js**: v24+
- **pnpm**: Latest
- **PostgreSQL**: v14+
- **Redis**: v7+
- **libvips**: v8.18+
  - **All Platforms**: Run `bun scripts/js/setup-libvips.ts` (Recommended)
  - **Manual**: See `docs/INSTALL_LIBVIPS.md`

### Step 1: Infrastructure Setup

Since you are not using Docker, you must set up the database and redis manually.

1.  **Start Redis**: Ensure Redis is running on port `6379`.
2.  **Start PostgreSQL**: Ensure Postgres is running.
3.  **Create Database & User**:
You need to manually create the role and database that the app expects. Connect to your local Postgres (`psql postgres`) and run:

```sql
-- Replace 'myuser' and 'mypassword' with values from your .env
CREATE ROLE myuser WITH LOGIN SUPERUSER PASSWORD '<mypassword>';
CREATE DATABASE imagine OWNER myuser;
```

> Note: The application will handle table creation (AutoMigrate) on startup.

### Step 2: Backend (API)

1.  Ensure `.env` exists and points to your local DB/Redis.
2.  Install dependencies:
```bash
go mod download
```
3.  Run the API:
```bash
# Run from the project root using the workspace
go run ./cmd/api
```
The server should start on port `7770` (or as defined in `imagine.json` / `.env`).

### Step 3: Frontend (Viz)

1.  **Install dependencies** (from the project root):
```bash
pnpm install
```
2.  **Start the development server**:
```bash
pnpm dev
```
Access at `http://localhost:7777`.

---

## 3. Manual Production Build

In production, the Go backend can serve the compiled frontend assets, deploying as a single binary.

1.  **Build Frontend** (from the project root):
```bash
pnpm build
```
This generates the single-page application JS file in `./build/viz`.

2.  **Build Backend**:
From the project root:
```bash
go build -o bin/api ./cmd/api
```

3.  **Run**:
    Set the `IMAGINE_FRONTEND_BUILD_PATH` environment variable to point to the built assets.
    
```bash
export IMAGINE_FRONTEND_BUILD_PATH="./build/viz"
./bin/api
```
    
Accessing `http://localhost:7770` (API port) will serve the frontend app for any non-API routes.

---

## Troubleshooting

- **`connect: connection refused`**: Check if Redis and Postgres are running.
- **Go Version Errors**: Ensure you are using Go 1.25. The project uses a Go Workspace (`go.work`).
- **"libvips not found"**: Ensure `pkg-config` can find libvips. On Windows, check your `PATH` and `PKG_CONFIG_PATH`.
