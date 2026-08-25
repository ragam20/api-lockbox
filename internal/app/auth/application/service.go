package application

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"lockbox/internal"
	"lockbox/internal/app/auth"
	"lockbox/internal/app/auth/domain"
	"lockbox/internal/app/vault"
	"lockbox/internal/cache"
	"lockbox/internal/config"
	dbtx "lockbox/internal/database/transaction"
	"lockbox/internal/http/response"
	"lockbox/internal/idgen"
	"lockbox/internal/smtp"

	"gorm.io/gorm"
)

const (
	verificationTokenTTL    = 10 * time.Minute
	registerSessionTTL      = 5 * time.Minute
	otpTTL                  = 1*time.Minute + (10 * time.Second)
	externalResourceTimeout = 5 * time.Second

	otpDigits           = 6
	otpExpiresInSeconds = 60
	tokenByteSize       = 32
)

type (
	AuthService interface {
		RequestOTP(ctx context.Context, email string) (response.RequestOTPResponse, error)
		VerifyOTP(ctx context.Context, input VerifyOTPInput) (response.VerifyOTPResponse, error)
		Logout(ctx context.Context, userID int64, token string)
		Login(ctx context.Context, loginInput LoginInput) (bool, error)
		RefreshToken(ctx context.Context, refreshToken string) (response.TokenResponse, error)
		GetRegisterSession(ctx context.Context, session string) (response.RegisterSessionResponse, error)
		SendVerifyLink(ctx context.Context, input SendVerificationLinkInput) error
		VerifyEmail(ctx context.Context, token string) (string, error)
	}
	authService struct {
		txManager    dbtx.TransactionManager
		authRepo     domain.AuthRepository
		vaultRepo    vault.VaultRepository
		cache        cache.Store
		mail         smtp.MailService
		appCfg       *config.AppConfig
		jwtCfg       *config.JWTConfig
		idGen        idgen.IDGenerator
		tokenManager auth.TokenManager
	}
)

func NewAuthService(
	txManager dbtx.TransactionManager,
	authRepo domain.AuthRepository,
	vaultRepo vault.VaultRepository,
	cache cache.Store,
	mail smtp.MailService,
	appCfg *config.AppConfig,
	jwtCfg *config.JWTConfig,
	idGen idgen.IDGenerator,
	tokenManager auth.TokenManager,
) AuthService {
	return &authService{
		txManager:    txManager,
		authRepo:     authRepo,
		vaultRepo:    vaultRepo,
		cache:        cache,
		mail:         mail,
		appCfg:       appCfg,
		jwtCfg:       jwtCfg,
		idGen:        idGen,
		tokenManager: tokenManager,
	}
}

func (s *authService) RequestOTP(ctx context.Context, email string) (response.RequestOTPResponse, error) {
	email = normalizeEmail(email)
	if email == "" {
		return response.RequestOTPResponse{}, response.BadRequestError(auth.ErrFieldEmailRequired.Error(), "email is required")
	}

	if exists, _ := s.authRepo.EmailExists(ctx, email); !exists {
		return response.RequestOTPResponse{
			ExpiresIn: otpExpiresInSeconds,
		}, nil
	}

	otp, err := generateOTP()
	if err != nil {
		return response.RequestOTPResponse{}, response.InternalServerError("failed to generate otp")
	}

	if err := s.cache.Set(ctx, otpCacheKey(email), otp, otpTTL); err != nil {
		return response.RequestOTPResponse{}, response.InternalServerError("failed to store otp")
	}

	if err := s.mail.SendOTP(email, otp); err != nil {
		_ = s.cache.Del(ctx, otpCacheKey(email))
		return response.RequestOTPResponse{}, err
	}

	return response.RequestOTPResponse{
		ExpiresIn: otpExpiresInSeconds,
	}, nil
}

