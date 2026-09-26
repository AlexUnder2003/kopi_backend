package handlers

import (
	"errors"
	"net/http"

	"KopiBackend/internal/apperrors"

	"github.com/labstack/echo/v5"
)

const (
	invalidRequestBody = "invalid_request_body"
	unauthorized       = "unauthorized"
	invalidUUID        = "invalid_uuid"
)

func mapAppError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, apperrors.ErrBadRequest):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, apperrors.ErrUnauthorized):
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	case errors.Is(err, apperrors.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	case errors.Is(err, apperrors.ErrConflict):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
}
