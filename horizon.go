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
	Threshold(summed, 128)

	blobs := FindBlobs(summed.Pix)
	scale := img.Bounds().Dy() / len(summed.Pix)

	avgs := make([]uint8, 0, len(blobs))
	for _, b := range blobs {
		avgs = append(avgs, AverageDeltaC(img, b.First * scale, b.Second * scale))
	}
	meanAvg := Mean(avgs)

	for i := len(avgs) - 1; i >= 0; i-- {
		a := avgs[i]
		b := blobs[i]
		if a >= meanAvg {
			return float32((b.First + b.Second) / 2) / float32(len(summed.Pix))
		}
	}

	return float32(math.NaN())
}
