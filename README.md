# Imagine 

**Imagine** is a self-hosted image management and processing platform designed for photographers, designers, and creative professionals. It provides a modern API-driven backend (Go) and a responsive web interface (SvelteKit) for organizing, searching, and sharing image collections.

![Home Page Screenshot](./docs/images/home_page_screenshot.png)

> **Note**: This is a work in progress and is not yet production-ready.

---

## Features

- **Image Upload & Organization**: Upload and automatically process images with thumbnails, EXIF extraction, and thumbhash generation.
- **Collections**: Group images into collections for better organization and curation.
- **Search & Sort**: Fast semantic search and flexible sorting (by date, name, size, etc.).
- **Auto-Rotation**: Automatically respects EXIF orientation tags using libvips for correct thumbnail display.
- **Job Queue**: Background processing with Watermill for thumbnail generation, image optimization, and metadata extraction.
- **Modern UI**: Built with SvelteKit 5 (runes-based reactivity), featuring drag-and-drop uploads, asset grids, and responsive design.
- **Multi-Storage Support**: Ready for local storage and cloud providers (GCP storage buckets planned).

---

## Architecture

### Backend (Go)
- **API Server** (`cmd/api`): RESTful API for image uploads, collections, and search operations.
- **Job Workers** (`internal/jobs/workers`): Background processing for thumbnail creation and metadata extraction.
- **Image Operations** (`internal/imageops`): libvips integration for high-performance image processing with EXIF-aware transformations.
- **Database**: PostgreSQL with GORM for image metadata, collections, and user data.

### Frontend (SvelteKit)
- **Location**: `viz/` directory
- **Tech Stack**: SvelteKit 5, TypeScript, Vite, SCSS
- **Features**: Asset grid view, upload panel with progress tracking, collections management, search interface.

### Configuration
All settings are managed via `imagine.json`:
- Server ports and database connections
- Upload directories and base storage paths
- GCP bucket names (for future cloud storage integration)
- Logging levels

---

## Project Status

> ⚠️ **Work in Progress**  
> Imagine is actively under development and not yet production-ready. Current limitations and planned improvements include:

### Known Limitations
- **Authentication**: No user authentication or authorization system yet (planned).
- **GCP Integration**: Cloud storage buckets are configured but not fully integrated.
- **Image Editing**: No in-browser editing capabilities (cropping, filters, etc.).
- **Batch Operations**: Limited bulk actions (delete multiple, move to collection, etc.).
- **API Documentation**: OpenAPI/Swagger spec is pending.
- **Deployment**: No Docker/Kubernetes setup or production deployment guide yet.

### Upcoming Features
- Multi-user support with role-based access control
- Image editing and transformation pipeline
- Advanced search with filters (date range, camera model, location)
- Album/collection sharing with public links
- CDN integration for optimized delivery
- Full GCP/AWS/S3 storage adapter
- Desktop/mobile apps (planned)

---

## Getting Started

### Prerequisites
- **Go** 1.23+ (with cgo support for libvips)
- **libvips** 8.10+ installed on your system
  - Windows: See `docs/Install Libvips Windows.md`
  - macOS: `brew install vips`
  - Linux: `apt install libvips-dev` (Debian/Ubuntu) or equivalent
- **PostgreSQL** 14+
- **Node.js** 20+ and **pnpm** (for the frontend)

### Installation

1. **Clone the repository**:
   ```bash
   git clone https://github.com/yourusername/imagine.git
   cd imagine
   ```

2. **Configure the database**:
   - Create a PostgreSQL database named `imagine`
   - Update `imagine.json` with your database credentials

3. **Set up the backend**:
   ```bash
   # Install Go dependencies
   go mod download
   
   # Build the API server
   go build -o bin/api ./cmd/api
   ```

4. **Set up the frontend**:
   ```bash
   cd viz
   pnpm install
   ```

5. **Run the development servers**:

   **Backend (API server)**:
   ```bash
   # From project root
   ./bin/api
   # Or with Go run
   go run ./cmd/api
   ```

   **Frontend (Vite dev server)**:
   ```bash
   # From viz/ directory
   pnpm dev
   ```

6. **Access the application**:
   - Frontend: `http://localhost:7777`
   - API: `http://localhost:7770`

---

## Configuration

Edit `imagine.json` to customize:
- **Server ports**: `servers.api-server.port`, `servers.viz.port`
- **Storage location**: `base_directory` and `upload.location`
- **Database**: `database.username`, `database.name`
- **GCP buckets**: `gcloud.storageBuckets` (for future cloud storage)
- **Logging**: `logging.level` (debug, info, warn, error)

---

## Development

### Project Structure
```
imagine/
├── cmd/
│   ├── api/          # API server entry point
│   └── desktop/      # (Planned) Desktop app
├── internal/
│   ├── entities/     # Database models (Image, Collection, User, etc.)
│   ├── imageops/     # Image processing (libvips wrappers, thumbnail generation)
│   ├── jobs/         # Job queue and worker management
│   ├── http/         # HTTP server utilities and middleware
│   └── db/           # Database connection and utilities
├── viz/              # SvelteKit frontend
│   ├── src/
│   │   ├── lib/      # Shared components, state, utilities
│   │   └── routes/   # SvelteKit routes (app pages)
│   └── static/       # Static assets (fonts, CSS)
├── var/
│   └── library/      # Default local image storage
└── imagine.json      # Main configuration file
```

### Running Tests
```bash
# Backend (Go)
go test ./...

# Frontend (Vitest + Playwright)
cd viz
pnpm test
```

---

## Contributing

Contributions are welcome! Since this project is still in early development, please open an issue first to discuss any major changes or features you'd like to add.

---

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).

See the `LICENSE` file for the full text of the license.

Copyright (c) 2025 Les

---

## Acknowledgments

- **libvips**: High-performance image processing library
- **SvelteKit**: Modern web framework for the frontend
- **Watermill**: Go library for message-driven architecture
- **GORM**: ORM for database interactions

---

**Questions or feedback?** Open an issue or reach out via the repository discussions.

> **Note:** This initial `README` was partially written by Co-Pilot