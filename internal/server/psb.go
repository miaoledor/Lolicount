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

// Model file names inside assets/psb/<model>/. The gzip-compressed form
// is the preferred storage: PSB files are mostly raw RGBA texture data
// and compress roughly 8x, which keeps both the git repo and the
// transfer small. Browsers transparently decompress Content-Encoding:
// gzip responses before the bytes reach the widget's XHR, so the driver
// still receives pure PSB bytes.
const (
	psbModelFile    = "model.psb"
	psbModelGzFile  = "model.psb.gz"
	psbModelPattern = "^[a-zA-Z0-9-]+$"
)

// psbModelNameRe restricts model directory names to the same charset as
// theme names, which also blocks path traversal.
var psbModelNameRe = regexp.MustCompile(psbModelPattern)

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

// psbModelHandler answers GET /psb/:model with the model bytes. The bytes
// are fixed at build time (embed.FS), so the response is immutable.
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
	file := psbModelFileFor(s.psbFS, name)
	if file == "" {
		return fiber.NewError(fiber.StatusNotFound, "model not found")
	}
	data, err := fs.ReadFile(s.psbFS, name+"/"+file)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "model not found")
	}
	c.Set("Content-Type", "application/octet-stream")
	c.Set("Cache-Control", "public, max-age=31536000, immutable")
	c.Set("Access-Control-Allow-Origin", "*")
	if file == psbModelGzFile {
		c.Set("Content-Encoding", "gzip")
	}
	return c.Status(fiber.StatusOK).Send(data)
}

// psbModelFileFor returns the model file name present in the model
// directory — "model.psb.gz" when stored compressed (preferred), plain
// "model.psb" otherwise, or "" when the directory holds no model.
func psbModelFileFor(fsys fs.FS, name string) string {
	for _, f := range []string{psbModelGzFile, psbModelFile} {
		if _, err := fs.Stat(fsys, name+"/"+f); err == nil {
			return f
		}
	}
	return ""
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
		if psbModelFileFor(s.psbFS, e.Name()) != "" {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names
}
