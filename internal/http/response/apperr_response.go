package response

import (
	"net/http"
	"strings"
	"time"
)

type AppErrResponse struct {
	Status    int       `json:"status"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (a AppErrResponse) Error() string {
	return a.Message
}

func NewAppErr(status int, msg string) *AppErrResponse {
	return &AppErrResponse{
		Status:    status,
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
		Message:   msg,
		Timestamp: time.Now().UTC(),
	}
}

func BadRequestError(msg string) *AppErrResponse {
	return &AppErrResponse{
		Status:    http.StatusBadRequest,
		Message:   msg,
		Timestamp: time.Now().UTC(),
	}
}

func NotFoundError(msg string) *AppErrResponse {
	return &AppErrResponse{
		Status:    http.StatusNotFound,
		Message:   msg,
		Timestamp: time.Now().UTC(),
	}
}

func Unauthorized() *AppErrResponse {
	return &AppErrResponse{
		Status:    http.StatusUnauthorized,
		Message:   "unauthorized",
		Timestamp: time.Now().UTC(),
	}
}

func ForbiddenError(msg string) *AppErrResponse {
	return &AppErrResponse{
		Status:    http.StatusForbidden,
		Message:   msg,
		Timestamp: time.Now().UTC(),
	}
}
