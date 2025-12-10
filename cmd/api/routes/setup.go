package routes

import (
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"gorm.io/gorm"

	"imagine/internal/auth"
	"imagine/internal/crypto"
	"imagine/internal/dto"
	"imagine/internal/entities"
	libhttp "imagine/internal/http"
	"imagine/internal/settings"
	"imagine/internal/uid"
	"imagine/internal/utils"
)

type setupHandlers struct {
	db     *gorm.DB
	logger *slog.Logger
}

// SetupSuperadmin handles the POST /api/setup/superadmin endpoint
func (h *setupHandlers) SetupSuperadmin(w http.ResponseWriter, req *http.Request) {
	// First, check if setup is already complete by checking the first_run_complete setting
	firstRunCompleteStr, err := settings.GetSetting(h.db, "first_run_complete", nil)
	if err == nil && firstRunCompleteStr == "true" {
		render.Status(req, http.StatusConflict)
		render.JSON(w, req, dto.ErrorResponse{Error: "Application already initialized"})
		return
	}

	var body dto.SuperadminSetupRequest
	if err := render.DecodeJSON(req.Body, &body); err != nil {
		render.Status(req, http.StatusBadRequest)
		render.JSON(w, req, dto.ErrorResponse{Error: "Invalid request body"})
		return
	}

	if body.Username == "" || body.Password == "" || string(body.Email) == "" {
		render.Status(req, http.StatusBadRequest)
		render.JSON(w, req, dto.ErrorResponse{Error: "Required fields are missing"})
		return
	}

	if !utils.IsValidEmail(string(body.Email)) {
		render.Status(req, http.StatusBadRequest)
		render.JSON(w, req, dto.ErrorResponse{Error: "Email is invalid"})
		return
	}

	// Check if user already exists
	var existingUser entities.User
	tx := h.db.Where("email = ?", string(body.Email)).Or("username = ?", body.Username).First(&existingUser)
	switch tx.Error {
	case nil:
		render.Status(req, http.StatusConflict)
		render.JSON(w, req, dto.ErrorResponse{Error: "User with this email or username already exists"})
		return
	case gorm.ErrRecordNotFound:
		// OK, continue
	default:
		h.logger.Error("setup.superadmin: db error when checking existing user", slog.Any("error", tx.Error))
		libhttp.ServerError(w, req, tx.Error, h.logger, nil,
			"Failed to check existing user",
			"Something went wrong, please try again later",
		)
		return
	}

	id, err := uid.Generate()
	if err != nil {
		libhttp.ServerError(w, req, err, h.logger, nil,
			"Failed to generate user ID",
			"Something went wrong, please try again later",
		)
		return
	}

	userEnt := entities.User{
		Uid:       id,
		Email:     string(body.Email),
		Username:  body.Username,
		FirstName: func() string { if body.FirstName != nil { return *body.FirstName } else { return "" } }(),
		LastName:  func() string { if body.LastName != nil { return *body.LastName } else { return "" } }(),
		Role:      dto.UserRoleSuperadmin, // Assign superadmin role
	}

	argon := crypto.CreateArgon2Hash(3, 32, 2, 32, 16)
	salt := argon.GenerateSalt()
	hashedPass, _ := argon.Hash([]byte(body.Password), salt)
	hashed := hex.EncodeToString(salt) + ":" + hex.EncodeToString(hashedPass)

	uwp := entities.FromUser(userEnt, &hashed)

	txErr := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&uwp).Error; err != nil {
			return err
		}

		// Set onboarding_complete to true for superadmin
		onboardingOverride := entities.SettingOverride{
			UserId: id,
			Name:   "onboarding_complete",
			Value:  "true",
		}
		if err := tx.Create(&onboardingOverride).Error; err != nil {
			return err
		}

		// Set first_run_complete to true
		firstRunSettingOverride := entities.SettingOverride{
			UserId: "", // Global setting
			Name:   "first_run_complete",
			Value:  "true",
		}
		if err := tx.Create(&firstRunSettingOverride).Error; err != nil {
			return err
		}
		
		return nil
	})

	if txErr != nil {
		libhttp.ServerError(w, req, txErr, h.logger, nil,
			"Failed to create superadmin user",
			"Something went wrong, please try again later",
		)
		return
	}

	// Generate session token for the new superadmin
	// Create auth token and persistent session
	authToken := auth.GenerateAuthToken()
	expiryTime := carbon.Now().AddYear().StdTime()
	http.SetCookie(w, libhttp.CreateAuthTokenCookie(expiryTime, authToken))

	// Persist session for server-side validation
	lastActive := time.Now()
	sess := entities.Session{
		Token:      authToken,
		Uid:        uid.MustGenerate(),
		UserUid:    id,
		ClientIp:   &req.RemoteAddr,
		UserAgent:  utils.StringPtr(req.UserAgent()),
		LastActive: &lastActive,
		ExpiresAt:  &expiryTime,
	}

	if err := h.db.Create(&sess).Error; err != nil {
		libhttp.ServerError(w, req, err, h.logger, nil,
			"failed to create session",
			"Something went wrong while signing you in. Please try again.",
		)
		return
	}

	render.Status(req, http.StatusCreated)
	render.JSON(w, req, dto.SuperadminSetupResponse{
		Message:      "Superadmin setup complete and logged in.",
		User:         uwp.User.DTO(),
		SessionToken: authToken, // Return the generated authToken
	})
}

// SetupRouter creates a new chi.Router for setup-related endpoints
func SetupRouter(dbClient *gorm.DB, logger *slog.Logger) chi.Router {
	r := chi.NewRouter()
	handlers := &setupHandlers{db: dbClient, logger: logger}

	r.Post("/superadmin", handlers.SetupSuperadmin)

	return r
}
