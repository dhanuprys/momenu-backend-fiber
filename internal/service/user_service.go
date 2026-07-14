package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/dhanuprys/momenu-backend-fiber/internal/models"
	"github.com/dhanuprys/momenu-backend-fiber/internal/repository"
	"github.com/dhanuprys/momenu-backend-fiber/pkg/utils"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
)

type UserService interface {
	RegisterUser(email, password string) (*models.User, error)
	LoginUser(email, password string) (string, string, *models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GoogleLogin(ctx context.Context, code string, oauthConf *oauth2.Config) (string, string, *models.User, error)
	RefreshToken(refreshToken string) (string, string, *models.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) RegisterUser(email, password string) (*models.User, error) {
	// Check if user exists
	existingUser, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, errors.New("email already in use")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:    email,
		Password: string(hashedPassword),
		// IsAdmin and Verified are false by default as per model
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) LoginUser(email, password string) (string, string, *models.User, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", "", nil, err
	}
	if user == nil {
		return "", "", nil, errors.New("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", "", nil, errors.New("invalid email or password")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.IsAdmin, user.Verified)
	if err != nil {
		return "", "", nil, err
	}
	
	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}

	return token, refreshToken, user, nil
}

func (s *userService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *userService) GoogleLogin(ctx context.Context, code string, oauthConf *oauth2.Config) (string, string, *models.User, error) {
	token, err := oauthConf.Exchange(ctx, code)
	if err != nil {
		return "", "", nil, errors.New("failed to exchange token: " + err.Error())
	}

	client := oauthConf.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", nil, errors.New("failed to get user info: " + err.Error())
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return "", "", nil, errors.New("failed to decode user info: " + err.Error())
	}

	user, err := s.repo.GetUserByEmail(userInfo.Email)
	if err != nil {
		return "", "", nil, err
	}

	if user == nil {
		// Create new user
		user = &models.User{
			Email:    userInfo.Email,
			Password: "", // No password for Google users
			GoogleID: &userInfo.ID,
			Verified: true, // Google emails are verified
		}
		if err := s.repo.CreateUser(user); err != nil {
			return "", "", nil, err
		}
	} else {
		// Update existing user with Google ID if missing
		if user.GoogleID == nil || *user.GoogleID != userInfo.ID {
			user.GoogleID = &userInfo.ID
			user.Verified = true // Mark as verified since Google confirmed the email
			if err := s.repo.UpdateUser(user); err != nil {
				return "", "", nil, err
			}
		}
	}

	jwtToken, err := utils.GenerateToken(user.ID, user.Email, user.IsAdmin, user.Verified)
	if err != nil {
		return "", "", nil, err
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}

	return jwtToken, refreshToken, user, nil
}

func (s *userService) RefreshToken(refreshTokenStr string) (string, string, *models.User, error) {
	claims, err := utils.ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		return "", "", nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.repo.GetUserByID(claims.UserID)
	if err != nil {
		return "", "", nil, err
	}
	if user == nil {
		return "", "", nil, errors.New("user not found")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, user.IsAdmin, user.Verified)
	if err != nil {
		return "", "", nil, err
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", nil, err
	}

	return token, newRefreshToken, user, nil
}
