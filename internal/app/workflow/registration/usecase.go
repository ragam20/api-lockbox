package registration

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"lockbox/internal/app/auth"
	authapp "lockbox/internal/app/auth/application"
	authdomain "lockbox/internal/app/auth/domain"
	vaultdomain "lockbox/internal/app/vault/domain"

	"lockbox/internal/cache"
	dbtx "lockbox/internal/database/transaction"
	"lockbox/internal/http/response"
	"lockbox/internal/idgen"

	"gorm.io/gorm"
)

type Registration interface {
	Execute(ctx context.Context, input authapp.RegisterInput) (*authdomain.User, error)
}

type registration struct {
	txManager    dbtx.TransactionManager
	authRepo     authdomain.AuthRepository
	vaultRepo    vaultdomain.VaultRepository
	cache        cache.Store
	idGen        idgen.IDGenerator
	tokenManager auth.TokenManager
}

func NewRegistration(
	txManager dbtx.TransactionManager,
	authRepo authdomain.AuthRepository,
	vaultRepo vaultdomain.VaultRepository,
	cache cache.Store,
	idGen idgen.IDGenerator,
	tokenManager auth.TokenManager,
) Registration {
	return &registration{
		txManager:    txManager,
		authRepo:     authRepo,
		vaultRepo:    vaultRepo,
		cache:        cache,
		idGen:        idGen,
		tokenManager: tokenManager,
	}
}

func (s *registration) Execute(ctx context.Context, input authapp.RegisterInput) (*authdomain.User, error) {
	input.Email = normalizeEmail(input.Email)
	input.Name = strings.TrimSpace(input.Name)
	input.Session = strings.TrimSpace(input.Session)

	if err := validateRegisterInput(input); err != nil {
		return nil, err
	}

	sessionEmail, err := s.cache.Get(ctx, s.tokenManager.RegisterSessionKey(input.Session))
	if err != nil {
		if cache.IsMiss(err) {
			return nil, response.BadRequestError(auth.ErrSessionExpired.Error(), "register session is invalid or expired")
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
		return nil, response.NewAppErr(http.StatusConflict, auth.ErrEmailAlreadyExists.Error(), "email already registered")
	}

	user := authdomain.User{
		ID:               s.idGen.NewID(),
		Email:            input.Email,
		Name:             input.Name,
		MasterKeySalt:    input.MasterKeySalt,
		AuthVerifier:     input.AuthVerifier,
		AuthVerifierSalt: input.AuthVerifierSalt,
		EmailVerified:    true,
	}

	if err := s.txManager.WithinTransaction(ctx, func(tx *gorm.DB) error {
		authRepo := s.authRepo.WithTx(tx)
		vaultRepo := s.vaultRepo.WithTx(tx)

		if err := authRepo.CreateUser(ctx, user); err != nil {
			return auth.ErrFailedRegister
		}

		argonConfig := vaultdomain.NewDefaultArgonConfig(s.idGen.NewID())
		if err := vaultRepo.CreateArgonConfig(ctx, argonConfig, user.ID); err != nil {
			return auth.ErrFailedRegister
		}

		vaultData := vaultdomain.NewDefaultVault(s.idGen.NewID(), input.VaultKeySalt)
		if err := vaultRepo.CreateVault(ctx, vaultData, user.ID); err != nil {
			return auth.ErrFailedRegister
		}

		return nil
	}); err != nil {
		if errors.Is(err, auth.ErrFailedRegister) {
			return nil, response.BadRequestError(auth.ErrFailedRegister.Error(), "failed to register")
		}
		return nil, response.InternalServerError(err.Error())
	}

	_ = s.cache.Del(ctx, s.tokenManager.RegisterSessionKey(input.Session))

	return &user, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateRegisterInput(input authapp.RegisterInput) error {
	if input.Session == "" {
		return response.BadRequestError(auth.ErrSessionRequired.Error(), "register session is required")
	}
	if input.Email == "" {
		return response.BadRequestError(auth.ErrFieldEmailRequired.Error(), "email is required")
	}
	if input.Name == "" {
		return response.BadRequestError(auth.ErrFieldNameRequired.Error(), "name is required")
	}

	base64Fields := map[string]string{
		"authVerifier":     input.AuthVerifier,
		"authVerifierSalt": input.AuthVerifierSalt,
		"masterKeySalt":    input.MasterKeySalt,
		"vaultKeySalt":     input.VaultKeySalt,
	}
	for field, value := range base64Fields {
		if !isBase64(value) {
			return response.BadRequestError(auth.ErrValidation.Error(), fmt.Sprintf("%s must be a valid base64 string", field))
		}
	}

	return nil
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
