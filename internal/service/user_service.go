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
	RegisterUser(name, email, password string) (string, string, *models.User, error)
	LoginUser(email, password string) (string, string, *models.User, error)
	GetUserByID(id uint) (*models.User, error)
	GoogleLogin(ctx context.Context, code string, oauthConf *oauth2.Config) (string, string, *models.User, error)
	RefreshToken(refreshToken string) (string, string, *models.User, error)
	UpdateUserProfile(id uint, name string) (*models.User, error)
	ChangePassword(id uint, oldPassword, newPassword string) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) RegisterUser(name, email, password string) (string, string, *models.User, error) {
	// Check if user exists
	existingUser, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", "", nil, err
	}
	if existingUser != nil {
		return "", "", nil, errors.New("email already in use")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", nil, err
	}

	user := &models.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		// IsAdmin and Verified are false by default as per model
	}

	err = s.repo.CreateUser(user)
	if err != nil {
		return "", "", nil, err
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
		Name  string `json:"name"`
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
			Name:     userInfo.Name,
			Email:    userInfo.Email,
			Password: "", // No password for Google users
			GoogleID: &userInfo.ID,
			Verified: true, // Google emails are verified
		}
		if err := s.repo.CreateUser(user); err != nil {
			return "", "", nil, err
		}
	} else {
		// Update existing user with Google ID and Name if missing
		needsUpdate := false
		if user.GoogleID == nil || *user.GoogleID != userInfo.ID {
			user.GoogleID = &userInfo.ID
			user.Verified = true // Mark as verified since Google confirmed the email
			needsUpdate = true
		}
		if user.Name == "" && userInfo.Name != "" {
			user.Name = userInfo.Name
			needsUpdate = true
		}
		if needsUpdate {
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

func (s *userService) UpdateUserProfile(id uint, name string) (*models.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	user.Name = name
	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}


func (s *userService) ChangePassword(id uint, oldPassword, newPassword string) error {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// For Google login users without a password
	if user.Password == "" {
		return errors.New("cannot change password for OAuth user")
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	if err := s.repo.UpdateUser(user); err != nil {
		return err
	}

	return nil
}
