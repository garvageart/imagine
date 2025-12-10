package routes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gorm.io/gorm"
)

const (
	DefaultTheme          = "viz-blue"
	ThemeStylePlaceholder = "%viz.css.theme_style%"
	ThemeAttrPlaceholder  = "%THEME_ATTR%"
)

type FrontendHandler struct {
	BuildPath string
	Logger    *slog.Logger
	DB        *gorm.DB
	// Cache index.html content
	indexContent []byte
	indexMutex   sync.RWMutex
}

func NewFrontendHandler(buildPath string, logger *slog.Logger, db *gorm.DB) *FrontendHandler {
	return &FrontendHandler{
		BuildPath: buildPath,
		Logger:    logger,
		DB:        db,
	}
}

func (h *FrontendHandler) getIndexContent() ([]byte, error) {
	h.indexMutex.RLock()
	if len(h.indexContent) > 0 {
		content := h.indexContent
		h.indexMutex.RUnlock()
		return content, nil
	}
	h.indexMutex.RUnlock()

	h.indexMutex.Lock()
	defer h.indexMutex.Unlock()

	// Double check
	if len(h.indexContent) > 0 {
		return h.indexContent, nil
	}

	content, err := os.ReadFile(filepath.Join(h.BuildPath, "index.html"))
	if err != nil {
		return nil, err
	}

	h.indexContent = content
	return content, nil
}

func (h *FrontendHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// Default to index.html for root
	if path == "/" {
		path = "index.html"
	}

	// Clean path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(h.BuildPath, cleanPath)

	// Check if file exists and is not a directory
	info, err := os.Stat(fullPath)
	isStaticFile := err == nil && !info.IsDir()

	// If it's a static file (and not index.html), serve it directly
	if isStaticFile && cleanPath != "index.html" && cleanPath != "." {
		http.ServeFile(w, r, fullPath)
		return
	}

	// Otherwise, serve index.html with theme injection (SPA catch-all)
	h.serveIndex(w, r)
}

func (h *FrontendHandler) serveIndex(w http.ResponseWriter, r *http.Request) {
	// Don't serve index.html for missing API routes
	if strings.HasPrefix(r.URL.Path, "/api") {
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	indexData, err := h.getIndexContent()
	if err != nil {
		if os.IsNotExist(err) {
			h.Logger.Debug("index.html not found (frontend build missing)", slog.String("path", h.BuildPath))
			http.Error(w, "Frontend build not found. If you are in development, ensure the Vite dev server is running and you are accessing the correct port (e.g. 7777).", http.StatusNotFound)
			return
		}

		h.Logger.Error("failed to read index.html", slog.Any("error", err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Theme logic
	cookie, err := r.Cookie("viz:theme")
	themeValue := DefaultTheme
	if err == nil {
		themeValue = cookie.Value
	}

	colorTheme := "viz-blue"
	modeTheme := "light"

	if strings.HasPrefix(themeValue, "viz-") {
		parts := strings.Split(themeValue, "-")
		if len(parts) >= 2 {
			colorTheme = fmt.Sprintf("%s-%s", parts[0], parts[1])
		}
		if len(parts) > 2 && (parts[2] == "light" || parts[2] == "dark") {
			modeTheme = parts[2]
		}
	} else if themeValue == "light" || themeValue == "dark" {
		modeTheme = themeValue
	}

	// Read CSS file
	cssPath := filepath.Join(h.BuildPath, "themes", colorTheme+".css")
	cssContent, err := os.ReadFile(cssPath)
	criticalCss := ""
	if err != nil {
		h.Logger.Warn("theme file not found", slog.String("theme", colorTheme), slog.String("path", cssPath))
	} else {
		criticalCss = fmt.Sprintf("<style id=\"generated-theme\">%s</style>", string(cssContent))
	}

	themeAttr := fmt.Sprintf("data-theme=\"%s\"", modeTheme)

	// Replace placeholders
	// Note: Doing string replacement on every request might be slow for high load,
	// but fine for this scale. For optimization, use bytes.Replace.
	responseHtml := bytes.Replace(indexData, []byte(ThemeStylePlaceholder), []byte(criticalCss), 1)
	responseHtml = bytes.Replace(responseHtml, []byte(ThemeAttrPlaceholder), []byte(themeAttr), 1)

	// Inject System Status
	status, err := GetSystemStatus(h.DB, h.Logger, r)
	configJson := "{}"
	if err == nil {
		if b, err := json.Marshal(&status); err == nil {
			configJson = string(b)
		}
	} else {
		h.Logger.Error("failed to get system status for injection", slog.Any("error", err))
	}

	configScript := fmt.Sprintf("<script>window.__IMAGINE_CONFIG__.system  = %s;</script>", configJson)
	// Inject before </head>. If </head> is missing (unlikely), it won't inject, which is acceptable fallback.
	responseHtml = bytes.Replace(responseHtml, []byte("</head>"), []byte(configScript+"</head>"), 1)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(responseHtml)
}
