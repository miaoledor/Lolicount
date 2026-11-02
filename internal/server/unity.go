package server

import (
	"io/fs"
	"regexp"
	"sort"

	"github.com/gofiber/fiber/v3"
)

var unityModelNameRe = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
var unityFileNameRe = regexp.MustCompile(`^(model\.json|[a-zA-Z0-9_-]+\.webp)$`)

// listUnityModels returns browser-ready Unity UGUI models embedded at build time.
func (s *Server) listUnityModels(c fiber.Ctx) error {
	names := s.unityModelNames()
	if names == nil {
		names = []string{}
	}
	c.Set("Cache-Control", "public, max-age=60")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"models": names})
}

// unityModelHandler serves a converted model manifest or one of its sprites.
func (s *Server) unityModelHandler(c fiber.Ctx) error {
	name := c.Params("name")
	file := c.Params("file")
	if !unityModelNameRe.MatchString(name) || !unityFileNameRe.MatchString(file) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid model name")
	}
	if s.unityFS == nil {
		return fiber.NewError(fiber.StatusNotFound, "no models")
	}
	data, err := fs.ReadFile(s.unityFS, name+"/"+file)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "file not found")
	}
	if file == "model.json" {
		c.Set("Content-Type", "application/json")
		// Model URLs are stable across releases, so manifests must be revalidated.
		// Their sprite files are content-stable and can remain immutable below.
		c.Set("Cache-Control", "no-store")
	} else {
		c.Set("Content-Type", "image/webp")
		c.Set("Cache-Control", "public, max-age=31536000, immutable")
	}
	c.Set("Access-Control-Allow-Origin", "*")
	return c.Status(fiber.StatusOK).Send(data)
}

// unityModelNames lists directories containing a converted model manifest.
func (s *Server) unityModelNames() []string {
	if s.unityFS == nil {
		return nil
	}
	entries, err := fs.ReadDir(s.unityFS, ".")
	if err != nil {
		return nil
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || !unityModelNameRe.MatchString(entry.Name()) {
			continue
		}
		if _, err := fs.Stat(s.unityFS, entry.Name()+"/model.json"); err == nil {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names
}
