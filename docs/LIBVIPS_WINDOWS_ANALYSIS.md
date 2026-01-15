# Libvips Windows Binary Analysis (RAW Support)

## Overview
This document summarizes the investigation into RAW image support for `libvips` on Windows (native, non-WSL) as of January 2026.

## Resolution (2026-01-15)
**We successfully enabled native RAW support on Windows.**

### The Fix
1.  **Upgrade to libvips 8.18.0:**
    *   Previous versions (8.17.2) on Windows (via `build-win64-mxe` "all") did not include functional `libraw` support or lacked the `vips-raw` module in a way that `vips` could detect.
    *   Version **8.18.0** explicitly adds `dcrawload` (load via libraw) support which is included in the "all" distribution.
2.  **Regenerate Go Bindings:**
    *   We updated `vipsgen` and regenerated the Go bindings against the 8.18.0 headers.
    *   This exposed `NewDcrawloadBuffer` and other `dcraw*` functions.
3.  **Code Fallback:**
    *   Standard `vips.NewImageFromBuffer` (wrapping `vips_image_new_from_buffer`) does **not** automatically detect or use the `dcraw` loader for buffers on Windows in this build.
    *   We implemented a fallback in `internal/imageops/transforms.go`: if generic loading fails, we explicitly attempt `vips.NewDcrawloadBuffer`.

## Findings

### 1. Binary Distributions ("all" vs "web")
- **"all" distribution:** (v8.18.0) ~21 MB ZIP. Includes `libraw` linked internally (verified by `vips -l foreign` showing `VipsForeignLoadDcRaw`).
- **"web" distribution:** Smaller, lacks RAW support.

### 2. RAW Loader Status
- **Native libvips Support (v8.18+):**
  - `vips -l foreign` now lists `VipsForeignLoadDcRaw` (priority 100).
  - `vipsheader` works on RAW files.
  - Buffer loading requires explicit calls to `dcrawload_buffer`.

### 3. Verification
To verify support in a new environment:
```powershell
# 1. Run setup script
bun scripts/js/setup-libvips.ts

# 2. Check for dcraw loader
vips -l foreign | Select-String "DcRaw"
```

## Performance Note
Processing RAW files (especially large ones like 40MP+ RAF/CR3) is CPU intensive.
- **Canon R5 (45MP):** ~15s
- **Fujifilm XT-5 (40MP X-Trans):** ~4-5 minutes (in tests, likely single-threaded bottleneck or specific demosaicing complexity).

---
*Updated on 2026-01-15*