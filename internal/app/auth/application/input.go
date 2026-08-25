package application

type SendVerificationLinkInput struct {
	Email string
}

type VerifyOTPInput struct {
	Email     string
	OTP       string
	UserAgent string
	IPAddress string
	Device    string
}

type LoginInput struct {
	Email    string
	Verifier string
}

type VaultInput struct {
	VaultKeySalt string
}

type RegisterInput struct {
	Session          string
	Email            string
	Name             string
	AuthVerifier     string
	AuthVerifierSalt string
	MasterKeySalt    string
	VaultKeySalt     string
}
