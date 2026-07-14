package utils

import (
	"time"

	"github.com/dhanuprys/momenu-backend-fiber/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

type JWTClaim struct {
	UserID   uint   `json:"user_id"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
	Verified bool   `json:"verified"`
	jwt.RegisteredClaims
}

type RefreshTokenClaim struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken generates a new JWT token for a user (short lived)
func GenerateToken(userID uint, email string, isAdmin, verified bool) (string, error) {
	expirationTime := time.Now().Add(15 * time.Minute) // 15 minutes for access token
	claims := &JWTClaim{
		UserID:   userID,
		Email:    email,
		IsAdmin:  isAdmin,
		Verified: verified,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "momenu",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	return tokenString, err
}

// ValidateToken parses and validates a JWT token
func ValidateToken(signedToken string) (*JWTClaim, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*JWTClaim)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}

// GenerateRefreshToken generates a long-lived refresh token
func GenerateRefreshToken(userID uint) (string, error) {
	expirationTime := time.Now().Add(7 * 24 * time.Hour) // 7 days
	claims := &RefreshTokenClaim{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "momenu-refresh",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	return tokenString, err
}

// ValidateRefreshToken parses and validates a refresh token
func ValidateRefreshToken(signedToken string) (*RefreshTokenClaim, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&RefreshTokenClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		},
	)
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*RefreshTokenClaim)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}
	return claims, nil
}
