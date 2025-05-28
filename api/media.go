package main

***REMOVED***
***REMOVED***
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"

***REMOVED***

	libhttp "imagine/common/http"
***REMOVED***

func setupImageRouter(***REMOVED*** *chi.Mux {
	imageRouter := chi.NewRouter(***REMOVED***

	imageRouter.Get("/download", func(res http.ResponseWriter, req *http.Request***REMOVED*** {
		panic("not implemented"***REMOVED***
***REMOVED******REMOVED***

	imageRouter.Get("/upload", func(res http.ResponseWriter, req *http.Request***REMOVED*** {
		panic("not implemented"***REMOVED***
***REMOVED******REMOVED***

	return imageRouter
***REMOVED***

func main(***REMOVED*** {
	router := chi.NewRouter(***REMOVED***
	imageRouter := setupImageRouter(***REMOVED***

	logger := libhttp.SetupChiLogger(***REMOVED***
	correctLogger := slog.NewLogLogger(logger.Handler(***REMOVED***, slog.LevelDebug***REMOVED***

	router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: correctLogger,
***REMOVED******REMOVED******REMOVED***

	router.Use(middleware.AllowContentEncoding("deflate", "gzip"***REMOVED******REMOVED***
	router.Use(middleware.RequestID***REMOVED***

	router.Get("/ping", func(res http.ResponseWriter, req *http.Request***REMOVED*** {
		jsonResponse := map[string]any{"message": "pong"***REMOVED***
		render.JSON(res, req, jsonResponse***REMOVED***
***REMOVED******REMOVED***

	router.Mount("/image", imageRouter***REMOVED***

	logger.Info("Starting Server on :8080"***REMOVED***
	err := http.ListenAndServe(":8080", router***REMOVED***

***REMOVED***
		logger.Error(fmt.Sprintf("failed to start server: %s", err***REMOVED******REMOVED***
***REMOVED***
***REMOVED***
