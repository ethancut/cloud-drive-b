package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidHeader    = errors.New("invalid authorization header format")
	ErrInvalidTokenType = errors.New("invalid token type")
	ErrInvalidClaims    = errors.New("invalid or expired token claims")
	ErrEmptySecret      = errors.New("jwt secret key cannot be empty")
)

type TokenType string

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type TokenService struct {
	secretKey          []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

func NewTokenService(secretKey string, accessTokenExpiry, refreshTokenExpiry time.Duration) (*TokenService, error) {
	if strings.TrimSpace(secretKey) == "" {
		return nil, ErrEmptySecret
	}
	return &TokenService{
		secretKey:          []byte(secretKey),
		accessTokenExpiry:  accessTokenExpiry,
		refreshTokenExpiry: refreshTokenExpiry,
	}, nil
}

func (s *TokenService) ValidateAccessToken(authHeader string) (uuid.UUID, error) {
	rawToken, err := extractBearerToken(authHeader)
	if err != nil {
		return uuid.Nil, err
	}

	claims, err := s.parseToken(rawToken)
	if err != nil {
		return uuid.Nil, err
	}
	if claims.Type != TokenTypeAccess {
		return uuid.Nil, ErrInvalidTokenType
	}
	return claims.UserID, nil
}
func (s *TokenService) GenerateTokenPair(userID uuid.UUID) (*TokenPair, error) {
	accessToken, err := s.generateToken(userID, TokenTypeAccess, s.accessTokenExpiry, "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %v", err)
	}

	jti := make([]byte, 16)
	if _, err := rand.Read(jti); err != nil {
		return nil, err
	}

	refreshToken, err := s.generateToken(userID, TokenTypeRefresh, s.refreshTokenExpiry, hex.EncodeToString(jti))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
func (s *TokenService) RotateRefreshToken(rawRefreshToken string) (*TokenPair, *Claims, error) {
	claims, err := s.parseToken(rawRefreshToken)
	if err != nil {
		return nil, nil, err
	}

	if claims.Type != TokenTypeRefresh {
		return nil, nil, fmt.Errorf("%w: expected refresh token", ErrInvalidTokenType)
	}

	pair, err := s.GenerateTokenPair(claims.UserID)
	if err != nil {
		return nil, nil, err
	}

	return pair, claims, nil
}
func (s *TokenService) generateToken(userID uuid.UUID, tokenType TokenType, duration time.Duration, jti string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        jti,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *TokenService) parseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidClaims
	}

	return claims, nil
}
func extractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", ErrInvalidHeader
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", ErrInvalidHeader
	}
	return strings.TrimSpace(parts[1]), nil
}
