package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"

	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/oauth2"

	"imagine/common/auth"
	oauth "imagine/common/auth/oauth"
	"imagine/common/crypto"
	libhttp "imagine/common/http"
	"imagine/db"
	imalog "imagine/log"
	"imagine/utils"
)

const (
	serverKey = "auth-server"
)

type ImagineAuthServer struct {
	*libhttp.ImagineServer
	oauth.ImagineOAuth
}

type ImagineAuthCodeFlow struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// Idk what this is or what I meant to put here
type ImagineAuthPasswordFlow struct {
	State string
}

type ImagineUser struct {
	UUID          string `json:"uuid" bson:"uuid"`
	ID            string `json:"id" bson:"id"`
	Username      string `json:"username" bson:"username"`
	FirstName     string `json:"first_name" bson:"first_name"`
	LastName      string `json:"last_name" bson:"last_name"`
	DisplayName   string `json:"display_name" bson:"display_name"`
	Email         string `json:"email" bson:"email"`
	Password      string `json:"password" bson:"password"`
	CreatedAt     string `json:"created_at" bson:"created_at"`
	UpdatedAt     string `json:"updated_at" bson:"updated_at"`
	DeletedAt     string `json:"deleted_at" bson:"deleted_at"`
	UsedOAuth     bool   `json:"used_oauth" bson:"used_oauth"`
	OAuthProvider string `json:"oauth_provider" bson:"oauth_provider"`
	OAuthState    string `json:"oauth_state" bson:"oauth_state"`
	UserToken     string `json:"user_token" bson:"user_token"`
}

type ImagineAPIData struct {
	APIKeyHashed    string    `json:"api_key_hashed" bson:"api_key_hashed"`
	GeneratedAt     time.Time `json:"generated_at" bson:"generated_at"`
	ApplicationID   string    `json:"application_id" bson:"application_id"`
	ApplicationName string    `json:"application_name" bson:"application_name"`
}

