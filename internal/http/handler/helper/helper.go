package helper

import (
	"strconv"

	"lockbox/internal/constant"
	apphttp "lockbox/internal/http"
	"lockbox/internal/http/response"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page  int
	Limit int
}

func GetPagination(c *gin.Context) (Pagination, error) {
	const (
		defaultPage = 1
		defaulLimit = 10
		maxLimit    = 100
	)

	page := defaultPage
	limit := defaulLimit

	if pageQ := c.Query("page"); pageQ != "" {
		p, err := strconv.Atoi(pageQ)
		if err != nil {
			return Pagination{}, response.BadRequestError(apphttp.ErrTransport.Error(), "page must be a number")
		}
		page = p
	}

	if limitQ := c.Query("limit"); limitQ != "" {
		l, err := strconv.Atoi(limitQ)
		if err != nil {
			return Pagination{}, response.BadRequestError(apphttp.ErrTransport.Error(), "limit must be a number")
		}
		limit = l
	}

	if page < 1 {
		page = 1
	}

	if limit < 1 {
		limit = defaulLimit
	}

	if limit > maxLimit {
		limit = maxLimit
	}

	return Pagination{
		Page:  page,
		Limit: limit,
	}, nil
}

func ArchiveQuery(c *gin.Context) (bool, error) {
	archiveQ := c.Query("archive")
	if archiveQ == "" {
		return false, nil
	}

	archive, err := strconv.ParseBool(archiveQ)
	if err != nil {
		return false, response.BadRequestError(apphttp.ErrTransport.Error(), "archive must be a boolean")
	}

	return archive, nil
}

func FavoriteQuery(c *gin.Context) (bool, error) {
	isFavoriteQ := c.Query("favorite")
	if isFavoriteQ == "" {
		return false, nil
	}

	isFavorite, err := strconv.ParseBool(isFavoriteQ)
	if err != nil {
		return false, response.BadRequestError(apphttp.ErrTransport.Error(), "favorite must be a boolean")
	}

	return isFavorite, nil
}

func UserIDFromContext(c *gin.Context) (int64, error) {
	value, exists := c.Get(constant.UserIDCtx)
	if !exists {
		return 0, response.Unauthorized()
	}

	switch userID := value.(type) {
	case int64:
		return userID, nil
	case int:
		return int64(userID), nil
	case float64:
		return int64(userID), nil
	default:
		return 0, response.Unauthorized()
	}
}
