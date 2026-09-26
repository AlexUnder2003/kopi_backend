package handlers

import (
	"net/http"

	"KopiBackend/internal/dto"
	"KopiBackend/internal/middleware"
	"KopiBackend/internal/models"
	"KopiBackend/internal/services"
	"KopiBackend/internal/utils"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

type CategoryHandler struct {
	categoryService *services.CategoryService
}

func NewCategoryHandler(categoryService *services.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

func (h *CategoryHandler) Register(g *echo.Group) {
	categories := g.Group("/categories")
	categories.Use(middleware.UserMiddleware)

	categories.POST("", h.Create)
	categories.GET("", h.List)
	categories.GET("/:id", h.GetByID)
	categories.PATCH("/:id", h.Update)
	categories.DELETE("/:id", h.Delete)
}

func (h *CategoryHandler) Create(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	var req dto.CategoryPost
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}
	if err := req.Validate(); err != nil {
		return mapCategoryValidationErrors(err)
	}

	resp, err := h.categoryService.Create(c.Request().Context(), &models.Category{
		Name:   req.Name,
		Icon:   req.Icon,
		UserID: &userID,
	})
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *CategoryHandler) List(c *echo.Context) error {
	userID, err := utils.GetUserIDFromContext(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	resp, err := h.categoryService.List(c.Request().Context(), userID)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *CategoryHandler) GetByID(c *echo.Context) error {
	if _, err := utils.GetUserIDFromContext(c); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	resp, err := h.categoryService.GetByID(c.Request().Context(), id)
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *CategoryHandler) Update(c *echo.Context) error {
	if _, err := utils.GetUserIDFromContext(c); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	var req dto.CategoryUpdate
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidRequestBody)
	}

	resp, err := h.categoryService.Update(c.Request().Context(), &models.Category{
		ID:   id,
		Name: req.Name,
		Icon: req.Icon,
	})
	if err != nil {
		return mapAppError(err)
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *CategoryHandler) Delete(c *echo.Context) error {
	if _, err := utils.GetUserIDFromContext(c); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, unauthorized)
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, invalidUUID)
	}

	if err := h.categoryService.Delete(c.Request().Context(), id); err != nil {
		return mapAppError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

func mapCategoryValidationErrors(err error) error {
	if err.Error() == "name" {
		return echo.NewHTTPError(http.StatusBadRequest, "bad_request_category_name")
	}
	return err
}
