package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"lockbox/internal"
	"lockbox/internal/app/vault"
	"lockbox/internal/cache"
	"lockbox/internal/config"
	"lockbox/internal/http/response"
	"lockbox/internal/idgen"
	"lockbox/internal/smtp"

	"gorm.io/gorm"
)

const (
	verificationTokenTTL    = 10 * time.Minute
	registerSessionTTL      = 5 * time.Minute
	externalResourceTimeout = 5 * time.Second

	tokenByteSize = 32
)

type (
	AuthService interface {
		Logout(ctx context.Context, userID int64, token string)
		Login(ctx context.Context, loginInput LoginInput) (response.LoginResponse, error)
		RefreshToken(ctx context.Context, refreshToken string) (response.TokenResponse, error)
		GetRegisterSession(ctx context.Context, session string) (response.RegisterSessionResponse, error)
		SendVerifyLink(ctx context.Context, input SendVerificationLinkInput) error
		VerifyEmail(ctx context.Context, token string) (string, error)
		Register(ctx context.Context, input RegisterInput) (*User, error)
	}
	authService struct {
		authRepo     AuthRepository
		vaultService vault.VaultService
		cache        cache.Store
		mail         smtp.MailService
		appCfg       *config.AppConfig
		jwtCfg       *config.JWTConfig
		idGen        idgen.IDGenerator
		tokenManager TokenManager
	}
)

func NewAuthService(
	authRepo AuthRepository,
	vaultService vault.VaultService,
	cache cache.Store,
	mail smtp.MailService,
	appCfg *config.AppConfig,
	jwtCfg *config.JWTConfig,
	idGen idgen.IDGenerator,
	tokenManager TokenManager,
) AuthService {
	return &authService{
		authRepo:     authRepo,
		vaultService: vaultService,
		cache:        cache,
		mail:         mail,
		appCfg:       appCfg,
		jwtCfg:       jwtCfg,
		idGen:        idGen,
		tokenManager: tokenManager,
	}
}

func (s *authService) Logout(ctx context.Context, userID int64, token string) {
	if userID < 0 {
		return
	}

	if token == "" {
		return
	}

	tokenHash := internal.HashString(token)
	_ = s.authRepo.RevokeSession(ctx, userID, tokenHash)
}

func (s *authService) Login(ctx context.Context, loginInput LoginInput) (response.LoginResponse, error) {
	email := normalizeEmail(loginInput.Email)
	if email == "" {
		return response.LoginResponse{}, response.BadRequestError(ErrFieldEmailRequired.Error(), "email is required")
	}

	user, err := s.authRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.LoginResponse{}, response.BadRequestError(ErrInvalidCredential.Error(), "invalid credentials")
		}
		return response.LoginResponse{}, response.InternalServerError(err.Error())
	}

	if !internal.VerifyPassword([]byte(loginInput.Verifier), []byte(user.AuthVerifier)) {
		return response.LoginResponse{}, response.BadRequestError(ErrInvalidCredential.Error(), "invalid credentials")
	}

	accessToken, err := s.tokenManager.GenerateAccessToken(user.ID)
	if err != nil {
		return response.LoginResponse{}, response.InternalServerError(err.Error())
	}

	userSession, err := s.tokenManager.GenerateUserSession(loginInput.IPAddress, loginInput.UserAgent, loginInput.Device)
	if err != nil {
		return response.LoginResponse{}, response.InternalServerError(err.Error())
	}

	if err := s.authRepo.CreateSession(ctx, user.ID, userSession); err != nil {
		return response.LoginResponse{}, response.BadRequestError(ErrInvalidCredential.Error(), "invalid create session")
	}

	return response.LoginResponse{
		Token: response.TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: userSession.RefreshToken,
		},
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (response.TokenResponse, error) {
	if s.jwtCfg == nil {
		return response.TokenResponse{}, response.InternalServerError("jwt config is required")
	}

	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return response.TokenResponse{}, response.Unauthorized()
	}

	refreshTokenHash := internal.HashString(refreshToken)
	session, err := s.authRepo.GetSessionByRefreshTokenHash(ctx, refreshTokenHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.TokenResponse{}, response.Unauthorized()
		}
		return response.TokenResponse{}, response.InternalServerError("failed to get session")
	}

	if session.RevokeAt != nil {
		return response.TokenResponse{}, response.BadRequestError(ErrTokenExpired.Error(), "refresh token expired")
	}

	now := time.Now().UTC()
	if !session.ExpiresAt.IsZero() && now.After(session.ExpiresAt) {
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
	expiresAt := now.Add(s.jwtCfg.Expired)

	if err := s.authRepo.UpdateRefreshToken(ctx, session.ID, refreshTokenHash, newRefreshTokenHash, expiresAt, now); err != nil {
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
		return response.RegisterSessionResponse{}, response.BadRequestError(ErrSessionRequired.Error(), "session is required")
	}

	isSession, err := s.cache.Get(ctx, s.tokenManager.RegisterSessionKey(session))
	if err != nil || isSession == "" {
		return response.RegisterSessionResponse{}, response.BadRequestError(ErrSessionRequired.Error(), "session not found")
	}

	email, err := s.cache.Get(ctx, s.tokenManager.RegisterSessionKey(session))
	if err != nil || email == "" {
		return response.RegisterSessionResponse{}, response.BadRequestError(ErrSessionRequired.Error(), "session data not found")
	}

	return response.RegisterSessionResponse{
		Email: email,
	}, nil
}

