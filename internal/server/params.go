package server

import (
	"fmt"
	"math/rand"

	"github.com/gofiber/fiber/v3"
	"github.com/go-playground/validator/v10"

	"github.com/miaoledor/lolicount/internal/theme"
)

// queryParams is the validated query contract for GET /@:name.
// Defaults match docs/projectDesign.md so the frontend can skip params
// equal to the default when building URLs.
type queryParams struct {
	Theme     string  `query:"theme"     validate:"omitempty,alphanum|eq=random"`
	Bg        string  `query:"bg"        validate:"omitempty,alphanum"` // M5
	X         float64 `query:"x"         validate:"omitempty,gte=0"`
	Y         float64 `query:"y"         validate:"omitempty,gte=0"`
	FontSize  int     `query:"fsize"     validate:"omitempty,gte=8,lte=200"`
	Scale     float64 `query:"scale"     validate:"omitempty,gte=0.1,lte=2"`
	Align     string  `query:"align"     validate:"omitempty,oneof=top center bottom"`
	Padding   int     `query:"padding"   validate:"omitempty,gte=0,lte=16"`
	Offset    float64 `query:"offset"    validate:"omitempty,gte=-500,lte=500"`
	Pixelated string  `query:"pixelated" validate:"omitempty,oneof=0 1"`
	DarkMode  string  `query:"darkmode"  validate:"omitempty,oneof=0 1 auto"`
	Num       int64   `query:"num"       validate:"omitempty,gte=0,lte=1000000000000000"`
	Prefix    int64   `query:"prefix"    validate:"omitempty,gte=-1,lte=999999"`
}

var queryValidator = validator.New()

// parseParams binds and validates the query string, then fills defaults.
// It returns 400 on any constraint violation.
func parseParams(c fiber.Ctx) (*queryParams, error) {
	var q queryParams
	if err := c.Bind().Query(&q); err != nil {
		return nil, fmt.Errorf("bind query: %w", err)
	}
	if err := queryValidator.Struct(&q); err != nil {
		return nil, fmt.Errorf("validate query: %w", err)
	}
	q.applyDefaults()
	return &q, nil
}

// applyDefaults mirrors the documented default table.
func (q *queryParams) applyDefaults() {
	if q.Theme == "" {
		q.Theme = "loli"
	}
	if q.Scale == 0 {
		q.Scale = 1
	}
	if q.Align == "" {
		q.Align = "top"
	}
	if q.Padding == 0 {
		q.Padding = 7
	}
	if q.Pixelated == "" {
		q.Pixelated = "1"
	}
	if q.DarkMode == "" {
		q.DarkMode = "auto"
	}
	if q.Prefix == 0 {
		q.Prefix = -1
	}
}

// toRenderParams converts the validated query into theme.RenderParams.
func (q *queryParams) toRenderParams(count int64) theme.RenderParams {
	return theme.RenderParams{
		Count:     count,
		Padding:   q.Padding,
		Prefix:    q.Prefix,
		Offset:    q.Offset,
		Align:     q.Align,
		Scale:     q.Scale,
		FontSize:  q.FontSize,
		Pixelated: q.Pixelated,
		DarkMode:  q.DarkMode,
	}
}

// resolveTheme returns the theme to render with, handling the reserved
// "random" value by picking from the registry. "demo" is NOT a theme —
// it is a reserved counter name handled by the caller.
func resolveTheme(reg theme.Registry, name string) (*theme.Theme, error) {
	if name == "random" {
		list := reg.List()
		if len(list) == 0 {
			return nil, fmt.Errorf("no themes available for random")
		}
		t, ok := reg.Get(list[rand.Intn(len(list))])
		if !ok {
			return nil, fmt.Errorf("random theme %q missing", list)
		}
		return t, nil
	}
	t, ok := reg.Get(name)
	if !ok {
		return nil, fmt.Errorf("theme %q not found", name)
	}
	return t, nil
}