func (s *authService) VerifyOTP(ctx context.Context, input VerifyOTPInput) (response.VerifyOTPResponse, error) {
	email := normalizeEmail(input.Email)
	otp := strings.TrimSpace(input.OTP)
	if email == "" {
		return response.VerifyOTPResponse{}, response.BadRequestError(auth.ErrFieldEmailRequired.Error(), "email is required")
	}
	if !isOTP(otp) {
		return response.VerifyOTPResponse{}, response.BadRequestError(auth.ErrInvalidOTP.Error(), "otp must be a 6 digit number")
	}

	storedOTP, err := s.cache.Get(ctx, otpCacheKey(email))
	if err != nil {
		if cache.IsMiss(err) {
			return response.VerifyOTPResponse{}, response.BadRequestError(auth.ErrInvalidOTP.Error(), "otp is invalid or expired")
		}
		return response.VerifyOTPResponse{}, response.InternalServerError("failed to validate otp")
	}
	if storedOTP != otp {
		return response.VerifyOTPResponse{}, response.BadRequestError(auth.ErrInvalidOTP.Error(), "otp is invalid or expired")
	}

	user, err := s.authRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, auth.ErrUserNotFound) {
			return response.VerifyOTPResponse{}, response.BadRequestError(auth.ErrUserNotFound.Error(), "user not found")
		}
		return response.VerifyOTPResponse{}, response.InternalServerError(err.Error())
	}

	_ = s.cache.Del(ctx, otpCacheKey(email))

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID)
	if err != nil {
		return response.VerifyOTPResponse{}, response.InternalServerError(err.Error())
	}

	userSession, err := s.tokenManager.GenerateUserSession(input.IPAddress, input.UserAgent, input.Device)
	if err != nil {
		return response.VerifyOTPResponse{}, response.InternalServerError(err.Error())
	}

	if err := s.authRepo.CreateSession(ctx, user.ID, userSession); err != nil {
		return response.VerifyOTPResponse{}, response.BadRequestError(auth.ErrInvalidCredential.Error(), "invalid create session")
	}

	return response.VerifyOTPResponse{
		Email:      email,
		IsVerified: true,
		Token: response.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: userSession.RefreshToken,
		},
	}, nil
}

func (s *authService) Logout(ctx context.Context, userID int64, token string) {
	if userID < 0 {
		return
	}

	if token == "" {
		return
	}

	tokenHash := internal.HashString(token)
	if err := s.authRepo.RevokeSession(ctx, userID, tokenHash); err != nil {
		if errors.Is(err, auth.ErrTokenNotFound) {
			return
		}
	}
}

