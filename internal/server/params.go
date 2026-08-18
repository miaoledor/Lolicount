package server

import (
	"fmt"
	"math/rand"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/ftheme"
	"github.com/miaoledor/lolicount/internal/theme"
)

// queryParams is the validated query contract for GET /@:name.
//
// M5.5: theme IS the background; there is no separate bg concept. The
// counter value is rendered as <text> overlaid on the theme frame.
// M5.6: scale controls the image display size (uniform base when 0);
// unshowf hides the counter text.
type queryParams struct {
	Theme    string  `query:"theme"    validate:"omitempty,themename|eq=random"`
	Number   int64   `query:"number"   validate:"omitempty,gte=0,lte=999999"`
	FSize    int     `query:"fsize"    validate:"omitempty,gte=0,lte=500"`
	Scale    float64 `query:"scale"    validate:"omitempty,gte=0.1,lte=4"`
	UnshowF  bool    `query:"unshowf"`
	FTheme   string  `query:"ftheme"   validate:"omitempty,themename|eq=random"`
	X        int     `query:"x"        validate:"omitempty,gte=-500,lte=2000"`
	Y        int     `query:"y"        validate:"omitempty,gte=-500,lte=2000"`
	RX       float64 `query:"rx"       validate:"omitempty,gte=0,lte=1"`
	RY       float64 `query:"ry"       validate:"omitempty,gte=0,lte=1"`
}

var queryValidator = validator.New()

func init() {
	// Theme names may contain hyphens (e.g. lian-st); alphanum alone
	// rejects them. Allow letters, digits and hyphens. The registry
	// lookup is the final authority, so a non-existent name still 400s.
	if err := queryValidator.RegisterValidation("themename", func(fl validator.FieldLevel) bool {
		for _, r := range fl.Field().String() {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
				continue
			}
			return false
		}
		return true
	}); err != nil {
		panic(err)
	}
}

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
		q.Theme = theme.DefaultTheme
	}
}

// resolveFTheme returns the font style for name, handling the reserved
// "random" value by picking from the registry. Empty name yields the
// zero Style, which the renderer maps to its defaults.
func resolveFTheme(reg ftheme.Registry, name string) (ftheme.Style, error) {
	if name == "" {
		return ftheme.Style{}, nil
	}
	if name == "random" {
		list := reg.List()
		if len(list) == 0 {
			return ftheme.Style{}, fmt.Errorf("no f-themes available for random")
		}
		st, ok := reg.Get(list[rand.Intn(len(list))])
		if !ok {
			return ftheme.Style{}, fmt.Errorf("random f-theme %q missing", list)
		}
		return st, nil
	}
	st, ok := reg.Get(name)
	if !ok {
		return ftheme.Style{}, fmt.Errorf("f-theme %q not found", name)
	}
	return st, nil
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
