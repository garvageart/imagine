# Viewfinder

Viewfinder is the web-based graphical user interface for the Viz platform. It is a single-page application (SPA) built to provide a high-performance, responsive environment for managing large-scale image collections.

## Core Capabilities

-   **High-Performance Rendering**: Utilizes a WASM-powered justified layout engine (`@immich/justified-layout-wasm`) to efficiently render thousands of images in a seamless grid.
-   **Semantic Search**: deeply integrated with the Viz backend to support natural language queries, allowing users to find assets based on content description rather than just keywords.
-   **Metadata Editing**: Provides comprehensive read/write access to EXIF, IPTC, and XMP metadata standards. Support includes bulk editing and preset management.
-   **Real-time Synchronization**: Implements WebSocket connections to reflect backend state changes (e.g., job progress, file system updates) immediately in the UI.
-   **Asset Management**: Features drag-and-drop ingestion, nested collection organization, and album management.
-   **Theming**: Includes a robust theming system with support for light/dark modes and custom color palettes via SCSS variables.

## Technical Architecture

Viewfinder is built on a modern stack designed for reactivity and type safety.

### Stack

-   **Framework**: SvelteKit (Static Adapter)
-   **UI Library**: Svelte 5 (utilizing Runes for state management)
-   **Language**: TypeScript
-   **Styling**: SCSS with global variables and modular component styles
-   **API Integration**: Auto-generated TypeScript client based on OpenAPI 3.1 specifications

### Key Directories

-   `src/lib/api`: Generated API client and http interceptors.
-   `src/lib/states`: Global application state managed via Svelte 5 Runes (`.svelte.ts` files).
-   `src/lib/components`: Reusable UI components.
-   `src/lib/photo-layout`: Logic for the justified grid and virtual scrolling.
-   `src/routes`: SvelteKit file-based routing definitions.

## Development

### Prerequisites

-   **Node.js**: v24.13.0 or higher (recommended)
-   **pnpm**: v8 or higher

### Environment Setup

1.  **Install Dependencies**:
    ```bash
    pnpm install
    ```

2.  **Backend Connection**:
    The development server proxies requests to the backend. Ensure the Viz API is running locally on port `7770`.
    
    If the API is hosted elsewhere, configure the proxy settings in `vite.config.ts`.

### Scripts

-   **`pnpm dev`**: Starts the Vite development server. Accessible at `http://localhost:7777`.
-   **`pnpm build`**: Compiles the application for production using `@sveltejs/adapter-static`. Output is located in the `build/` directory.
-   **`pnpm check`**: Runs `svelte-check` to validate TypeScript and Svelte syntax.
-   **`pnpm test:unit`**: Executes unit tests via Vitest.
-   **`pnpm test:e2e`**: Executes end-to-end tests via Playwright.
-   **`pnpm gen:api`**: Regenerates the API client from the backend OpenAPI specification.

## License

This project is licensed under the AGPL-3.0. See the root `LICENSE` file for details.
