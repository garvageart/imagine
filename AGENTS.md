# Project Context for Agents

This document provides a comprehensive guide for AI agents and developers working on the **Imagine** project. Adhering to these guidelines is crucial for maintaining code quality and consistency.

## 1. Project Overview
**Imagine** is a high-performance, self-hosted image management and processing platform. It is designed as a modern, open-source alternative to commercial products like PhotoShelter, tailored for professional photographers, media teams, and content creators who need full control over their digital assets.

**Core Features**:
- **High-Performance Asset Ingestion**: Fast uploads with checksum verification to prevent duplicates.
- **Asynchronous Processing**: Background jobs handle thumbnail generation, metadata extraction, and other long-running tasks without blocking the UI.
- **Non-Destructive Transforms**: Original images are never altered. All transformations (resizing, cropping, format changes) are generated on-the-fly and cached.
- **Powerful Search**: Flexible search capabilities based on metadata, tags, and other attributes.
- **Rich Metadata Management**: Reads, writes, and preserves EXIF, IPTC, and XMP metadata.
- **Flexible Organization**: Group images into logical **Collections**.

## 2. Technical Stack

### Backend (`/` root)
- **Language**: Go 1.25+
- **Web Framework**: `go-chi` (v5) is used for routing. It's lightweight, idiomatic, and avoids reflection.
- **Database**: PostgreSQL (via Docker in development).
- **ORM**: GORM is used for database interactions. Database logic is primarily located in `internal/db/`.
- **Image Processing**: `libvips` (v8.18+) provides fast and memory-efficient image processing. It's accessed via Go bindings. Core logic is in `internal/imageops/`.
- **Queue/Async**: Watermill provides a robust pub/sub system for background jobs (e.g., thumbnailing, metadata extraction). It's configured to use Redis in production and in-memory for simple setups.
- **Logging**: `slog` (standard library structured logging) is used throughout. Logs are written as JSON for machine-parsability.
- **Configuration**: Viper manages configuration from `imagine.json` and environment variables. `internal/config/` holds the structures.
- **Authentication**: Supports both stateful sessions (cookies) for the frontend and stateless API Keys for third-party integrations. Logic is in `internal/auth/`.
- **API Documentation**: An OpenAPI 3.0 specification is maintained at `api/openapi/openapi.yaml`. This spec is the source of truth for generating both Go DTOs and the TypeScript API client.

### Frontend (`/viz` directory)
- **Framework**: SvelteKit, configured as a pure **Single-Page Application (SPA)** using `@sveltejs/adapter-static`.
- **Language**: TypeScript is used for all Svelte components and modules, ensuring type safety.
- **Reactivity Model**: Embrace Svelte 5 Runes for all state management (`$state`, `$derived`, `$effect`). Avoid legacy stores (`writable`, `readable`). `$props` are used for component inputs.
- **Styling**: SCSS (Sass) is the preprocessor of choice. Write component-scoped styles in `<style lang="scss">` blocks. Global styles, variables, and mixins are located in `viz/src/lib/styles/`.
- **API Client**: The primary client is generated from the OpenAPI spec using `oazapfts`. A custom wrapper in `viz/src/lib/api/index.ts` enhances this client, primarily for handling binary file uploads with the correct `Content-Type`.
- **State Management**: Global and session-level state is managed through custom modules in `viz/src/lib/states/`. These modules are simple `.ts` files that export state runes.
- **Testing**:
  - **Vitest**: For unit and component testing.
  - **Playwright**: For end-to-end (E2E) testing.

## 3. Key Workflows & Logic

### Code Generation: The Single Source of Truth
The project relies heavily on code generation from the OpenAPI specification (`api/openapi/openapi.yaml`). This file is the **single source of truth** for API data structures.

The `make generate-types` command orchestrates this entire process:
1.  **Go DTOs**: It uses `oapi-codegen` to generate Go data transfer objects (structs used in API requests/responses) into `internal/dto/types.gen.go`.
2.  **GORM Entities**: It runs a **custom tool** (`tools/genentities`) that reads the same OpenAPI spec. Schemas marked with the `x-entity: true` extension are converted into GORM database models in `internal/entities/generated.go`. This tool also generates helper methods (`DTO()` and `FromDTO()`) to convert between the entity and the DTO.
3.  **TypeScript Client**: It uses `oazapfts` to generate the frontend's TypeScript API client in `viz/src/lib/api/client.gen.ts`.

### API Endpoint Creation Workflow
1.  **Define Schema in OpenAPI**: All new data structures (for both API transfer and database storage) **must** be defined under `components/schemas` in `api/openapi/openapi.yaml`.
2.  **Add `x-` extensions**:
    - If the schema represents a database table, add the extension `x-entity: true`.
    - To define database indexes, add `x-go-gorm-index` with the required fields.
