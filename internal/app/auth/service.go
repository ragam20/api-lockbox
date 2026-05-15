package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"lockbox/internal/cache"
	"lockbox/internal/config"
	"lockbox/internal/http/response"
	"lockbox/internal/idgen"
	"lockbox/internal/smtp"
)

const (
	verificationTokenTTL    = 10 * time.Minute
	registerSessionTTL      = 5 * time.Minute
	externalResourceTimeout = 5 * time.Second

	tokenByteSize = 32

	defaultArgonMemory      = 64 * 1024
	defaultArgonIterations  = 3
	defaultArgonParallelism = 2
	defaultVaultTimeout     = 10
	defaultVaultTimeoutType = VaultTimeoutTypeLock
)

type (
	AuthService interface {
		SendVerifyLink(ctx context.Context, input SendVerificationLinkInput) error
		VerifyEmail(ctx context.Context, token string) (string, error)
		Register(ctx context.Context, input RegisterInput) (*User, error)
	}
	authService struct {
		authRepo AuthRepository
		cache    cache.Store
		mail     smtp.MailService
		appCfg   *config.AppConfig
		idGen    idgen.IDGenerator
	}
)

func NewAuthService(
	authRepo AuthRepository,
	cache cache.Store,
	mail smtp.MailService,
	appCfg *config.AppConfig,
	idGen idgen.IDGenerator,
) AuthService {
	return &authService{
		authRepo: authRepo,
		cache:    cache,
		mail:     mail,
		appCfg:   appCfg,
		idGen:    idGen,
	}
}

func (s *authService) SendVerifyLink(ctx context.Context, input SendVerificationLinkInput) error {
	email := normalizeEmail(input.Email)
	if email == "" {
		return response.BadRequestError("email is required")
	}

	opCtx, cancel := context.WithTimeout(ctx, externalResourceTimeout)
	defer cancel()

	exists, err := s.authRepo.EmailExists(opCtx, email)
	if err != nil {
		return response.InternalServerError("failed to check email")
	}
	if exists {
		return response.NewAppErr(http.StatusConflict, "email already registered")
	}

	token, err := GenerateSecureToken(tokenByteSize)
	if err != nil {
		return response.InternalServerError("failed to generate verification token")
	}

	if err := s.cache.Set(opCtx, RegisterVerificationKey(token), email, verificationTokenTTL); err != nil {
		return response.InternalServerError("failed to store verification token")
	}

	link := s.verificationLink(token)
	if err := s.mail.SendVerifyLink(opCtx, email, link); err != nil {
		return response.InternalServerError("failed to send verification link")
	}

	return nil
}

func (s *authService) VerifyEmail(ctx context.Context, token string) (string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", response.BadRequestError("token is required")
	}

	opCtx, cancel := context.WithTimeout(ctx, externalResourceTimeout)
	defer cancel()

	email, err := s.cache.GetDel(opCtx, RegisterVerificationKey(token))
	if err != nil {
		if cache.IsMiss(err) {
			return "", response.BadRequestError("verification token is invalid or expired")
		}
		return "", response.InternalServerError("failed to validate verification token")
	}

	session, err := GenerateSecureToken(tokenByteSize)
	if err != nil {
		return "", response.InternalServerError("failed to generate register session")
	}

	if err := s.cache.Set(opCtx, RegisterSessionKey(session), email, registerSessionTTL); err != nil {
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

	opCtx, cancel := context.WithTimeout(ctx, externalResourceTimeout)
	defer cancel()

	sessionEmail, err := s.cache.Get(opCtx, RegisterSessionKey(input.Session))
	if err != nil {
		if cache.IsMiss(err) {
			return nil, response.BadRequestError("register session is invalid or expired")
		}
		return nil, response.InternalServerError("failed to validate register session")
	}
	if sessionEmail != input.Email {
		return nil, response.ForbiddenError("register session email mismatch")
	}

	exists, err := s.authRepo.EmailExists(opCtx, input.Email)
	if err != nil {
		return nil, response.InternalServerError("failed to check email")
	}
	if exists {
		return nil, response.NewAppErr(http.StatusConflict, "email already registered")
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

	argonConfig := ArgonConfig{
		ID:          s.idGen.NewID(),
		UserID:      user.ID,
		Memory:      defaultArgonMemory,
		Iterations:  defaultArgonIterations,
		Parallelism: defaultArgonParallelism,
	}

	vault := Vault{
		ID:           s.idGen.NewID(),
		UserID:       user.ID,
		Timeout:      defaultVaultTimeout,
		TimeoutType:  defaultVaultTimeoutType,
		VaultKeySalt: input.Vault.VaultKeySalt,
	}

	if err := s.authRepo.WithTransaction(opCtx, func(repo AuthRepository) error {
		if err := repo.CreateUser(opCtx, user); err != nil {
			return err
		}

		if err := repo.CreateArgonConfig(opCtx, argonConfig); err != nil {
			return err
		}

		return repo.CreateVault(opCtx, vault)
	}); err != nil {
		return nil, response.InternalServerError("failed to register user")
	}

	if err := s.cache.Del(opCtx, RegisterSessionKey(input.Session)); err != nil {
		return nil, response.InternalServerError("failed to delete register session")
	}

	return &user, nil
}

func (s *authService) verificationLink(token string) string {
	baseURL := strings.TrimRight(s.appCfg.FrontendURL, "/")
	return fmt.Sprintf("%s/verify-email?token=%s", baseURL, url.QueryEscape(token))
}

func validateRegisterInput(input RegisterInput) error {
	if input.Session == "" {
		return response.BadRequestError("register session is required")
	}
	if input.Email == "" {
		return response.BadRequestError("email is required")
	}
	if input.Name == "" {
		return response.BadRequestError("name is required")
	}

	base64Fields := map[string]string{
		"authVerifier":     input.AuthVerifier,
		"authVerifierSalt": input.AuthVerifierSalt,
		"masterKeySalt":    input.MasterKeySalt,
		"vaultKeySalt":     input.Vault.VaultKeySalt,
	}
	for field, value := range base64Fields {
		if !isBase64(value) {
			return response.BadRequestError(fmt.Sprintf("%s must be a valid base64 string", field))
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
