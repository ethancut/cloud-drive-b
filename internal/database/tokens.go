package database

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

type Claims struct {
	UserID int       `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"

	AccessTokenExpiry  = 15 * time.Minute   // 15 min
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GenerateTokenPair(userID int) (*TokenPair, error) {
	accessClaims := Claims{
		UserID: userID,
		Type:   TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(os.Getenv("JWT_SECRET")))

	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	jti := make([]byte, 16)
	if _, err := rand.Read(jti); err != nil {
		return nil, err
	}
	refreshClaims := Claims{
		UserID: userID,
		Type:   TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        hex.EncodeToString(jti),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}
	// return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
func ValidateAccessToken(authHeader string) (int, error) {
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return -1, errors.New("missing or invalid Bearer token prefix")
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	claims, err := parseToken(tokenString, os.Getenv("JWT_SECRET"))
	if err != nil {
		return -1, err
	}

	if claims.Type != TokenTypeAccess {
		return -1, errors.New("invalid token type: expected access token")
	}

	return claims.UserID, nil
}

func RefreshTokens(refreshTokenString string) (*TokenPair, error) {
	claims, err := parseToken(refreshTokenString, os.Getenv("JWT_REFRESH_SECRET"))
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	if claims.Type != TokenTypeRefresh {
		return nil, errors.New("invalid token type: expected refresh token")
	}

	// TODO: Check JTI against a revoked token db here

	return GenerateTokenPair(claims.UserID)
}
func parseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
