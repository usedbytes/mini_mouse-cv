package cv

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

var debug = false

func FindBoard(in image.Image, c color.Color, roi image.Rectangle) (left, right, bottom float32) {
	left, right, bottom = 0.0, 1.0, 1.0

	// Find and amplify edges
	diff := DeltaCByCol(in)
	minMax := MinMaxRowwise(diff)
	ExpandContrastRowWise(diff, minMax)
	Threshold(diff, 128)

	// Attempt to ignore noisy rows (likely above/below the target)
	w, h := ImageDims(diff)
	for y := 0; y < h; y++ {
		row := diff.Pix[y * diff.Stride : y * diff.Stride + w]
		blobs := FindBlobs(row)
		if len(blobs) != 2 {
			for x := 0; x < len(row); x++ {
				row[x] = 0
			}
		}
	}

	// Find vertical lines in the non-noisy bits
	summed := FindVerticalLines(diff)
	minMax = MinMaxRowwise(summed)
	ExpandContrastRowWise(summed, minMax)
	Threshold(summed, 128)

	if debug {
		fmt.Println(summed.Pix)
	}

	// Hopefully we're left with exactly two blobs, marking the edges
	// TODO: Should handle 1 (one edge only) and 0 (full FoV filled) blob too
	blobs := FindBlobs(summed.Pix)
	scale := in.Bounds().Dx() / len(summed.Pix)

	if debug {
		fmt.Println("Blobs", blobs)
	}

	var target = Tuple{ 0, in.Bounds().Dx() }
	if len(blobs) == 2 {
		target = Tuple{
			RoundUp((blobs[0].First + blobs[0].Second) * scale / 2, 2),
			RoundUp((blobs[1].First + blobs[1].Second) * scale / 2, 2),
		}
	}

	if debug {
		fmt.Println("target", target)
	}

	left = float32(target.First) / float32(in.Bounds().Dx())
	right = float32(target.Second) / float32(in.Bounds().Dx())

	// Only look for the bottom if we know what color we are after
	if c != nil {
		roi := image.Rect(target.First, 0, target.Second, in.Bounds().Dy())
		diff := DeltaCByRowROI(in, roi)
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
			return left, right, float32(math.NaN())
		}

		avgs := make([]uint8, 0, len(blobs))
		for _, b := range blobs {
			avgs = append(avgs, AverageDeltaCROIConst(in, b.First * scale, c, roi))
		}

		if debug {
			fmt.Println("Avgs:", avgs)
		}

		min := uint8(255)
		minIdx := -1
		for i, m := range avgs {
			if m < min {
				min = m
				minIdx = i
			}
		}

		b := blobs[minIdx]
		if debug {
			fmt.Println("Blob", minIdx, "at", b)
		}
		bottom = float32((b.First + b.Second + 1) / 2) / float32(len(summed.Pix))
	}

	return left, right, bottom
}
