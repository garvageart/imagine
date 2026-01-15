# Installing libvips

Imagine requires **libvips** to be installed on your system. We rely on `pkg-config` to locate the library, ensuring a consistent build process across all operating systems.

We provide a cross-platform setup script to automate this process.

## Automated Installation (Recommended)

**Prerequisites:**
- **Node.js** (v18+) or **Bun** installed.

Run the setup script from the project root:

**Using Bun:**
```bash
bun scripts/js/setup-libvips.ts
```

**Using Node:**
```bash
# You may need to compile it or run with ts-node/tsx if not using Bun
npx tsx scripts/js/setup-libvips.ts
```

This script will:
1.  **Windows:** Download the correct pre-compiled binaries (with RAW support), install them to `%LOCALAPPDATA%\Programs\vips`, and configure your user environment variables (`PATH`, `PKG_CONFIG_PATH`).
2.  **macOS:** Install `vips` and `pkg-config` via Homebrew.
3.  **Linux:** Attempt to install `libvips-dev` using your package manager (`apt`, `dnf`, `pacman`).

## Manual Installation

If you prefer to install manually:

### Windows
1.  Download `vips-dev-w64-all-x.y.z.zip` (matching `.libvips-version`) from [libvips releases](https://github.com/libvips/build-win64-mxe/releases).
2.  Extract to a permanent location (e.g., `C:\vips`).
3.  Add `C:\vips\bin` to your `PATH`.
4.  Set `PKG_CONFIG_PATH` to `C:\vips\lib\pkgconfig`.

### macOS
```bash
brew install vips pkg-config
```

### Linux (Ubuntu/Debian)
```bash
sudo apt-get install libvips-dev pkg-config
```

## Verification

After installation, verify that `pkg-config` can see `libvips` and that it has the expected features.

```bash
vips --version
pkg-config --cflags --libs vips
```