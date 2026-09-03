// Package server psb.go serves the embedded E-mote (PSB) models consumed
// by the emote widget (docs/emote-widget.md). Models are NOT imgcore
// themes: they are raw bytes rendered client-side by the WebGL driver, so
// they only need to be listed and streamed.
package server

import (
	"io/fs"
	"regexp"
	"sort"

	"github.com/gofiber/fiber/v3"
)

// psbModelFile is the fixed file name every emote model directory must
// contain to be exposed (assets/psb/<model>/model.psb).
const psbModelFile = "model.psb"

// psbModelNameRe restricts model directory names to the same charset as
// theme names, which also blocks path traversal.
var psbModelNameRe = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

// listPsbModels answers GET /api/psb/models with the embedded emote model
// names. Read-only and stable per build, so the short cache used by the
// other /api list endpoints applies.
func (s *Server) listPsbModels(c fiber.Ctx) error {
	names := s.psbModelNames()
	if names == nil {
		names = []string{}
	}
	c.Set("Cache-Control", "public, max-age=60")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"models": names})
}

// psbModelHandler answers GET /psb/:model with the raw model bytes. The
// bytes are fixed at build time (embed.FS), so the response is immutable.
// Swapping a model's contents therefore requires a new directory name (or
// clients will keep the stale copy for the cache lifetime). CORS is open:
// the bytes are public and the dev Nuxt server (and any tooling) loads
// them cross-origin via XHR.
func (s *Server) psbModelHandler(c fiber.Ctx) error {
	name := c.Params("model")
	if !psbModelNameRe.MatchString(name) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid model name")
	}
	if s.psbFS == nil {
		return fiber.NewError(fiber.StatusNotFound, "no models")
	}
	data, err := fs.ReadFile(s.psbFS, name+"/"+psbModelFile)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "model not found")
	}
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Cache-Control", "public, max-age=31536000, immutable")
	c.Set("Access-Control-Allow-Origin", "*")
	return c.Status(fiber.StatusOK).Send(data)
}

// psbModelNames returns the sorted list of model directories that contain
// a model file. An empty (or missing) tree yields an empty list, never an
// error — the widget simply shows its empty state.
func (s *Server) psbModelNames() []string {
	if s.psbFS == nil {
		return nil
	}
	entries, err := fs.ReadDir(s.psbFS, ".")
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if _, err := fs.Stat(s.psbFS, e.Name()+"/"+psbModelFile); err == nil {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}
