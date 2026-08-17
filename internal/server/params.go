package server

import (
	"fmt"
	"math/rand"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/bg"

	"github.com/miaoledor/lolicount/internal/theme"
)

// queryParams is the validated query contract for GET /@:name.
//
// M2.5 base: theme + number. M5 adds background-overlay params: when bg
// is set, the response uses theme.RenderWithBg (background image + digit
// overlay); otherwise it falls back to the pure-digit Render.
type queryParams struct {
	Theme  string  `query:"theme"  validate:"omitempty,alphanum|eq=random"`
	Number int64   `query:"number" validate:"omitempty,gte=0,lte=999999"`
	BG     string  `query:"bg"     validate:"omitempty,excludesall=/ \\,excludes=@"`
	X      int     `query:"x"      validate:"omitempty,gte=-500,lte=2000"`
	Y      int     `query:"y"      validate:"omitempty,gte=-500,lte=2000"`
	Align  string  `query:"align"  validate:"omitempty,oneof=top center bottom"`
	FSize  int     `query:"fsize"  validate:"omitempty,gte=0,lte=500"`
	Scale  float64 `query:"scale"  validate:"omitempty,gte=0.1,lte=2"`
}

var queryValidator = validator.New()

// parseParams binds and validates the query string, then fills defaults.
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
	if q.Align == "" {
		q.Align = "top"
	}
	if q.Scale == 0 {
		q.Scale = 1
	}
	// X/Y default to 0 when bg is used without explicit position; the
	// caller can place the digit block anywhere via x/y.
}

// resolveTheme returns the theme to render with, handling the reserved
// "random" value by picking from the registry.
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

// resolveBackground returns the background to overlay, handling the
// reserved "random" value by picking from the registry.
func resolveBackground(reg bg.Registry, name string) (bg.Background, error) {
	if name == "random" {
		list := reg.List()
		if len(list) == 0 {
			return bg.Background{}, fmt.Errorf("no backgrounds available for random")
		}
		b, ok := reg.Get(list[rand.Intn(len(list))])
		if !ok {
			return bg.Background{}, fmt.Errorf("random background %q missing", list)
		}
		return b, nil
	}
	b, ok := reg.Get(name)
	if !ok {
		return bg.Background{}, fmt.Errorf("background %q not found", name)
	}
	return b, nil
}
