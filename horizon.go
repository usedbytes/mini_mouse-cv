// Copyright 2018 Brian Starkey <stark3y@gmail.com>
package cv

import (
	"image"
	"math"
)

func FindHorizon(img image.Image) float32 {
	diff := DeltaCByRow(img)
	minMax := MinMaxColwise(diff)
	ExpandContrastColWise(diff, minMax)
	Threshold(diff, 128)

	summed := FindHorizontalLines(diff)
	minMax = MinMaxColwise(summed)
	ExpandContrastColWise(summed, minMax)
	Threshold(summed, 180)

	h := summed.Bounds().Dy()
	for y := h - 1; y >= 0; y-- {
		g := summed.GrayAt(0, y).Y
		if g > 0 {
			return float32(h - y) / float32(h)
		}
	}

	return float32(math.NaN())
}
