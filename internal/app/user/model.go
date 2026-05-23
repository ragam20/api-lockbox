package user

import "time"

type User struct {
	ID               int64
	Name             string
	Email            string
	ImgURL           *string
	AuthVerifierSalt string
	MasterKeySalt    string
	EmailVerified    bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
