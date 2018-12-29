package cv

import (
	"image"
	"image/color"
	"math"
)

type RawYCbCrColor struct {
	color.YCbCr
}

func expandCopyMSBs(col uint8) uint32 {
	return (uint32(col) << 8) | uint32(col)
}

func (c RawYCbCrColor) RGBA() (uint32, uint32, uint32, uint32) {
	return expandCopyMSBs(c.Y), expandCopyMSBs(c.Cb), expandCopyMSBs(c.Cr), 0xffff
}

type RawYCbCr struct {
	*image.YCbCr
}

func NewRawYCbCr(img *image.YCbCr) *RawYCbCr {
	return &RawYCbCr{ YCbCr: img }
}

func (r *RawYCbCr) At(x, y int) color.Color {
	pix := r.YCbCr.At(x, y).(color.YCbCr)
	return RawYCbCrColor{ YCbCr: pix }
}

func MinMaxColwise(img *image.Gray) []image.Point {
	ret := make([]image.Point, img.Bounds().Dx())
	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	for i := 0; i < w; i++ {
		var min, max uint8 = 255, 0
		for j := 0; j < h; j++ {
			pix := img.At(i, j).(color.Gray)
			if pix.Y < min {
				min = pix.Y
			}
			if pix.Y > max {
				max = pix.Y
			}
		}
		ret[i] = image.Pt(int(max), int(min))
	}

	return ret
}

func ExpandContrastColWise(img *image.Gray, minMax []image.Point) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	for i := 0; i < w; i++ {
		scale := 255.0 / float32(minMax[i].X - minMax[i].Y)

		for j := 0; j < h; j++ {
			pix := img.At(i, j).(color.Gray)
			newVal := float32(pix.Y - uint8(minMax[i].Y)) * scale
			img.Set(i, j, color.Gray{uint8(newVal)})
		}
	}
}

func Threshold(img *image.Gray, threshold uint8) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			pix := img.At(j, i).(color.Gray)
			if pix.Y >= threshold {
				img.Set(j, i, color.Gray{255})
			} else {
				img.Set(j, i, color.Gray{0})
			}
		}
	}
}

func SumLines(img *image.Gray) []int {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	sums := make([]int, h)
	for y := 0; y < h; y++ {
		sum := 0
		for x := 0; x < w; x++ {
			if img.GrayAt(x, y).Y > 0 {
				sum++
			}
		}
		sums[y] = sum
	}

	return sums
}

func absdiff_uint8(a, b uint8) int {
	if a < b {
		return int(b - a)
	} else {
		return int(a - b)
	}
}

func DeltaCNRGBA(a, b color.NRGBA) uint8 {
	deltaR := float64(absdiff_uint8(a.R, b.R))
	deltaG := float64(absdiff_uint8(a.G, b.G))
	deltaB := float64(absdiff_uint8(a.B, b.B))

	deltaC := math.Sqrt( (2 * deltaR * deltaR) +
			    (4 * deltaG * deltaG) +
			    (3 * deltaB * deltaB) +
			    (deltaR * ((deltaR * deltaR) - (deltaB * deltaB)) / 256.0))

	return uint8(deltaC)
}

func DeltaCYCbCr(a, b color.YCbCr) uint8 {
	// FIXME: What do?
	cbdiff := float64(absdiff_uint8(a.Cb, b.Cb)) / 255.0
	crdiff := float64(absdiff_uint8(a.Cr, b.Cr)) / 255.0
	return uint8(math.Sqrt(cbdiff * cbdiff + crdiff * crdiff) * 255.0)
}

func DeltaC(a, b color.Color) uint8 {
	// Try some special cases
	switch aV := a.(type) {
	case color.YCbCr:
		if bV, ok := b.(color.YCbCr); ok {
			return DeltaCYCbCr(aV, bV)
		}
	case color.NRGBA:
		if bV, ok := b.(color.NRGBA); ok {
			return DeltaCNRGBA(aV, bV)
		}
	}

	aR, aG, aB, aA := a.RGBA()
	bR, bG, bB, bA := b.RGBA()

	aNRGBA := color.NRGBA{
		R: uint8(float64(aR) * 255.0 / float64(aA)),
		G: uint8(float64(aG) * 255.0 / float64(aA)),
		B: uint8(float64(aB) * 255.0 / float64(aA)),
		A: uint8(aA >> 8),
	}
	bNRGBA := color.NRGBA{
		R: uint8(float64(bR) * 255.0 / float64(bA)),
		G: uint8(float64(bG) * 255.0 / float64(bA)),
		B: uint8(float64(bB) * 255.0 / float64(bA)),
		A: uint8(bA >> 8),
	}

	return DeltaCNRGBA(aNRGBA, bNRGBA)
}
