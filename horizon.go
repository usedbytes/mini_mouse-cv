// Copyright 2018 Brian Starkey <stark3y@gmail.com>
package cv

import (
	"fmt"
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

	if len(blobs) == 0 {
		return float32(math.NaN())
	}

	avgs := make([]uint8, 0, len(blobs))
	for _, b := range blobs {
		avgs = append(avgs, AverageDeltaC(img, b.First * scale, b.Second * scale))
	}
	meanAvg := Mean(avgs)

	for i := len(avgs) - 1; i >= 0; i-- {
		a := avgs[i]
		b := blobs[i]
		if a >= meanAvg {
			return float32((b.First + b.Second + 1) / 2) / float32(len(summed.Pix))
		}
	}

	return float32(math.NaN())
}

func FindHorizonROI(img image.Image, roi image.Rectangle) float32 {
	diff := DeltaCByRowROI(img, roi)
	minMax := MinMaxColwise(diff)
	ExpandContrastColWise(diff, minMax)
	Threshold(diff, 128)

	summed := FindHorizontalLines(diff)
	minMax = MinMaxColwise(summed)
	ExpandContrastColWise(summed, minMax)
	Threshold(summed, 128)

	blobs := FindBlobs(summed.Pix)
	scale := roi.Dy() / len(summed.Pix)

	if len(blobs) == 0 {
		return float32(math.NaN())
	}

	avgs := make([]uint8, 0, len(blobs))
	for _, b := range blobs {
		avgs = append(avgs, AverageDeltaCROI(img, b.First * scale, b.Second * scale, roi))
	}
	meanAvg := Mean(avgs)

	fmt.Println(meanAvg, avgs)

	for i := len(avgs) - 1; i >= 0; i-- {
		a := avgs[i]
		b := blobs[i]
		if a >= meanAvg {
			return float32((b.First + b.Second + 1) / 2) / float32(len(summed.Pix))
		}
	}

	return float32(math.NaN())
}
