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
	"strings"
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
	UUID          string `json:"uuid"`
	ID            string `json:"id"`
	Username      string `json:"username"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	DisplayName   string `json:"display_name"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	Hash          string `json:"hash"`
	Salt          string `json:"salt"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	DeletedAt     string `json:"deleted_at"`
	UsedOAuth     bool   `json:"used_oauth"`
	OAuthProvider string `json:"oauth_provider"`
	OAuthState    string `json:"oauth_state"`
	UserToken     string `json:"user_token"`
}

func (server ImagineAuthServer) Launch(router *chi.Mux) {
	// Setup logger
	logger := server.Logger
	correctLogger := slog.NewLogLogger(logger.Handler(), slog.LevelDebug)
	router.Use(middleware.RequestLogger(&middleware.DefaultLogFormatter{
		Logger: correctLogger,
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

	var database db.DBClient = db.DB{
		Address:      "localhost",
		Port:         27017,
		User:         os.Getenv("MONGO_USER"),
		Password:     os.Getenv("MONGO_PASSWORD"),
		AppName:      utils.AppName,
		DatabaseName: "imagine-dev",
		Collection:   "images",
		Context:      mongoCtx,
	}

	client, mongoErr := database.Connect()
	defer func() {
		database.Disconnect(client)
		logger.Error("Disconnected from MongoDB")
	}()

	if mongoErr != nil {
		panic("error connecting mongo db " + mongoErr.Error())
	}

	router.Get("/ping", func(res http.ResponseWriter, req *http.Request) {
		jsonResponse := map[string]any{"message": "You have been PONGED by the auth server. What do you want here? 🤨"}
		render.JSON(res, req, jsonResponse)
	})

	router.Get("/oauth", func(res http.ResponseWriter, req *http.Request) {
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

		state, err := gonanoid.New(24)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"error generating oauth state",
				"",
			)

			return
		}

		stateHash := crypto.CreateHash([]byte(state))
		encryptedState := base64.URLEncoding.EncodeToString(stateHash)

		http.SetCookie(res, &http.Cookie{
			Name:     "imag-state",
			Value:    encryptedState,
			Expires:  carbon.Now().AddYear().StdTime(),
			Path:     "/",
			Secure:   true,
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
			"sub": userEmail,         // Subject (user identifier
			"iss": serverKey,         // Issuer
			"iat": time.Now().Unix(), // Issued at
			"exp": expiryTime,        // Expiration time
			"fgp": hashString,        // Fingerprint
		})

		tokenString, err := claims.SignedString(jwtSecretBytes)
		if err != nil {
			res.Write([]byte("Error creating JWT token:" + err.Error()))
			return
		}

		jsonBytes, err := json.Marshal(actualUserData)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil, "Failed to marshal user data to JSON", "Could not process your login information.")
			return
		}

		http.SetCookie(res, &http.Cookie{
			Name:     "imag-auth_token",
			Value:    tokenString,
			Expires:  expiryTime,
			HttpOnly: true,
			Path:     "/",
			Secure:   true,
			SameSite: http.SameSiteNoneMode,
		})

		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(jsonBytes)
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

	logger.Info(fmt.Sprintf("Hellooooooo! It is I, the protector of secrets - %s: %s", serverKey, address))
	err := http.ListenAndServe(address, router)
	if err != nil {
		errMsg := fmt.Sprintf("failed to start server: %s", err)
		logger.Error(errMsg)
		panic(errMsg)
	}
}

func main() {
	router := chi.NewRouter()
	logger := libhttp.SetupChiLogger(serverKey)

	var host string
	if utils.IsProduction {
		host = "0.0.0.0"
	} else {
		host = "localhost"
	}

	var server = &ImagineAuthServer{
		ImagineServer: &libhttp.ImagineServer{
			Host:   host,
			Key:    serverKey,
			Logger: logger,
		},
	}

	config, err := server.ReadConfig()
	if err != nil {
		panic("Unable to read config file")
	}

	server.Port = config.GetInt(fmt.Sprintf("servers.%s.port", serverKey))
	server.Launch(router)
}