func (server ImagineAuthServer) Launch(router *chi.Mux) {
	// Setup logger
	logger := server.Logger
	httpMiddlewareLogger := slog.NewLogLogger(logger.Handler(), imalog.DefaultLogLevel)
	router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: httpMiddlewareLogger,
	}))

	// Setup general middleware
	router.Use(middleware.AllowContentEncoding("deflate", "gzip"))
	router.Use(middleware.RequestID)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://localhost:7777", "http://localhost:7777"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "set-cookie"},
		AllowCredentials: true,
	}))

	// Set up auth
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtSecretBytes := []byte(jwtSecret)

	// Setup DB
	mongoCtx, cancelMongo := context.WithTimeout(context.Background(), time.Second*60)
	defer cancelMongo()

	var database = &db.DB{
		Address:        "localhost",
		Port:           27017,
		User:           os.Getenv("MONGO_USER"),
		Password:       os.Getenv("MONGO_PASSWORD"),
		AppName:        utils.AppName,
		DatabaseName:   "imagine-dev",
		CollectionName: "api",
		Context:        mongoCtx,
		Logger:         logger,
	}

	client, mongoErr := database.Connect()
	defer func() {
		err := database.Disconnect(client)
		if err != nil {
			logger.Error("error disconnecting from mongodb", slog.Any("error", err))
		}

		logger.Info("Disconnected from MongoDB")
	}()

	if mongoErr != nil {
		logger.Error("error connecting to mongodb", slog.Any("error", mongoErr))
		panic("")
	}

	router.Get("/ping", func(res http.ResponseWriter, req *http.Request) {
		jsonResponse := map[string]any{"message": "You have been PONGED by the auth server. What do you want here? 🤨"}
		logger.Info("server pinged", slog.String("request_id", libhttp.GetRequestID(req)))
		render.JSON(res, req, jsonResponse)
	})

	router.Get("/apikey", func(res http.ResponseWriter, req *http.Request) {
		keys, err := auth.GenerateAPIKey()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"error generating api key",
				"There was an error generating your API key",
			)

			return
		}

		consumerKey := keys["consumer_key"]
		generatedAt := carbon.Now()

		apiDataDocument, err := db.ToBSONDocument(
			&ImagineAPIData{
				APIKeyHashed: keys["hashed_key"],
				GeneratedAt:  generatedAt.StdTime(),
			},
		)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil, "error marshaling api key to BSON into database", "Something went wrong on our side, please try again later")
			return
		}

		_, err = database.Insert(apiDataDocument)
		if err != nil {
			// Return the key
			libhttp.ServerError(res, req, err, logger, nil, "error inserting api key into database", "Something went wrong on our side, please try again later")
			return
		}

		logger.Info("Generated an API key", slog.String("request_id", libhttp.GetRequestID(req)))
		jsonResponse := map[string]any{"consumer_key": consumerKey}

		res.WriteHeader(http.StatusOK)
		render.JSON(res, req, jsonResponse)
	})

	router.Get("/oauth", func(res http.ResponseWriter, req *http.Request) {
		authTokenCookie := req.CookiesNamed("imag-auth_token")

		if len(authTokenCookie) > 0 {
			authTokenVal := authTokenCookie[0].Value
			authToken, err := jwt.Parse(authTokenVal, func(token *jwt.Token) (interface{}, error) {
				return jwtSecretBytes, nil
			})

			switch {
			case authToken.Valid:
				logger.Info("user already authenticated", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"message": "user already authenticated"}

				render.JSON(res, req, jsonResponse)
				return
			case errors.Is(err, jwt.ErrTokenMalformed):
				logger.Info("malformed token", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"error": "malformed token"}

				res.WriteHeader(http.StatusBadRequest)
				render.JSON(res, req, jsonResponse)
				return
			case errors.Is(err, jwt.ErrTokenSignatureInvalid):
				logger.Error("invalid token signature", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"error": "invalid token signature"}

				res.WriteHeader(http.StatusBadRequest)
				render.JSON(res, req, jsonResponse)
				return
			case errors.Is(err, jwt.ErrTokenExpired):
				logger.Info("token expired", slog.String("request_id", libhttp.GetRequestID(req)))
				jsonResponse := map[string]any{"error": "token expired"}

				res.WriteHeader(http.StatusForbidden)
				http.Redirect(res, req, "localhost:7777/", http.StatusTemporaryRedirect)
				render.JSON(res, req, jsonResponse)
				return
			default:
				libhttp.ServerError(res, req, err, logger, nil,
					"couldn't handle token error",
					"Something went wrong on our side, please try again later",
				)
			}
		}

		var oauthConfig *oauth2.Config
		provider := req.FormValue("provider")

		switch provider {
		case "google":
			oauthConfig = oauth.GoogleOAuthConfig
		case "github":
			oauthConfig = oauth.GithubOAuthConfig
		default:
			providerErr := errors.New("unsupported provider")
			if provider != "" {
				providerErr = errors.New("no provider... provided")
				libhttp.ServerError(res, req, providerErr, logger, nil,
					"",
					"Error siging you in. Please try again later.",
				)
			} else {
				libhttp.ServerError(res, req, providerErr, logger, nil,
					"",
					"Error siging you in. Please try again later.",
				)
			}
		}

		// idk just seems better to use 48
		state, err := gonanoid.New(48)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"error generating oauth state",
				"",
			)
			return
		}

		stateHash := crypto.CreateHash([]byte(state))
		encryptedStateB64 := base64.URLEncoding.EncodeToString(stateHash)

		// 5 minute max window to login using the generated state
		http.SetCookie(res, &http.Cookie{
			Name:     "imag-redirect-state",
			Value:    encryptedStateB64,
			Expires:  carbon.Now().AddMinutes(5).StdTime(),
			Path:     "/",
			Secure:   true,
			HttpOnly: true, // client doesn't use this value, make HttpOnly
			SameSite: http.SameSiteNoneMode,
		})

		oauthUrl, err := oauth.SetupOAuthURL(res, req, oauthConfig, provider, state)
		if err != nil {

			libhttp.ServerError(res, req, err, logger, nil,
				"",
				"",
			)
			return
		}

		http.Redirect(res, req, oauthUrl, http.StatusTemporaryRedirect)
	})

	router.Post("/oauth/{provider}", func(res http.ResponseWriter, req *http.Request) {
		provider := strings.ToLower(chi.URLParam(req, "provider"))
		state := req.FormValue("state")
		var actualUserData any // To store the user data struct
		var userEmail string

		switch provider {
		case "google":
			resp, err := oauth.GoogleOAuthHandler(res, req, logger)
			if err != nil {

				libhttp.ServerError(res, req, err, logger, nil,
					"Error getting user data from Google",
					"We encountered an issue while trying to fetch your Google profile. Please try again.",
				)
				return
			}

			userEmail = resp.Email
			actualUserData = resp
		case "github":
			resp, err := oauth.GithubOAuthHandler(res, req, logger)
			if err != nil {
				libhttp.ServerError(res, req, err, logger, nil,
					"Error getting user data from Github",
					"We encountered an issue while trying to fetch your Github profile. Please try again.",
				)
				return
			}

			userEmail = resp.GetEmail()
			actualUserData = resp
		default:
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("OAuth provider unsupported"))
			http.Redirect(res, req, "localhost:7777/", http.StatusTemporaryRedirect)
		}

		userFingerprint := auth.GenerateRandomBytes(50)
		userFingerprintString := hex.EncodeToString(userFingerprint)

		tokenSHA256 := sha256.New()
		tokenSHA256.Write([]byte(userFingerprintString))
		sha256Sum := tokenSHA256.Sum(nil)
		hashString := hex.EncodeToString(sha256Sum)

		expiryTime := carbon.Now().AddYear().StdTime()

		// Create a new JWT token with claims
		claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			// This probably bad and maybe we should generate an ID here instead
			"sub": userEmail, // Subject (user identifier)
			"aud": "viz",
			"iss": serverKey,         // Issuer
			"iat": time.Now().Unix(), // Issued at
			"exp": expiryTime,        // Expiration time
			"fgp": hashString,        // Fingerprint
		})

		tokenString, err := claims.SignedString(jwtSecretBytes)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to sign JWT token",
				"Could not process your login information. Please try again later",
			)
			return
		}

		jsonBytes, err := json.Marshal(actualUserData)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil, "Failed to marshal user data to JSON", "Could not process your login information.")
			return
		}

		// at this point, the state has been validated to be correct
		// and unmodified to use that
		http.SetCookie(res, &http.Cookie{
			Name:     "imag-state",
			Value:    state,
			Expires:  carbon.Now().AddYear().StdTime(),
			Path:     "/",
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})

		// delete the temporary redirect state from the browser
		http.SetCookie(res, &http.Cookie{
			Name:     "imag-redirect-state",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
		})

		http.SetCookie(res, &http.Cookie{
			Name:     "imag-auth_token",
			Value:    tokenString,
			Expires:  expiryTime,
			HttpOnly: true,
			Path:     "/",
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})

		logger.Info("User logged in with OAuth", slog.String("provider", provider))
		render.JSON(res, req, jsonBytes)
	})

	router.Put("/user/create", func(res http.ResponseWriter, req *http.Request) {
		oauthRedirect := req.FormValue("oauth_redirect")
		continueUrl := req.FormValue("continue")
		oauthState := req.CookiesNamed("imag-state")[0]
		email := req.FormValue("email")
		name := req.FormValue("name")
		userUUID := uuid.New()
		userIdBytes, err := userUUID.MarshalBinary()
		usedOAuth := oauthRedirect != ""

		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"error marshaling uuid for user id",
				"Error creating your account, please try again later",
			)
		}

		userId := hex.EncodeToString(userIdBytes)
		userStruct := &ImagineUser{
			UUID:          userUUID.String(),
			ID:            userId,
			Email:         email,
			Username:      name,
			CreatedAt:     carbon.Now().String(),
			UsedOAuth:     usedOAuth,
			OAuthState:    oauthState.Value,
			OAuthProvider: oauthRedirect,
		}

		userDocument, err := db.ToBSONDocument(userStruct)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"error marshalling user document",
				"Error creating your account, please try again later",
			)
		}

		_, err = database.Insert(userDocument)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"err creating user account on database",
				"Error creating your account, please try again later",
			)
		}

		time.Sleep(2000)
		if continueUrl != "" {
			continueUrl = "/"
		}

		http.Redirect(res, req, continueUrl, http.StatusTemporaryRedirect)
	})

	address := fmt.Sprintf("%s:%d", server.Host, server.Port)

	go func() {
		logger.Info(fmt.Sprintf("Hellooooooo! It is I, the protector of secrets - %s: %s", serverKey, address))
		err := http.ListenAndServe(address, router)
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error(fmt.Sprintf("failed to start server: %s", err))
			}

			errMsg := fmt.Sprintf("failed to start server: %s", err)
			logger.Error(errMsg)
			panic("")
		}
	}()

	// Taken and adjusted from https://github.com/bluesky-social/social-app/blob/main/bskyweb/cmd/bskyweb/server.go
	// Wait for a signal to exit.
	logger.Info("registering OS exit signal handler")
	quit := make(chan struct{})
	exitSignals := make(chan os.Signal, 1)
	signal.Notify(exitSignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-exitSignals
		logger.Info(fmt.Sprintf("received OS exit signal: %s", sig))

		// Trigger the return that causes an exit.
		close(quit)
	}()
	<-quit
	logger.Info("graceful shutdown complete")
}

func main() {
	router := chi.NewRouter()
	logger := libhttp.SetupChiLogger(serverKey)

	server := ImagineAuthServer{ImagineServer: libhttp.ImagineServers[serverKey]}
	server.ImagineServer.Logger = logger

	server.Launch(router)
}
