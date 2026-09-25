package routes

import (
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Handler interface {
	Register(g *echo.Group)
}

func Router(handlers []Handler) *echo.Echo {
	e := echo.New()
	g := e.Group("/api")
	g.Use(middleware.RequestLogger())

	for _, handler := range handlers {
		handler.Register(g)
	}

	return e
}
