package render

import "github.com/miaoledor/lolicount/internal/imgcore"

// Kind returns LayerRandomPick.
func (l *RandomPickLayer) Kind() imgcore.LayerKind { return imgcore.LayerRandomPick }

// ZIndex returns the stack order.
func (l *RandomPickLayer) ZIndex() int { return l.Z }

// Fixed reports whether the layer cannot be deleted.
func (l *RandomPickLayer) Fixed() bool { return l.IsFixed }

// Render uses the PRNG to pick one ImageOption by weight, then delegates
// to that option's ImageLayer.Render. The selected option inherits the
// RandomPickLayer's Transform (merged with the option's own Transform)
// so the whole category can be positioned/scaled as a unit.
//
// This replaces characterthemedrawer's Assemble (which iterated
// hardcoded part categories and picked from index ranges) with a single
// weighted random pick over self-describing candidates.
func (l *RandomPickLayer) Render(ctx imgcore.RenderCtx) imgcore.LayerOutput {
	if len(l.Options) == 0 {
		return imgcore.LayerOutput{}
	}

	weights := make([]float64, len(l.Options))
	for i, opt := range l.Options {
		w := opt.Weight
		if w <= 0 {
			w = 1
		}
		weights[i] = w
	}
	idx := ctx.PRNG.WeightedPick(weights)

	picked := l.Options[idx]
	// Merge the layer-level transform into the picked option's transform
	// so both apply. The layer transform is applied first (outer), then
	// the option's own transform (inner).
	merged := picked.ImageLayer
	merged.Transform = mergeTransforms(l.Transform, picked.ImageLayer.Transform)
	return merged.Render(ctx)
}

// mergeTransforms combines two Transforms into one. The outer transform
// is applied first (translate/scale/rotate), then the inner. For
// simplicity, ranges are taken from the outer when non-default,
// otherwise from the inner.
func mergeTransforms(outer, inner imgcore.Transform) imgcore.Transform {
	result := inner
	if outer.X.Min != 0 || outer.X.Max != 0 {
		result.X = outer.X
	}
	if outer.Y.Min != 0 || outer.Y.Max != 0 {
		result.Y = outer.Y
	}
	if outer.Scale.Min != 1 || outer.Scale.Max != 1 {
		result.Scale = outer.Scale
	}
	if outer.Rotation.Min != 0 || outer.Rotation.Max != 0 {
		result.Rotation = outer.Rotation
	}
	return result
}
