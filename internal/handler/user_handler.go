package handler

import (
	"context"
	"fmt"

	"github.com/dhanuprys/momenu-backend-fiber/internal/config"
	"github.com/dhanuprys/momenu-backend-fiber/internal/service"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/response"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"github.com/gofiber/fiber/v3"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type RegisterRequest struct {
	Name           string `json:"name" form:"name" validate:"required"`
	Email          string `json:"email" form:"email" validate:"required,email"`
	Password       string `json:"password" form:"password" validate:"required,min=6"`
	TurnstileToken string `json:"turnstile_token" validate:"required"`
}

// LoginRequest defines the expected payload for login
type LoginRequest struct {
	Email          string `json:"email" form:"email" validate:"required,email"`
	Password       string `json:"password" form:"password" validate:"required"`
	TurnstileToken string `json:"turnstile_token" validate:"required"`
}

// Register handles user registration
func (h *UserHandler) Register(c fiber.Ctx) error {
	var req RegisterRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	if err := utils.VerifyTurnstile(req.TurnstileToken); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Gagal memverifikasi captcha", "INVALID_CAPTCHA")
	}

	token, refreshToken, user, err := h.userService.RegisterUser(req.Name, req.Email, req.Password)
	if err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, err.Error(), "REGISTRATION_FAILED")
	}

	return response.JSONSuccess(c, fiber.StatusCreated, "User registered successfully", fiber.Map{
		"token":         token,
		"refresh_token": refreshToken,
		"user":          user,
	}, nil)
}

// Login handles user authentication
func (h *UserHandler) Login(c fiber.Ctx) error {
	var req LoginRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	if err := utils.VerifyTurnstile(req.TurnstileToken); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Gagal memverifikasi captcha", "INVALID_CAPTCHA")
	}

	token, refreshToken, user, err := h.userService.LoginUser(req.Email, req.Password)
	if err != nil {
		return response.JSONError(c, fiber.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Login successful", fiber.Map{
		"token":         token,
		"refresh_token": refreshToken,
		"user":          user,
	}, nil)
}

// Me handles fetching the current user profile
func (h *UserHandler) Me(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return response.JSONError(c, fiber.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, "Failed to get user profile", "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "User profile retrieved", user, nil)
}

// UpdateProfileRequest defines the expected payload for updating a user profile
type UpdateProfileRequest struct {
	Name string `json:"name" validate:"required"`
}

// UpdateProfile handles updating the current user's profile
func (h *UserHandler) UpdateProfile(c fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uint)
	if !ok {
		return response.JSONError(c, fiber.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
	}

	var req UpdateProfileRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	user, err := h.userService.UpdateUserProfile(userID, req.Name)
	if err != nil {
		return response.JSONError(c, fiber.StatusInternalServerError, err.Error(), "INTERNAL_SERVER_ERROR")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "User profile updated successfully", user, nil)
}

func (h *UserHandler) getOAuthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.AppConfig.GoogleClientID,
		ClientSecret: config.AppConfig.GoogleClientSecret,
		RedirectURL:  config.AppConfig.GoogleRedirectURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

// GoogleLoginURL redirects the user to Google's consent screen
func (h *UserHandler) GoogleLoginURL(c fiber.Ctx) error {
	oauthConf := h.getOAuthConfig()
	// Create a random state token
	state := utils.GenerateSlug("state") // simple random string generator using existing util

	// Set state in a secure cookie
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		MaxAge:   300, // 5 minutes
		HTTPOnly: true,
		Secure:   config.AppConfig.Env == "production",
		SameSite: "Lax",
	})

	url := oauthConf.AuthCodeURL(state)
	return c.Redirect().To(url)
}

// GoogleCallback handles the response from Google
func (h *UserHandler) GoogleCallback(c fiber.Ctx) error {
	state := c.Query("state")
	cookieState := c.Cookies("oauth_state")

	// Clear the cookie immediately
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    "",
		MaxAge:   -1,
		HTTPOnly: true,
	})

	if state == "" || cookieState == "" || state != cookieState {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid state token", "INVALID_STATE")
	}

	code := c.Query("code")
	if code == "" {
		return response.JSONError(c, fiber.StatusBadRequest, "Missing code", "MISSING_CODE")
	}

	oauthConf := h.getOAuthConfig()
	token, refreshToken, _, err := h.userService.GoogleLogin(context.Background(), code, oauthConf)
	if err != nil {
		redirectURL := fmt.Sprintf("%s/login?error=google_login_failed", config.AppConfig.FrontendURL)
		return c.Redirect().To(redirectURL)
	}

	frontendCallback := fmt.Sprintf("%s/auth/callback?token=%s&refresh_token=%s", config.AppConfig.FrontendURL, token, refreshToken)
	return c.Redirect().To(frontendCallback)
}

// RefreshRequest defines the expected payload for refreshing a token
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshToken handles generating a new access token
func (h *UserHandler) RefreshToken(c fiber.Ctx) error {
	var req RefreshRequest
	if err := c.Bind().Body(&req); err != nil {
		return response.JSONError(c, fiber.StatusBadRequest, "Invalid request payload", "INVALID_PAYLOAD")
	}

	if errors := utils.ValidateStruct(req); errors != nil {
		return response.JSONValidationError(c, errors)
	}

	token, newRefreshToken, user, err := h.userService.RefreshToken(req.RefreshToken)
	if err != nil {
		return response.JSONError(c, fiber.StatusUnauthorized, err.Error(), "UNAUTHORIZED")
	}

	return response.JSONSuccess(c, fiber.StatusOK, "Token refreshed successfully", fiber.Map{
		"token":         token,
		"refresh_token": newRefreshToken,
		"user":          user,
	}, nil)
}