func (s *authService) Login(ctx context.Context, loginInput LoginInput) (bool, error) {
	email := normalizeEmail(loginInput.Email)
	if email == "" {
		return false, response.BadRequestError(auth.ErrFieldEmailRequired.Error(), "email is required")
	}

	user, err := s.authRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, response.BadRequestError(auth.ErrInvalidCredential.Error(), "invalid credentials")
		}
		return false, response.InternalServerError(err.Error())
	}

	if !internal.VerifyPassword([]byte(loginInput.Verifier), []byte(user.AuthVerifier)) {
		return false, response.BadRequestError(auth.ErrInvalidCredential.Error(), "invalid credentials")
	}

	return true, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (response.TokenResponse, error) {
	if s.jwtCfg == nil {
		return response.TokenResponse{}, response.InternalServerError("jwt config is required")
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return response.TokenResponse{}, response.Unauthorized()
	}

	oldRefreshTokenHash := internal.HashString(refreshToken)
	session, err := s.authRepo.GetSessionByRefreshTokenHash(ctx, oldRefreshTokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.TokenResponse{}, response.Unauthorized()
		}
		return response.TokenResponse{}, response.InternalServerError("failed to get session")
	}

	if session.RevokeAt != nil {
		return response.TokenResponse{}, response.BadRequestError(auth.ErrTokenExpired.Error(), "refresh token expired")
	}

	now := time.Now().UTC()
	if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
		fmt.Printf("Expired at %v", session.ExpiresAt)
		return response.TokenResponse{}, response.Unauthorized()
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(session.UserID)
	if err != nil {
		return response.TokenResponse{}, response.InternalServerError(err.Error())
	}

	newRefreshToken, err := s.tokenManager.GenerateSecureToken(tokenByteSize)
	if err != nil {
		return response.TokenResponse{}, response.InternalServerError("failed to generate refresh token")
	}

	newRefreshTokenHash := internal.HashString(newRefreshToken)
	expiresAt := now.Add(s.jwtCfg.RefreshExp)

	if err := s.authRepo.UpdateRefreshToken(ctx, session.ID, oldRefreshTokenHash, newRefreshTokenHash, expiresAt, now); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.TokenResponse{}, response.Unauthorized()
		}
		return response.TokenResponse{}, response.InternalServerError("failed to update refresh token")
	}

	return response.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) GetRegisterSession(ctx context.Context, session string) (response.RegisterSessionResponse, error) {
	if session == "" {
		return response.RegisterSessionResponse{}, response.BadRequestError(auth.ErrSessionRequired.Error(), "session is required")
	}

	isSession, err := s.cache.Get(ctx, s.tokenManager.RegisterSessionKey(session))
	if err != nil || isSession == "" {
		return response.RegisterSessionResponse{}, response.BadRequestError(auth.ErrSessionRequired.Error(), "session not found")
	}

	email, err := s.cache.Get(ctx, s.tokenManager.RegisterSessionKey(session))
	if err != nil || email == "" {
		return response.RegisterSessionResponse{}, response.BadRequestError(auth.ErrSessionRequired.Error(), "session data not found")
	}

	return response.RegisterSessionResponse{
		Email: email,
	}, nil
}

func (s *authService) SendVerifyLink(ctx context.Context, input SendVerificationLinkInput) error {
	email := normalizeEmail(input.Email)
	if email == "" {
		return response.BadRequestError(auth.ErrFieldEmailRequired.Error(), "email is required")
	}

	exists, _ := s.authRepo.EmailExists(ctx, email)
	if exists {
		return response.NewAppErr(http.StatusConflict, auth.ErrEmailAlreadyExists.Error(), "email already registered")
	}

	token, err := s.tokenManager.GenerateSecureToken(tokenByteSize)
	if err != nil {
		return response.InternalServerError("failed to generate verification token")
	}

	if err := s.cache.Set(ctx, s.tokenManager.RegisterVerificationKey(token), email, verificationTokenTTL); err != nil {
		return response.InternalServerError("failed to store verification token")
	}

	link := s.verificationLink(token)
	if err := s.mail.SendVerifyLink(email, link); err != nil {
		return response.InternalServerError("failed to send verification link")
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", response.BadRequestError(auth.ErrTokenExpired.Error(), "token is required")
	}

	email, err := s.cache.GetDel(ctx, s.tokenManager.RegisterVerificationKey(token))
	if err != nil {
		if cache.IsMiss(err) {
			return "", response.BadRequestError(auth.ErrTokenExpired.Error(), "verification token is invalid or expired")
		}
		return "", response.InternalServerError("failed to validate verification token")
	}

	session, err := s.tokenManager.GenerateSecureToken(tokenByteSize)
	if err != nil {
		return "", response.InternalServerError("failed to generate register session")
	}

	if err := s.cache.Set(ctx, s.tokenManager.RegisterSessionKey(session), email, registerSessionTTL); err != nil {
		return "", response.InternalServerError("failed to store register session")
	}

	return session, nil
}

func (s *authService) verificationLink(token string) string {
	baseURL := strings.TrimRight(s.appCfg.FrontendURL, "/")
	return fmt.Sprintf("%s/auth/verify-email?token=%s", baseURL, url.QueryEscape(token))
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0*d", otpDigits, n.Int64()), nil
}

func otpCacheKey(email string) string {
	return fmt.Sprintf("auth_otp:%s", email)
}

func isOTP(value string) bool {
	if len(value) != otpDigits {
		return false
	}

	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}
