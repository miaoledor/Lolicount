package cardthemedrawer

import (
	"strings"
	"testing"
)

// Draw renders a frame as a data-URI <image>, scaled to the uniform
// display size with aspect ratio preserved (M5.6).
func TestDrawUniformDisplaySize(t *testing.T) {
	frame := Frame{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}
	layer := Draw(frame, 0) // scale=0 -> default display size 400
	// 10x20 frame, longest edge 20 -> 400, ratio 20 -> 200x400
	if !strings.Contains(layer.Fragment, `width="200" height="400"`) {
		t.Errorf("image not scaled to uniform size: %s", layer.Fragment)
	}
	if layer.Width != 200 || layer.Height != 400 {
		t.Errorf("layer dims = %dx%d, want 200x400", layer.Width, layer.Height)
	}
}

func TestDrawScaleMultipliesSize(t *testing.T) {
	frame := Frame{Width: 10, Height: 20, Data: "data:image/gif;base64,QQ"}
	layer := Draw(frame, 2) // 400*2=800; 10x20 -> 400x800
	if !strings.Contains(layer.Fragment, `width="400" height="800"`) {
		t.Errorf("scale=2 should double display size: %s", layer.Fragment)
	}
}

func TestDrawAspectPreserved(t *testing.T) {
	frame := Frame{Width: 2000, Height: 1000, Data: "data:image/gif;base64,QQ"}
	layer := Draw(frame, 0) // longest 2000 -> 400, ratio 0.2 -> 400x200
	if !strings.Contains(layer.Fragment, `width="400" height="200"`) {
		t.Errorf("wide frame aspect not preserved: %s", layer.Fragment)
	}
}

func TestDrawContainsDataURI(t *testing.T) {
	frame := Frame{Width: 10, Height: 20, Data: "data:image/gif;base64,ABC"}
	layer := Draw(frame, 0)
	if !strings.Contains(layer.Fragment, frame.Data) {
		t.Errorf("data URI missing from fragment: %s", layer.Fragment)
	}
}
