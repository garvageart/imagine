package routes

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	_ "github.com/joho/godotenv/autoload"
	"gorm.io/gorm"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"golang.org/x/oauth2"

	"imagine/internal/auth"
	oauth "imagine/internal/auth/oauth"
	"imagine/internal/crypto"
	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/uid"
)

type ImagineAuthCodeFlow struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// Idk what this is or what I meant to put here
type ImagineAuthPasswordFlow struct {
	State string
}

func AuthRouter(db *gorm.DB, logger *slog.Logger) *chi.Mux {
	router := chi.NewRouter()

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
		apiKeyId, err := uid.Generate()
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"error generating id for api key",
				"There was an error generating your API key",
			)
			return
		}

		authUser, _ := libhttp.UserFromContext(req)

		apiDataDocument := &entities.APIKey{
			UID:       apiKeyId,
			KeyHashed: keys["hashed_key"],
			UserUID:   authUser.Uid,
			Revoked:   false,
		}

		tx := db.Create(apiDataDocument)
		if tx.Error != nil {
			if (tx.Error == gorm.ErrDuplicatedKey) || (tx.Error == gorm.ErrInvalidData) {
				// Return the key
				libhttp.ServerError(res, req, err, logger, nil, "error inserting api key into database", "Something went wrong on our side, please try again later")
			}
			return
		}

		logger.Info("Generated an API key", slog.String("request_id", libhttp.GetRequestID(req)))
		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.APIKeyResponse{ConsumerKey: consumerKey})
	})

	router.Post("/login", func(res http.ResponseWriter, req *http.Request) {
		// Accept minimal login payload to avoid coupling to entities
		var login struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		err := render.DecodeJSON(req.Body, &login)
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"invalid request body",
				"Something went wrong, please try again later",
			)
			return
		}

		if login.Email == "" || login.Password == "" {
			render.Status(req, http.StatusBadRequest)
			render.JSON(res, req, dto.ErrorResponse{Error: "Required fields are missing"})
			return
		}

		// Fetch password hash and uid directly from users table by email
		var row struct {
			UID      string
			Password string
		}
		tx := db.Table("users").Select("uid, password").Where("email = ?", login.Email).Scan(&row)
		if tx.Error != nil || row.Password == "" {
			if tx.Error == gorm.ErrRecordNotFound || row.Password == "" {
				render.Status(req, http.StatusNotFound)
				render.JSON(res, req, dto.ErrorResponse{Error: "Cannot find user with provided email"})
				return
			}
			return
		}

		argon := crypto.CreateArgon2Hash(3, 32, 2, 32, 16)
		dbPass := strings.Split(row.Password, ":")
		if len(dbPass) != 2 {
			render.Status(req, http.StatusInternalServerError)
			render.JSON(res, req, dto.ErrorResponse{Error: "Invalid password format"})
			return
		}

		salt, err := hex.DecodeString(dbPass[0])
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to decode salt",
				"Something went wrong, please try again later",
			)
			return
		}

		storedHash, err := hex.DecodeString(dbPass[1])
		if err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"Failed to decode hash",
				"Something went wrong, please try again later",
			)
			return
		}

		inputHash, _ := argon.Hash([]byte(login.Password), salt)
		isValidPass := bytes.Equal(inputHash, storedHash)

		if !isValidPass {
			render.Status(req, http.StatusUnauthorized)
			render.JSON(res, req, dto.ErrorResponse{Error: "Invalid password"})
			return
		}

		// Create auth token and persistent session
		authToken := auth.GenerateAuthToken()
		expiryTime := carbon.Now().AddYear().StdTime()
		http.SetCookie(res, libhttp.CreateAuthTokenCookie(expiryTime, authToken))

		// Persist session for server-side validation
		lastActive := time.Now()
		sess := entities.Session{
			Token:      authToken,
			Uid:        uid.MustGenerate(),
			UserUid:    row.UID,
			ClientIp:   &req.RemoteAddr,
			UserAgent:  ptrString(req.UserAgent()),
			LastActive: &lastActive,
			ExpiresAt:  &expiryTime,
		}

		if err := db.Create(&sess).Error; err != nil {
			libhttp.ServerError(res, req, err, logger, nil,
				"failed to create session",
				"Something went wrong while signing you in. Please try again.",
			)
			return
		}

		logger.Info("user authenticated", slog.String("request_id", libhttp.GetRequestID(req)))
		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.MessageResponse{Message: "User authenticated"})
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

		state, err := gonanoid.New(32)
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
			Name:     libhttp.RedirectCookie,
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
		var actualUserData struct {
			Email   string `json:"email"`
			Name    string `json:"name"`
			Picture string `json:"picture"`
		}

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

			actualUserData.Email = resp.Email
			actualUserData.Name = resp.Name
			actualUserData.Picture = resp.Picture
		case "github":
			resp, err := oauth.GithubOAuthHandler(res, req, logger)
			if err != nil {
				libhttp.ServerError(res, req, err, logger, nil,
					"Error getting user data from Github",
					"We encountered an issue while trying to fetch your Github profile. Please try again.",
				)
				return
			}

			actualUserData.Email = resp.GetEmail()
			actualUserData.Name = resp.GetName()
			actualUserData.Picture = resp.GetAvatarURL()
		default:
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("OAuth provider unsupported"))
			http.Redirect(res, req, "localhost:7777/", http.StatusTemporaryRedirect)
		}

		expiryTime := carbon.Now().AddYear().StdTime()

		tokenString := auth.GenerateAuthToken()

		// at this point, the state has been validated to be correct
		// and unmodified to use that
		http.SetCookie(res, &http.Cookie{
			Name:     libhttp.StateCookie,
			Value:    state,
			Expires:  expiryTime,
			Path:     "/",
			Secure:   true,
			SameSite: http.SameSiteNoneMode, //FIXME: this needs to change to same site
		})

		// delete the temporary redirect state from the browser
		http.SetCookie(res, &http.Cookie{
			Name:     libhttp.RedirectCookie,
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
		})

		http.SetCookie(res, libhttp.CreateAuthTokenCookie(expiryTime, tokenString))

		logger.Info("User logged in with OAuth", slog.String("provider", provider))
		render.JSON(res, req, actualUserData)
	})

	router.Post("/logout", func(res http.ResponseWriter, req *http.Request) {
		if cookie, err := req.Cookie(libhttp.AuthTokenCookie); err == nil && cookie.Value != "" {
			// don't fail the logout if DB delete errors
			tx := db.Where("token = ?", cookie.Value).Delete(&entities.Session{})
			if tx.Error != nil {
				logger.Warn("failed to delete session on logout", slog.String("request_id", libhttp.GetRequestID(req)), slog.Any("error", tx.Error))
			}

			// Invalidate any in-memory cache for this session token so other requests
			// don't continue using a now-deleted session until their cache entry expires.
			libhttp.ClearSessionCache(cookie.Value)
		}

		// clear anything related to auth in the browser. even stuff that may linger
		libhttp.ClearCookie(libhttp.AuthTokenCookie, res)
		libhttp.ClearCookie(libhttp.StateCookie, res)
		libhttp.ClearCookie(libhttp.RedirectCookie, res)
		libhttp.ClearCookie(libhttp.RefreshTokenCookie, res)

		render.Status(req, http.StatusOK)
		render.JSON(res, req, dto.MessageResponse{Message: "Logged out"})
	})

	return router
}
