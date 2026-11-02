package main

import (
	"bytes"
	"image"
	"image/png"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"
)

// TestValidateCharacterThemeContract covers the manifest, range, and
// image relationships required by the runtime character-theme loader.
func TestValidateCharacterThemeContract(t *testing.T) {
	tests := []struct {
		name      string
		manifest  string
		config    string
		imageIDs  []int
		wantError string
	}{
		{
			name: "placeholder outside active range",
			manifest: `[{
				"width":1,"height":1,"layer_id":0
			},{
				"width":4,"height":5,"layer_id":1
			}]`,
			config:   `{"canvasW":100,"canvasH":200,"ranges":{"body":{"first":1,"last":1}}}`,
			imageIDs: []int{1},
		},
		{
			name:     "zero based range without placeholder",
			manifest: `[{"width":4,"height":5,"layer_id":1}]`,
			config:   `{"canvasW":100,"canvasH":200,"ranges":{"body":{"first":0,"last":0}}}`,
			imageIDs: []int{1},
		},
		{
			name:      "active zero ID",
			manifest:  `[{"width":4,"height":5,"layer_id":0}]`,
			config:    `{"canvasW":100,"canvasH":200,"ranges":{"body":{"first":0,"last":0}}}`,
			wantError: "must have a positive layer_id",
		},
		{
			name:      "range out of bounds",
			manifest:  `[{"width":4,"height":5,"layer_id":1}]`,
			config:    `{"canvasW":100,"canvasH":200,"ranges":{"body":{"first":0,"last":1}}}`,
			imageIDs:  []int{1},
			wantError: "outside ren.json indexes",
		},
		{
			name:     "overlapping ranges",
			manifest: `[{"width":4,"height":5,"layer_id":1}]`,
			config: `{"canvasW":100,"canvasH":200,"ranges":{
				"body":{"first":0,"last":0},"face":{"first":0,"last":0}
			}}`,
			imageIDs:  []int{1},
			wantError: "overlap at manifest index 0",
		},
		{
			name:     "duplicate active layer ID",
			manifest: `[{"width":4,"height":5,"layer_id":1},{"width":4,"height":5,"layer_id":1}]`,
			config: `{"canvasW":100,"canvasH":200,"ranges":{
				"body":{"first":0,"last":1}
			}}`,
			imageIDs:  []int{1},
			wantError: "is duplicated at manifest indexes",
		},
		{
			name:      "missing layer image",
			manifest:  `[{"width":4,"height":5,"layer_id":1}]`,
			config:    `{"canvasW":100,"canvasH":200,"ranges":{"body":{"first":0,"last":0}}}`,
			wantError: "has no matching ren/<layer_id> image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"hero/ren.json":    {Data: []byte(tt.manifest)},
				"hero/config.json": {Data: []byte(tt.config)},
			}
			for _, layerID := range tt.imageIDs {
				fsys["hero/ren/"+strconv.Itoa(layerID)+".png"] = &fstest.MapFile{Data: validTestPNG(t)}
			}

			report := validateTheme(fsys, "hero")
			joinedErrors := strings.Join(report.errors, "\n")
			if tt.wantError == "" && len(report.errors) > 0 {
				t.Fatalf("unexpected validation errors: %s", joinedErrors)
			}
			if tt.wantError != "" && !strings.Contains(joinedErrors, tt.wantError) {
				t.Fatalf("validation errors %q do not contain %q", joinedErrors, tt.wantError)
			}
		})
	}
}

func validTestPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 4, 5))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode test PNG: %v", err)
	}
	return buf.Bytes()
}
