package server

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"

	"github.com/miaoledor/lolicount/internal/drawer/cardthemedrawer"
)

// queryParams is the validated query contract for GET /@:name.
type queryParams struct {
	Theme    string  `query:"theme"    validate:"omitempty,themename|eq=random"`
	Mode     string  `query:"mode"     validate:"omitempty,eq=seq|eq=random"`
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
		q.Theme = cardthemedrawer.DefaultTheme
	}
}
