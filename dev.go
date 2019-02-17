package cv

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
)

func RunAlgorithm(in, out image.Image, profile bool) image.Image {
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
	if !profile {
		fmt.Println(summed.Pix)
	}

	// Hopefully we're left with exactly two blobs, marking the edges
	// TODO: Should handle 1 (one edge only) and 0 (full FoV filled) blob too
	blobs := FindBlobs(summed.Pix)
	scale := in.Bounds().Dx() / len(summed.Pix)

	if !profile {
		fmt.Println("Blobs", blobs)
	}

	var target = Tuple{ 0, in.Bounds().Dx() }
	if len(blobs) == 2 {
		target = Tuple{
			RoundUp((blobs[0].First + blobs[0].Second) * scale / 2, 2),
			RoundUp((blobs[1].First + blobs[1].Second) * scale / 2, 2),
		}
	}

	if !profile {
		fmt.Println("target", target)
	}

	targetColor := in.(*image.YCbCr).YCbCrAt((target.First + target.Second) / 2, in.Bounds().Dy() / 2)
	if !profile {
		fmt.Println("targetColor", targetColor)
	}

	//horz := FindHorizonROI(in, image.Rect(target.First, 0, target.Second, in.Bounds().Dy()))
	horz := float32(math.NaN())
	roi := image.Rect(target.First, 0, target.Second, in.Bounds().Dy())
	{
		diff := DeltaCByRowROI(in, roi)
		minMax := MinMaxColwise(diff)
		ExpandContrastColWise(diff, minMax)
		Threshold(diff, 128)

		return diff

		summed := FindHorizontalLines(diff)
		minMax = MinMaxColwise(summed)
		ExpandContrastColWise(summed, minMax)
		Threshold(summed, 128)


		blobs := FindBlobs(summed.Pix)
		scale := roi.Dy() / len(summed.Pix)

		if len(blobs) == 0 {
			horz = float32(math.NaN())
			goto eh
		}

		avgs := make([]uint8, 0, len(blobs))
		for _, b := range blobs {
			avgs = append(avgs, AverageDeltaCROIConst(in, b.First * scale, targetColor, roi))
		}

		if !profile {
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
		if !profile {
			fmt.Println("Blob", minIdx, "at", b)
		}
		horz = float32((b.First + b.Second + 1) / 2) / float32(len(summed.Pix))
	}
eh:

	if !profile && out != nil {
		bottom := out.Bounds().Dy()

		if float64(horz) != math.NaN() {
			bottom = int(horz * float32(out.Bounds().Dy()))
		}

		red := &image.Uniform{color.RGBA{0x80, 0, 0, 0x80}}

		rect := image.Rect(target.First, 0, target.Second, bottom)
		draw.Draw(out.(draw.Image), rect, red, image.ZP, draw.Over)
	}

	//summed := FindVerticalLines(diff)
	//minMax = MinMaxRowwise(summed)
	//ExpandContrastRowWise(summed, minMax)
	//Threshold(summed, 150)

	return nil
}
