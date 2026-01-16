# Installing libvips

Imagine requires **libvips** to be installed on your system. We rely on `pkg-config` to locate the library, ensuring a consistent build process across all operating systems.

We provide a cross-platform setup script to automate this process.

## Automated Installation (Recommended)

**Prerequisites:**
- **Node.js** (v24+) or **Bun** installed.
- **Windows:** [MSYS2](https://www.msys2.org/) must be installed.

Run the setup script from the project root:

**Windows (MSYS2 MinGW 64-bit Terminal):**
```bash
# Ensure you are in the MSYS2 MinGW 64-bit terminal
bun scripts/js/setup-libvips.ts
# OR
npx tsx scripts/js/setup-libvips.ts
```

**macOS / Linux (Standard Terminal):**
```bash
bun scripts/js/setup-libvips.ts
# OR
npx tsx scripts/js/setup-libvips.ts
```

This script will:
1.  **Windows:** Use `pacman` within your MSYS2 environment to install `mingw-w64-x86_64-vips`. It will also configure your user environment variables (`PATH`, `PKG_CONFIG_PATH`) to point to the MSYS2 MinGW directory.
2.  **macOS:** Install `vips` and `pkg-config` via Homebrew.
3.  **Linux:** Attempt to install `libvips-dev` using your package manager (`apt`, `dnf`, `pacman`).

## Manual Installation

If you prefer to install manually:

### Windows
1.  Install [MSYS2](https://www.msys2.org/).
2.  Open the **MSYS2 MinGW 64-bit** terminal.
3.  Run: `pacman -Syu && pacman -S mingw-w64-x86_64-vips`
4.  Add `C:\msys64\mingw64\bin` to your **User PATH**.
5.  Set the `PKG_CONFIG_PATH` environment variable to `C:\msys64\mingw64\lib\pkgconfig`.

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
