package handlers

import (
	"net/http"
	"strconv"

	"KopiBackend/internal/repositories"

	"github.com/labstack/echo/v5"
)

func listParams(c *echo.Context) (repositories.ListParams, error) {
	params := repositories.ListParams{
		Filter: c.QueryParam("filter"),
	}

	if value := c.QueryParam("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 1 {
			return params, echo.NewHTTPError(http.StatusBadRequest, "bad_request_limit")
		}
		params.Limit = limit
	}

	if value := c.QueryParam("offset"); value != "" {
		offset, err := strconv.Atoi(value)
		if err != nil || offset < 0 {
			return params, echo.NewHTTPError(http.StatusBadRequest, "bad_request_offset")
		}
		params.Offset = offset
	}

	return params, nil
}
