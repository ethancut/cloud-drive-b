package auth

import "time"

const (
	AccessTokenExpiry  = 5 * time.Second    // 15 min
	RefreshTokenExpiry = 7 * 24 * time.Hour // 7 days
)
