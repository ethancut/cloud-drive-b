package auth

import "time"

const (
	AccessTokenExpiry  = 15 * time.Minute   // 15 min
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
)
