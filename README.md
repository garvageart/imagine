# Viz

<a href="https://opensource.org/license/agpl-v3"><img src="https://img.shields.io/badge/License-AGPL_v3-blue.svg?color=0F1726&style=for-the-badge&label=License&logoColor=000000&labelColor=eee" alt="License: AGPLv3"></a>

**Viz** is a self-hosted image management and processing platform designed for photographers and media teams. It provides a modern API-driven backend (Go) and a web interface (SvelteKit) for organizing, searching, and sharing image collection

![Home Page Screenshot](./docs/assets/collection_images_screenshot.png)

> **Work in Progress** 🚧
> 
> This project is in active development. Features and APIs may change frequently. Feedback and contributions are welcome!

---

## Features

- **Image Upload & Organization**: Upload and automatically process images with thumbnails, EXIF extraction.
- **Collections**: Group images into collections for better organization.
- **Search**: Fast semantic search.
- **Background Processing**: Robust job queue (Watermill + Redis) for non-blocking image operations.
- **Modern UI**: Built with SvelteKit 5, featuring a responsive image grid, metadata view and editing, and drag-and-drop uploads.
- **Deployment**: Docker Compose support for easy set up (API, Frontend, Postgres, Redis).

---

## Quick Start (Docker)

Get started quickly using [Docker Compose](https://docs.docker.com/compose/).

#### Clone & Configure:
```bash
git clone https://github.com/garvageart/viz.git
cd viz

# Configure environment variables
cp .env.example .env
```

#### Run:
```bash
docker compose up --build -d
```

#### Use:
    - Frontend: `http://localhost:7777`
    - API: `http://localhost:7770`

See [**docs/BUILDING.md**](./docs/BUILDING.md) for detailed setup instructions, including **Manual/Non-Docker** development guides (Windows/Linux/macOS).

---

## Architecture

### Backend (Go 1.25)
- **Framework**: go-chi Router
- **Database**: PostgreSQL (via GORM)
- **Queue**: In-Memory or Redis (via Watermill)
- **Search**: PostgreSQL Full-Text Search
- **Image Processing**: libvips

### Frontend (Viewfinder)
- **Framework**: Svelte 5
- **Styling**: SCSS
- **Icons**: Material Design

### Monorepo Structure
- **Go Workspace**: Managed via `go.work` for the backend.
- **PNPM Workspaces**: Frontend (`viz/`) and internal tools (`scripts/js/`) are managed as pnpm workspaces.
---

## License

This project is licensed under the GNU Affero General Public License v3.0 (AGPL-3.0).
See `LICENSE` for details.

---

## Questions or feedback?
Open an issue or reach out via the repository discussions.

Copyright (c) 2025 Les