func (s *authService) SendVerifyLink(ctx context.Context, input SendVerificationLinkInput) error {
	email := normalizeEmail(input.Email)
	if email == "" {
		return response.BadRequestError(ErrFieldEmailRequired.Error(), "email is required")
	}

	exists, err := s.authRepo.EmailExists(ctx, email)
	if err != nil {
		return response.InternalServerError("failed to check email")
	}
	if exists {
		return response.NewAppErr(http.StatusConflict, ErrEmailAlreadyExists.Error(), "email already registered")
	}

	token, err := s.tokenManager.GenerateSecureToken(tokenByteSize)
	if err != nil {
		return response.InternalServerError("failed to generate verification token")
	}

	if err := s.cache.Set(ctx, s.tokenManager.RegisterVerificationKey(token), email, verificationTokenTTL); err != nil {
		return response.InternalServerError("failed to store verification token")
	}

	link := s.verificationLink(token)
	if err := s.mail.SendVerifyLink(ctx, email, link); err != nil {
		return response.InternalServerError("failed to send verification link")
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", response.BadRequestError(ErrTokenExpired.Error(), "token is required")
	}

	email, err := s.cache.GetDel(ctx, s.tokenManager.RegisterVerificationKey(token))
	if err != nil {
		if cache.IsMiss(err) {
			return "", response.BadRequestError(ErrTokenExpired.Error(), "verification token is invalid or expired")
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

func (s *authService) Register(ctx context.Context, input RegisterInput) (*User, error) {
	input.Email = normalizeEmail(input.Email)
	input.Name = strings.TrimSpace(input.Name)
	input.Session = strings.TrimSpace(input.Session)

	if err := validateRegisterInput(input); err != nil {
		return nil, err
	}

	sessionEmail, err := s.cache.Get(ctx, s.tokenManager.RegisterSessionKey(input.Session))
	if err != nil {
		if cache.IsMiss(err) {
			return nil, response.BadRequestError(ErrSessionExpired.Error(), "register session is invalid or expired")
		}
		return nil, response.InternalServerError("failed to validate register session")
	}
	if sessionEmail != input.Email {
		return nil, response.ForbiddenError("register session email mismatch")
	}

	exists, err := s.authRepo.EmailExists(ctx, input.Email)
	if err != nil {
		return nil, response.InternalServerError("failed to check email")
	}
	if exists {
		return nil, response.NewAppErr(http.StatusConflict, ErrEmailAlreadyExists.Error(), "email already registered")
	}

	user := User{
		ID:               s.idGen.NewID(),
		Email:            input.Email,
		Name:             input.Name,
		MasterKeySalt:    input.MasterKeySalt,
		AuthVerifier:     input.AuthVerifier,
		AuthVerifierSalt: input.AuthVerifierSalt,
		EmailVerified:    true,
	}

	if err := s.authRepo.WithTransaction(ctx, func(repo AuthRepository) error {
		if err := repo.CreateUser(ctx, user); err != nil {
			return err
		}

		if err := s.vaultService.CreateArgonConfig(ctx, user.ID); err != nil {
			return err
		}

		if err := s.vaultService.CreateVault(ctx, user.ID, input.VaultKeySalt); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, response.BadRequestError(ErrFailedRegister.Error(), "failed to register")
	}

	if err := s.cache.Del(ctx, s.tokenManager.RegisterSessionKey(input.Session)); err != nil {
		return nil, response.InternalServerError("failed to delete register session")
	}

	return &user, nil
}

func (s *authService) verificationLink(token string) string {
	baseURL := strings.TrimRight(s.appCfg.FrontendURL, "/")
	return fmt.Sprintf("%s/auth/verify-email?token=%s", baseURL, url.QueryEscape(token))
}

func validateRegisterInput(input RegisterInput) error {
	if input.Session == "" {
		return response.BadRequestError(ErrSessionRequired.Error(), "register session is required")
	}
	if input.Email == "" {
		return response.BadRequestError(ErrFieldEmailRequired.Error(), "email is required")
	}
	if input.Name == "" {
		return response.BadRequestError(ErrFieldNameRequired.Error(), "name is required")
	}

	base64Fields := map[string]string{
		"authVerifier":     input.AuthVerifier,
		"authVerifierSalt": input.AuthVerifierSalt,
		"masterKeySalt":    input.MasterKeySalt,
		"vaultKeySalt":     input.VaultKeySalt,
	}
	for field, value := range base64Fields {
		if !isBase64(value) {
			return response.BadRequestError(ErrValidation.Error(), fmt.Sprintf("%s must be a valid base64 string", field))
		}
	}

	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isBase64(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}

	encodings := []*base64.Encoding{
		base64.StdEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.RawURLEncoding,
	}

	for _, encoding := range encodings {
		if _, err := encoding.DecodeString(value); err == nil {
			return true
		}
	}

	return false
}
