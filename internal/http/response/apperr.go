package response

import (
	"net/http"
	"strings"
	"time"
)

type AppErrResponse struct {
	Status    int       `json:"status"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (a AppErrResponse) Error() string {
	return a.Message
}

func NewAppErr(status int, code, msg string) *AppErrResponse {
	return &AppErrResponse{
		Status:    status,
		Code:      code,
		Message:   msg,
		Timestamp: time.Now().UTC(),
	}
}

func InternalServerError(msg string) *AppErrResponse {
	if strings.TrimSpace(msg) == "" {
		msg = "internal server error"
	}

	return &AppErrResponse{
		Status:    http.StatusInternalServerError,
		Code:      "INTERNAL_SERVER_ERROR",
		Message:   msg,
		Timestamp: time.Now().UTC(),
	}
}

func BadRequestError(code, msg string) *AppErrResponse {
	return &AppErrResponse{
		Status:    http.StatusBadRequest,
		Code:      code,
		Message:   msg,
		Timestamp: time.Now().UTC(),
	}
}

func NotFoundError(code, msg string) *AppErrResponse {
	return &AppErrResponse{
		Status:    http.StatusNotFound,
		Code:      code,
		Message:   msg,
		Timestamp: time.Now().UTC(),
	}
}

func Unauthorized() *AppErrResponse {
	return &AppErrResponse{
		Status:    http.StatusUnauthorized,
		Code:      "AUTH_UNAUTHORIZED",
		Message:   "unauthorized",
		Timestamp: time.Now().UTC(),
	}
}

func ForbiddenError(msg string) *AppErrResponse {
	return &AppErrResponse{
		Status:    http.StatusForbidden,
		Code:      "FORBIDDEN_ERROR",
		Message:   msg,
		Timestamp: time.Now().UTC(),
	}
}
