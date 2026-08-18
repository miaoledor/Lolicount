package server

import (
	"fmt"
	"math/rand"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/theme"
)

// queryParams is the validated query contract for GET /@:name.
//
// M5.5: theme IS the background; there is no separate bg concept. The
// counter value is rendered as <text> overlaid on the theme frame.
type queryParams struct {
	Theme  string  `query:"theme"  validate:"omitempty,alphanum|eq=random"`
	Number int64   `query:"number" validate:"omitempty,gte=0,lte=999999"`
	X      int     `query:"x"      validate:"omitempty,gte=-500,lte=2000"`
	Y      int     `query:"y"      validate:"omitempty,gte=-500,lte=2000"`
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
	if q.Scale == 0 {
		q.Scale = 1
	}
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
