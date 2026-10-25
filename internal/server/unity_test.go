package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gofiber/fiber/v3"
)

func TestUnityModelNames(t *testing.T) {
	t.Parallel()
	s := &Server{unityFS: fstest.MapFS{
		"z-model/model.json": &fstest.MapFile{Data: []byte(`{}`)},
		"a-model/model.json": &fstest.MapFile{Data: []byte(`{}`)},
		"ignored/readme.txt": &fstest.MapFile{Data: []byte("no manifest")},
	}}
	got := s.unityModelNames()
	if len(got) != 2 || got[0] != "a-model" || got[1] != "z-model" {
		t.Fatalf("unityModelNames() = %v", got)
	}
}

func TestUnityModelHandler(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	s := &Server{unityFS: fstest.MapFS{
		"test-model/model.json": &fstest.MapFile{Data: []byte(`{"version":1}`)},
		"test-model/body.webp":  &fstest.MapFile{Data: []byte("webp")},
	}}
	app.Get("/unity/models/:name/:file", s.unityModelHandler)

	response, err := app.Test(httptest.NewRequest(http.MethodGet, "/unity/models/test-model/model.json", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if got := response.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
	if got := response.Header.Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}

	sprite, err := app.Test(httptest.NewRequest(http.MethodGet, "/unity/models/test-model/body.webp", nil))
	if err != nil {
		t.Fatal(err)
	}
	if got := sprite.Header.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("sprite Cache-Control = %q", got)
	}

	bad, err := app.Test(httptest.NewRequest(http.MethodGet, "/unity/models/test-model/not-allowed.bin", nil))
	if err != nil {
		t.Fatal(err)
	}
	if bad.StatusCode != fiber.StatusBadRequest {
		t.Fatalf("bad path status = %d", bad.StatusCode)
	}
}
