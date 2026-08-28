package main

import (
	"os"
	"reflect"
	"testing"
)

func TestIsBodyLayer(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		// Recognized body/dress layer names across all characters.
		{"斜め", true},
		{"正面", true},
		{"斜め腕上", true},
		{"斜め腕下", true},

		// Eye, mouth, face, and gaze-direction names are not body layers.
		{"0", false},
		{"1", false},
		{"2", false},
		{"目", false},
		{"目／しょんぼり", false},
		{"口／通常", false},
		{"頬／通常", false},
		{"", false},
	}
	for _, c := range cases {
		got := isBodyLayer(c.name)
		if got != c.want {
			t.Errorf("isBodyLayer(%q) = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestBoundingBox(t *testing.T) {
	// Empty input returns a zero rect.
	t.Run("empty", func(t *testing.T) {
		got := boundingBox(nil)
		want := cropRect{}
		if got != want {
			t.Errorf("boundingBox(nil) = %+v, want %+v", got, want)
		}
	})

	// Single layer: bbox equals the layer rect.
	t.Run("single", func(t *testing.T) {
		layers := []layerRow{{left: 10, top: 20, width: 100, height: 200}}
		got := boundingBox(layers)
		want := cropRect{Left: 10, Top: 20, Width: 100, Height: 200}
		if got != want {
			t.Errorf("boundingBox = %+v, want %+v", got, want)
		}
	})

	// Multiple layers: bbox is the union, including negative offsets.
	t.Run("union", func(t *testing.T) {
		layers := []layerRow{
			{left: 100, top: 50, width: 300, height: 400},
			{left: 0, top: 200, width: 150, height: 100},
			{left: 500, top: 10, width: 50, height: 600},
		}
		got := boundingBox(layers)
		// union: left=0, top=10, right=550, bottom=610
		want := cropRect{Left: 0, Top: 10, Width: 550, Height: 600}
		if got != want {
			t.Errorf("boundingBox = %+v, want %+v", got, want)
		}
	})

	// Overlapping layers with negative coordinates.
	t.Run("negative_offsets", func(t *testing.T) {
		layers := []layerRow{
			{left: -120, top: 300, width: 240, height: 150},
			{left: -140, top: 240, width: 280, height: 120},
		}
		got := boundingBox(layers)
		// union: left=-140, top=240, right=140, bottom=450
		want := cropRect{Left: -140, Top: 240, Width: 280, Height: 210}
		if got != want {
			t.Errorf("boundingBox = %+v, want %+v", got, want)
		}
	})
}

func TestDecodeUTF16LE(t *testing.T) {
	cases := []struct {
		name string
		raw  []byte
		want string
	}{
		// ASCII "layer\t0" with BOM.
		{
			"ascii_with_bom",
			[]byte{0xFF, 0xFE, 'l', 0, 'a', 0, 'y', 0, 'e', 0, 'r', 0, '\t', 0, '0', 0},
			"layer\t0",
		},
		// ASCII without BOM.
		{
			"ascii_no_bom",
			[]byte{'#', 0, '0', 0},
			"#0",
		},
		// Japanese 斜め (U+659C U+3081) in UTF-16LE.
		{
			"japanese",
			[]byte{0xFF, 0xFE, 0x9C, 0x65, 0x81, 0x30},
			"斜め",
		},
		// Embedded null bytes are stripped.
		{
			"null_stripped",
			[]byte{'a', 0, 0, 0, 'b', 0},
			"ab",
		},
		// Empty input.
		{
			"empty",
			nil,
			"",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := decodeUTF16LE(c.raw)
			if got != c.want {
				t.Errorf("decodeUTF16LE = %q, want %q", got, c.want)
			}
		})
	}
}

func TestParseLayerInfo(t *testing.T) {
	// Build a minimal UTF-16LE TJS layer-info file matching the real
	// fgimage format: header row, canvas-dims row, then two type-0 rows.
	// Fields: layer_type name left top width height type opacity visible
	//         layer_id group_layer_id base images
	lines := []string{
		"#layer_type\tname\tleft\ttop\twidth\theight\ttype\topacity\tvisible\tlayer_id\tgroup_layer_id\tbase\timages\t",
		"\t\t\t\t2362\t4134\t\t\t\t\t\t\t\t",
		"0\t斜め\t116\t348\t2213\t3641\t13\t255\t1\t101\t102\t\t\t",
		"0\t目\t997\t396\t383\t250\t13\t255\t1\t3036\t3038\t\t\t",
	}
	content := ""
	for _, l := range lines {
		content += l + "\n"
	}
	// Encode as UTF-16LE with BOM.
	raw := []byte{0xFF, 0xFE}
	for _, r := range content {
		raw = append(raw, byte(r), byte(r>>8))
	}

	tmp := t.TempDir() + "/test_0.txt"
	if err := writeFileBytes(tmp, raw); err != nil {
		t.Fatalf("write tmp file: %v", err)
	}

	rows, canvasW, canvasH, err := parseLayerInfo(tmp)
	if err != nil {
		t.Fatalf("parseLayerInfo: %v", err)
	}
	if canvasW != 2362 || canvasH != 4134 {
		t.Errorf("canvas = %dx%d, want 2362x4134", canvasW, canvasH)
	}
	// Both type-0 rows should be parsed.
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	want := []layerRow{
		{layerType: "0", name: "斜め", left: 116, top: 348, width: 2213, height: 3641, visible: 1, layerID: 101, groupID: 102},
		{layerType: "0", name: "目", left: 997, top: 396, width: 383, height: 250, visible: 1, layerID: 3036, groupID: 3038},
	}
	if !reflect.DeepEqual(rows, want) {
		t.Errorf("rows = %+v, want %+v", rows, want)
	}
}

// writeFileBytes writes raw bytes to a file (os.WriteFile wrapper).
func writeFileBytes(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
