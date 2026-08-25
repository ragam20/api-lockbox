package domain

import "time"

type User struct {
	ID               int64
	Email            string
	Name             string
	PublicImgID      *string
	ImgURL           *string
	MasterKeySalt    string
	AuthVerifier     string
	AuthVerifierSalt string
	EmailVerified    bool
	UpdatedAt        time.Time
	CreatedAt        time.Time
}

type ArgonConfig struct {
	ID          int64
	UserID      int64
	Memory      uint32
	Iterations  uint32
	Parallelism uint32
}

type UserSession struct {
	ID               int64
	UserID           int64
	Device           string
	RefreshToken     string
	RefreshTokenHash string
	UserAgent        string
	IPAddress        string
	LastUsedAt       time.Time
	ExpiresAt        time.Time
	RevokeAt         *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}
