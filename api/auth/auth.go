package main

***REMOVED***
***REMOVED***
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
***REMOVED***
	"log/slog"
	"net/http"
***REMOVED***
	"strings"
***REMOVED***

	"github.com/dromara/carbon/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/oauth2"

	"github.com/go-chi/render"
***REMOVED***

	"imagine/common/auth"
	oauth "imagine/common/auth/oauth"
	"imagine/common/crypto"
	libhttp "imagine/common/http"
	"imagine/db"
***REMOVED***
***REMOVED***

const (
	serverKey = "auth-server"
***REMOVED***

type ImagineAuthServer struct {
	*libhttp.ImagineServer
	oauth.ImagineOAuth
***REMOVED***

type ImagineAuthCodeFlow struct {
	Code  string `json:"code"`
	State string `json:"state"`
***REMOVED***

type ImagineAuthPasswordFlow struct {
	Ctate string
***REMOVED***

func (server ImagineAuthServer***REMOVED*** Launch(router *chi.Mux***REMOVED*** {
	// Setup logger
	logger := server.Logger
	correctLogger := slog.NewLogLogger(logger.Handler(***REMOVED***, slog.LevelDebug***REMOVED***
	router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: correctLogger,
***REMOVED******REMOVED******REMOVED***

	// Setup general middleware
	router.Use(middleware.AllowContentEncoding("deflate", "gzip"***REMOVED******REMOVED***
	router.Use(middleware.RequestID***REMOVED***
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://localhost:7777", "http://localhost:7777"***REMOVED***,
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"***REMOVED***,
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "set-cookie"***REMOVED***,
		AllowCredentials: true,
***REMOVED******REMOVED******REMOVED***

	// Set up auth
	jwtSecret := os.Getenv("JWT_SECRET"***REMOVED***
	jwtSecretBytes := []byte(jwtSecret***REMOVED***

	// Setup DB
	mongoCtx, cancelMongo := context.WithTimeout(context.Background(***REMOVED***, 60*time.Second***REMOVED***
	defer cancelMongo(***REMOVED***

	var database db.DBClient = db.DB{
***REMOVED***
***REMOVED***
***REMOVED***
		User:         os.Getenv("MONGO_USER"***REMOVED***,
		Password:     os.Getenv("MONGO_PASSWORD"***REMOVED***,
		DatabaseName: "ImagineDev",
		Collection:   "auth",
***REMOVED***
***REMOVED***

	client, mongoErr := database.Connect(***REMOVED***
	defer func(***REMOVED*** {
		database.Disconnect(client***REMOVED***
		fmt.Println("Disconnected from MongoDB"***REMOVED***
***REMOVED***(***REMOVED***

	if mongoErr != nil {
		panic("error connecting mongo db"***REMOVED***
***REMOVED***

	router.Get("/ping", func(res http.ResponseWriter, req *http.Request***REMOVED*** {
		jsonResponse := map[string]any{"message": "You have been PONGED by the auth server. What do you want here? 🤨"***REMOVED***
		render.JSON(res, req, jsonResponse***REMOVED***
***REMOVED******REMOVED***

	router.Get("/oauth", func(res http.ResponseWriter, req *http.Request***REMOVED*** {
		var oauthConfig *oauth2.Config
		provider := req.FormValue("provider"***REMOVED***

		switch provider {
		case "google":
			oauthConfig = oauth.GoogleOAuthConfig
		case "github":
			oauthConfig = oauth.GithubOAuthConfig
		default:
			var providerErr = errors.New("unsupported provider"***REMOVED***
			if provider != "" {
				providerErr = errors.New("no provider... provided"***REMOVED***
		***REMOVED***

			libhttp.ServerError(res, req, providerErr, logger, nil,
				"",
				"Error siging you in. Please try again later.",
			***REMOVED***
	***REMOVED***

		state, err := gonanoid.New(24***REMOVED***
	***REMOVED***
			libhttp.ServerError(res, req, err, logger, nil,
				"error generating oauth state",
				"",
			***REMOVED***

			return
	***REMOVED***

		stateHash := crypto.CreateHash([]byte(state***REMOVED******REMOVED***
		encryptedState := base64.URLEncoding.EncodeToString(stateHash***REMOVED***

		http.SetCookie(res, &http.Cookie{
			Name:     "img-state",
			Value:    encryptedState,
			Expires:  carbon.Now(***REMOVED***.AddYear(***REMOVED***.StdTime(***REMOVED***,
			Path:     "/",
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
	***REMOVED******REMOVED***

		oauthUrl, err := oauth.SetupOAuthURL(res, req, oauthConfig, provider, state***REMOVED***

	***REMOVED***
			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"",
			***REMOVED***
			return
	***REMOVED***

		http.Redirect(res, req, oauthUrl, http.StatusTemporaryRedirect***REMOVED***
***REMOVED******REMOVED***

	router.Post("/oauth/{provider***REMOVED***", func(res http.ResponseWriter, req *http.Request***REMOVED*** {
		provider := strings.ToLower(chi.URLParam(req, "provider"***REMOVED******REMOVED***
		var actualUserData any // To store the user data struct
		var userEmail string

		switch provider {
		case "google":
			resp, err := server.ImagineOAuth.GoogleOAuthHandler(res, req, logger***REMOVED***

		***REMOVED***
				libhttp.ServerError(res, req, err, logger, nil,
					"Error getting user data from Google",
					"We encountered an issue while trying to fetch your Google profile. Please try again."***REMOVED***
				return
		***REMOVED***

			userEmail = resp.Email
			actualUserData = resp
		case "github":
			resp, err := server.ImagineOAuth.GithubOAuthHandler(res, req, logger***REMOVED***

		***REMOVED***
				libhttp.ServerError(res, req, err, logger, nil,
					"Error getting user data from Google",
					"We encountered an issue while trying to fetch your Google profile. Please try again."***REMOVED***
				return
		***REMOVED***

			userEmail = resp.GetEmail(***REMOVED***
			actualUserData = resp
		default:
			res.WriteHeader(http.StatusBadRequest***REMOVED***
			res.Write([]byte("OAuth provider unsupported"***REMOVED******REMOVED***
			http.Redirect(res, req, "localhost:7777/", http.StatusTemporaryRedirect***REMOVED***
	***REMOVED***

		userFingerprint := auth.GenerateRandomBytes(50***REMOVED***
		userFingerprintString := hex.EncodeToString(userFingerprint***REMOVED***

		tokenSHA256 := sha256.New(***REMOVED***
		tokenSHA256.Write([]byte(userFingerprintString***REMOVED******REMOVED***
		sha256Sum := tokenSHA256.Sum(nil***REMOVED***
		hashString := hex.EncodeToString(sha256Sum***REMOVED***

		expiryTime := carbon.Now(***REMOVED***.AddYear(***REMOVED***.StdTime(***REMOVED***

		// Create a new JWT token with claims
		claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": userEmail,         // Subject (user identifier***REMOVED***
			"iss": serverKey,         // Issuer
			"iat": time.Now(***REMOVED***.Unix(***REMOVED***, // Issued at
			"exp": expiryTime,        // Expiration time
			"fgp": hashString,        // Fingerprint
	***REMOVED******REMOVED***

		tokenString, err := claims.SignedString(jwtSecretBytes***REMOVED***
	***REMOVED***
			res.Write([]byte("Error creating JWT token:" + err.Error(***REMOVED******REMOVED******REMOVED***
	***REMOVED***

		jsonBytes, err := json.Marshal(actualUserData***REMOVED***
	***REMOVED***
			libhttp.ServerError(res, req, err, logger, nil, "Failed to marshal user data to JSON", "Could not process your login information."***REMOVED***
			return
	***REMOVED***

		http.SetCookie(res, &http.Cookie{
			Name:     "img-auth_token",
			Value:    tokenString,
			Expires:  expiryTime,
			HttpOnly: true,
			Path:     "/",
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
	***REMOVED******REMOVED***
		res.Header(***REMOVED***.Add("Content-Type", "application/json"***REMOVED***
		res.WriteHeader(http.StatusOK***REMOVED***
		res.Write(jsonBytes***REMOVED***
***REMOVED******REMOVED***

	router.Put("/user/create", func(res http.ResponseWriter, req *http.Request***REMOVED*** {
		oauthRedirect := req.FormValue("oauth_redirect"***REMOVED***
		continueUrl := req.FormValue("continue"***REMOVED***
		oauthState := req.CookiesNamed("img-state"***REMOVED***[0]
		email := req.FormValue("email"***REMOVED***
		name := req.FormValue("name"***REMOVED***
		userUUID := uuid.New(***REMOVED***
		userIdBytes, err := userUUID.MarshalBinary(***REMOVED***
		usedOAuth := oauthRedirect != ""

	***REMOVED***
			libhttp.ServerError(res, req, err, logger, nil,
				"error marshaling uuid for user id",
				"Error creating your account, please try again later",
			***REMOVED***
	***REMOVED***

		userId := hex.EncodeToString(userIdBytes***REMOVED***

		userStruct := &ImagineUser{
			UUID:          userUUID.String(***REMOVED***,
			ID:            userId,
			Email:         email,
			Username:      name,
			CreatedAt:     carbon.Now(***REMOVED***.String(***REMOVED***,
			UsedOAuth:     usedOAuth,
			OAuthState:    oauthState.Value,
			OAuthProvider: oauthRedirect,
	***REMOVED***

		userDocument, err := db.ToBSONDocument(userStruct***REMOVED***
	***REMOVED***
			libhttp.ServerError(res, req, err, logger, nil,
				"error marshalling user document",
				"Error creating your account, please try again later",
			***REMOVED***
	***REMOVED***

		_, err = database.Insert(*userDocument***REMOVED***
	***REMOVED***
			libhttp.ServerError(res, req, err, logger, nil,
				"err creating user account on database",
				"Error creating your account, please try again later",
			***REMOVED***
	***REMOVED***

		time.Sleep(2000***REMOVED***
		if continueUrl != "" {
			continueUrl = "/"
	***REMOVED***

		http.Redirect(res, req, continueUrl, http.StatusTemporaryRedirect***REMOVED***
***REMOVED******REMOVED***

	address := fmt.Sprintf("%s:%d", server.Host, server.Port***REMOVED***

	logger.Info(fmt.Sprintf("Hellooooooo! It is I, the protector of secrets - %s: %s", serverKey, address***REMOVED******REMOVED***
	err := http.ListenAndServe(address, router***REMOVED***

***REMOVED***
		logger.Error(fmt.Sprintf("failed to start server: %s", err***REMOVED******REMOVED***
***REMOVED***
***REMOVED***

func main(***REMOVED*** {
	router := chi.NewRouter(***REMOVED***
	logger := libhttp.SetupChiLogger(serverKey***REMOVED***

	var host string
	if utils.IsProduction {
		host = "0.0.0.0"
***REMOVED*** else {
		host = "localhost"
***REMOVED***

	var server = &ImagineAuthServer{
		ImagineServer: &libhttp.ImagineServer{
			Host:   host,
			Key:    serverKey,
			Logger: logger,
***REMOVED***
***REMOVED***

	config, err := server.ReadConfig(***REMOVED***
***REMOVED***
		panic("Unable to read config file"***REMOVED***
***REMOVED***

	server.Port = config.GetInt(fmt.Sprintf("servers.%s.port", serverKey***REMOVED******REMOVED***
	server.Launch(router***REMOVED***
***REMOVED***
