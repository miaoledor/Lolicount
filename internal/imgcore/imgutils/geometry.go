package imgutils

// DefaultDisplaySize is the longest-edge target (in px) every frame is
// scaled to when no explicit Scale is given (M5.6: all images show at a
// consistent size). Frames smaller than this are scaled up to it as well
// so output is uniform across themes.
const DefaultDisplaySize = 400

// DisplaySize returns the target longest-edge display size for a frame.
// When scale is 0 the uniform base size is used (M5.6); otherwise the
// base is multiplied by scale.
func DisplaySize(scale float64) int {
	if scale <= 0 {
		return DefaultDisplaySize
	}
	return int(float64(DefaultDisplaySize) * scale)
}

// ScaledDims computes the displayed width/height of a source image,
// preserving aspect ratio so the longest edge equals the display size.
// Scaling (not stretching) keeps the image undistorted (M5.6).
func ScaledDims(srcW, srcH, display int) (int, int) {
	if srcW <= 0 || srcH <= 0 || display <= 0 {
		return srcW, srcH
	}
	longest := srcW
	if srcH > longest {
		longest = srcH
	}
	if longest == 0 {
		return srcW, srcH
	}
	ratio := float64(display) / float64(longest)
	w := int(float64(srcW) * ratio)
	h := int(float64(srcH) * ratio)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return w, h
}
