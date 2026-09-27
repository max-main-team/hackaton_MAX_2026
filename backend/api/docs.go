// Package api держит OpenAPI-контракт и Swagger UI.
// Контракт — источник правды для фронта; правится вместе с кодом.
package api

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

//go:embed openapi.yaml
var openapiYAML string

//go:embed swagger.html
var swaggerHTML string

func OpenAPIYAML(c echo.Context) error {
	return c.Blob(http.StatusOK, "application/yaml; charset=utf-8", []byte(openapiYAML))
}

func SwaggerUI(c echo.Context) error {
	return c.HTML(http.StatusOK, swaggerHTML)
}
