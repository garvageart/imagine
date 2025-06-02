package main

***REMOVED***
***REMOVED***
***REMOVED***
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"


***REMOVED***

	gcp "imagine/common/gcp/storage"
	libhttp "imagine/common/http"
***REMOVED***

type ImagineMediaServer struct {
	*libhttp.ImagineServer
***REMOVED***

func (server ImagineMediaServer***REMOVED*** setupImageRouter(***REMOVED*** *chi.Mux {
	imageRouter := chi.NewRouter(***REMOVED***
	logger := server.Logger

	gcsContext, gcsContextCancel := context.WithCancel(context.Background(***REMOVED******REMOVED***
	defer gcsContextCancel(***REMOVED***

	storageClient, err := gcp.SetupClient(gcsContext***REMOVED***
***REMOVED***
		panic("Failed to setup GCP Storage client" + err.Error(***REMOVED******REMOVED***
***REMOVED***

	imageRouter.Get("/download", func(res http.ResponseWriter, req *http.Request***REMOVED*** {
		res.WriteHeader(http.StatusNotImplemented***REMOVED***
		res.Header(***REMOVED***.Add("Content-Type", "text/plain"***REMOVED***
		res.Write([]byte("not implemented"***REMOVED******REMOVED***
***REMOVED******REMOVED***
	
	imageRouter.Get("/upload", func(res http.ResponseWriter, req *http.Request***REMOVED*** {
		res.WriteHeader(http.StatusNotImplemented***REMOVED***
		res.Header(***REMOVED***.Add("Content-Type", "text/plain"***REMOVED***
		res.Write([]byte("not implemented"***REMOVED******REMOVED***

***REMOVED******REMOVED***

	return imageRouter
***REMOVED***

func (server ImagineMediaServer***REMOVED*** Launch(router *chi.Mux***REMOVED*** {
	imageRouter := server.setupImageRouter(***REMOVED***
	logger := server.Logger

	correctLogger := slog.NewLogLogger(logger.Handler(***REMOVED***, slog.LevelDebug***REMOVED***

	router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: correctLogger,
***REMOVED******REMOVED******REMOVED***

	router.Use(middleware.AllowContentEncoding("deflate", "gzip"***REMOVED******REMOVED***
	router.Use(middleware.RequestID***REMOVED***

	// Mount image router to main router
	router.Mount("/image", imageRouter***REMOVED***

	router.Get("/ping", func(res http.ResponseWriter, req *http.Request***REMOVED*** {
		jsonResponse := map[string]any{"message": "pong"***REMOVED***
		render.JSON(res, req, jsonResponse***REMOVED***
***REMOVED******REMOVED***

	logger.Info(fmt.Sprint("Starting server on port ", server.Port***REMOVED******REMOVED***
	err := http.ListenAndServe(server.Host+":"+fmt.Sprint(server.Port***REMOVED***, router***REMOVED***

***REMOVED***
		logger.Error(fmt.Sprintf("failed to start server: %s", err***REMOVED******REMOVED***
***REMOVED***
***REMOVED***

func main(***REMOVED*** {
	key := "media-server"
	router := chi.NewRouter(***REMOVED***
	logger := libhttp.SetupChiLogger(key***REMOVED***

	var server = &ImagineMediaServer{
		ImagineServer: &libhttp.ImagineServer{
			Host:   "localhost",
			Key:    key,
			Logger: logger,
***REMOVED***
***REMOVED***

	config, err := server.ReadConfig(***REMOVED***
***REMOVED***
		panic("Unable to read config file"***REMOVED***
***REMOVED***

	server.Port = config.GetInt(fmt.Sprintf("servers.%s.port", server.Key***REMOVED******REMOVED***
	server.Launch(router***REMOVED***
***REMOVED***