3.  **Define Endpoint**: Add the new API endpoint's path, method, parameters, and request/response schemas to the `paths` section of the OpenAPI spec.
4.  **Generate All Code**: Run `make generate-types`. This is a critical step that updates DTOs, GORM entities, and the TS client.
5.  **Create/Update Database Migration**: If adding a new table or column, create a new migration file to apply the schema change to the database.
6.  **Implement Handler**: Write the `http.HandlerFunc` logic for the new route in the appropriate file under `cmd/api/routes/`. Use the newly generated GORM entities for database logic and the DTOs for handling API requests and responses.

### Image Upload & Processing
1.  **Upload**: A client POSTs an image to `/api/images`. The API server saves the original file to a new directory named by a generated UID (e.g., `/data/images/<UID>/<original_filename>`).
2.  **Enqueue Job**: An `ImageProcessJob` is enqueued into the Watermill queue.
3.  **Worker Processing**: A background worker picks up the job.
4.  **Transforms & Metadata**: The worker generates permanent transforms (thumbnail, preview), extracts EXIF/XMP metadata, and creates a Thumbhash.
5.  **Database Update**: The worker updates the corresponding `images` entity in the database, marking it as `processed` and saving the new metadata.
6.  **WebSocket Notification**: The worker broadcasts a `job-completed` event via WebSocket to inform the frontend.

## 4. Directory Structure
- `cmd/api/`: Main application entry point (`api.go`) and route definitions (`routes/`).
- `internal/`: Core application logic, not to be imported by other applications.
  - `auth/`: Authentication/authorization logic (API keys, sessions, scopes).
  - `config/`: Configuration structs and loading logic (Viper).
  - `db/`: Database connection and generic operations.
  - `dto/`: **Generated** Data Transfer Objects. Do not edit `types.gen.go` manually.
  - `entities/`: **Generated** GORM models. Do not edit `generated.go` manually.
  - `http/`: Core HTTP server logic, middleware, and WebSocket broker.
  - `imageops/`: `libvips` wrappers for specific image operations.
  - `images/`: High-level image management (caching, storage, etc.).
  - `jobs/`: Core job queue and worker infrastructure.
    - `workers/`: Implementations of specific background jobs.
  - `transform/`: Data structures and pure functions for image transforms.
- `tools/genentities`: Custom tool to generate GORM entities from the OpenAPI spec.
- `viz/`: The SvelteKit frontend application (SPA).
  - `src/lib/api/`: **Generated** API client. Do not edit `client.gen.ts` manually.
  - `src/lib/components/`: Reusable Svelte components.
  - `src/routes/`: Client-side pages and routes. Does **not** contain backend logic.
- `docs/`: Developer, user, and architecture documentation.
- `scripts/js/`: Node.js-based scripts for maintenance and code generation orchestration.

## 5. Project Management & Build
- **Makefile**: The single source of truth for running project tasks.
  - `make build`: Compiles the Go backend and builds the SvelteKit frontend for production.
  - `make dev`: Provides instructions for running the separate backend and frontend development servers.
  - `make docker-up`: The recommended way to start the full stack (API, Viz, Postgres, Redis) for development.
  - `make generate-types`: **Crucial command**. Regenerates Go DTOs, GORM Entities, and the TypeScript API client from the OpenAPI spec.
- **Docker**:
  - `docker-compose.yml`: Orchestrates all services for a complete production-like environment.
  - `Dockerfile.api`: Builds the Go backend binary, which also serves the static frontend assets.

## 6. Coding Guidelines

### Go (Backend)
- **Formatting**: Code MUST be formatted with `gofmt`.
- **If-Statements**: ALWAYS use multi-line blocks.
- **Error Handling**: Handle all errors explicitly.
- **Logging**: Use the structured logger (`slog`).

### TypeScript / Svelte (Frontend)
- **Props**: Define component props using a `Props` interface in the `<script>` block.
- **API Data Loading**: Data for pages should be fetched within SvelteKit's universal `load` function in `+page.ts` or `+layout.ts` files. Since the project is a pure SPA, this function will always run on the client-side. This is the preferred way to fetch data before a page renders, as it automatically handles both initial loads and client-side navigations. Do **not** use `+page.server.ts` files.
- **Styling**: Prefer component-scoped SCSS.

## 7. Setup & Environment
- **Primary Method**: Use Docker via `make docker-up`.
- **Manual Setup**:
  - Run `bun scripts/js/setup-libvips.ts` to install `libvips` (v8.18+) and configure environment variables.
- **Ports**: API: `7770`, Frontend Dev: `7777`, PostgreSQL: `5432`, Redis: `6379`.
- **Configuration**: The default configuration is in `imagine.json`. Overrides can be placed in `imagine.local.json` or set via environment variables.