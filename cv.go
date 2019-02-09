package cv

import (
	"image"
	"image/color"
	"math"
)

type Tuple struct {
	First, Second int
}

func FindBlobs(pix []uint8) []Tuple {
	ret := make([]Tuple, 0, 3)

	for i := 0; i < len(pix); i++ {
		v := pix[i]
		if v == 0 {
			continue
		}

		t := i
		if i > 0 {
			t = i - 1
		}

		ret = append(ret, Tuple{ First: t })

		if i < len(pix) - 1 {
			for j := i + 1; j < len(pix); j++ {
				v = pix[j]
				if v > 0 && j < len(pix) - 1 {
					continue
				}

				ret[len(ret) - 1].Second = j
				i = j + 1
				break
			}
		} else {
			ret[len(ret) - 1].Second = i
		}
	}

	return ret
}

func Mean(vals []uint8) uint8 {
	a := 0
	for _, v := range vals {
		a += int(v)
	}
	return uint8(a / len(vals))
}

func AverageDeltaC(in image.Image, rowA, rowB int) uint8 {
	w, h := in.Bounds().Dx(), in.Bounds().Dy()

	total := 0

	switch v := in.(type) {
	case *image.YCbCr:
		hsub, vsub := 1, 1
		switch v.SubsampleRatio {
		case image.YCbCrSubsampleRatio422:
			hsub, vsub = 2, 1
		case image.YCbCrSubsampleRatio420:
			hsub, vsub = 2, 2
		}

		if rowB - rowA < vsub {
			if rowA > vsub {
				rowA -= vsub - 1
			} else if rowB < h - vsub {
				rowB += vsub - 1
			}
		}

		startYA := v.YOffset(0, rowA)
		startCA := v.COffset(0, rowA)
		startYB := v.YOffset(0, rowB)
		startCB := v.COffset(0, rowB)

		for x, subx := 0, 0; x < w; x, subx = x + hsub, subx + 1 {
			s := color.YCbCr{ Y: v.Y[startYA + x], Cb: v.Cb[startCA + subx], Cr: v.Cr[startCA + subx] }
			d := color.YCbCr{ Y: v.Y[startYB + x], Cb: v.Cb[startCB + subx], Cr: v.Cr[startCB + subx] }

			diff := DeltaC(s, d)
			total += int(diff)
		}
		total *= hsub
	default:
		_ = v
		for x := 0; x < w; x++ {
			diff := color.Gray{DeltaC(in.At(x, rowA), in.At(x, rowB))}
			total += int(diff.Y)
		}
	}

	return uint8(total / w)
}

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
			pix := img.Pix[j * img.Stride + i]
			if pix < min {
				min = pix
			}
			if pix > max {
				max = pix
			}
		}
		ret[i] = image.Pt(int(max), int(min))
	}

	return ret
}

func MinMaxRowwise(img *image.Gray) []image.Point {
	ret := make([]image.Point, img.Bounds().Dy())
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	cpp := 1

	for i := 0; i < h; i++ {
		var min, max uint8 = 255, 0
		row := img.Pix[img.Stride * i : img.Stride * i + w]
		for j := 0; j < w * cpp; j += cpp {
			if row[j] < min {
				min = row[j]
			}
			if row[j] > max {
				max = row[j]
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
			pix := img.Pix[j * img.Stride + i]
			newVal := float32(pix - uint8(minMax[i].Y)) * scale
			img.Pix[j * img.Stride + i] = uint8(newVal)
		}
	}
}

func ExpandContrastRowWise(img *image.Gray, minMax []image.Point) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	for i := 0; i < h; i++ {
		scale := 255.0 / float32(minMax[i].X - minMax[i].Y)
		for j := 0; j < w; j++ {
			pix := img.Pix[i * img.Stride + j]
			newVal := float32(pix - uint8(minMax[i].Y)) * scale
			img.Pix[i * img.Stride + j] = uint8(newVal)
		}
	}
}

func Threshold(img *image.Gray, threshold uint8) {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	for i := 0; i < h; i++ {
		row := img.Pix[i * img.Stride:i * img.Stride + w]
		for j := 0; j < w; j++ {
			pix := row[j]
			if pix >= threshold {
				row[j] = 255
			} else {
				row[j] = 0
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

func SumColumns(img *image.Gray) []int {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()

	sums := make([]int, w)
	for x := 0; x < w; x++ {
		sum := 0
		for y := 0; y < h; y++ {
			if img.GrayAt(x, y).Y > 0 {
				sum++
			}
		}
		sums[x] = sum
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

func DeltaCByRow(in image.Image) *image.Gray {
	w, h := in.Bounds().Dx(), in.Bounds().Dy()
	var out *image.Gray

	switch v := in.(type) {
	case *image.YCbCr:
		hsub, vsub := 1, 1
		switch v.SubsampleRatio {
		case image.YCbCrSubsampleRatio422:
			hsub, vsub = 2, 1
		case image.YCbCrSubsampleRatio420:
			hsub, vsub = 2, 2
		}
		cols, rows := w / hsub, h / vsub
		out = image.NewGray(image.Rect(0, 0, cols, rows - 1))
		for x, subx := 0, 0; x < w; x, subx = x + hsub, subx + 1 {
			for y, suby := 0, 0; y < h - vsub; y, suby = y + vsub, suby + 1 {
				yoff := v.YOffset(x, y)
				coff := v.COffset(x, y)
				s := color.YCbCr{ Y: v.Y[yoff], Cb: v.Cb[coff], Cr: v.Cr[coff] }

				yoff += v.YStride * vsub
				coff += v.CStride
				d := color.YCbCr{ Y: v.Y[yoff], Cb: v.Cb[coff], Cr: v.Cr[coff] }

				diff := DeltaCYCbCr(s, d)
				out.Pix[suby * out.Stride + subx] = diff
			}
		}
	default:
		_ = v
		out = image.NewGray(image.Rect(0, 0, w, h - 1))
		for x := 0; x < w; x++ {
			for y := 0; y < h - 1; y++ {
				diff := color.Gray{DeltaC(in.At(x, y), in.At(x, y + 1))}
				out.SetGray(x, y, diff)
			}
		}
	}

	return out
}

func DeltaCByCol(in image.Image) *image.Gray {
	w, h := in.Bounds().Dx(), in.Bounds().Dy()
	var out *image.Gray

	switch v := in.(type) {
	case *image.YCbCr:
		hsub, vsub := 1, 1
		switch v.SubsampleRatio {
		case image.YCbCrSubsampleRatio422:
			hsub, vsub = 2, 1
		case image.YCbCrSubsampleRatio420:
			hsub, vsub = 2, 2
		}
		cols, rows := w / hsub, h / vsub
		out = image.NewGray(image.Rect(0, 0, cols - 1, rows))

		for y, suby := 0, 0; y < h; y, suby = y + vsub, suby + 1 {
			for x, subx := 0, 0; x < w - hsub; x, subx = x + hsub, subx + 1 {
				yoff := v.YOffset(x, y)
				coff := v.COffset(x, y)

				s := color.YCbCr{ Y: v.Y[yoff], Cb: v.Cb[coff], Cr: v.Cr[coff] }
				d := color.YCbCr{ Y: v.Y[yoff + hsub], Cb: v.Cb[coff + 1], Cr: v.Cr[coff + 1] }
				diff := DeltaCYCbCr(s, d)
				out.Pix[suby * out.Stride + subx] = diff
			}
		}

	default:
		_ = v
		out = image.NewGray(image.Rect(0, 0, w, h - 1))
		for y := 0; y < h; y++ {
			for x := 0; x < w - 1; x++ {
				diff := color.Gray{DeltaC(in.At(x, y), in.At(x + 1, y))}
				out.SetGray(x, y, diff)
			}
		}
	}

	return out
}

func FindHorizontalLines(img *image.Gray) *image.Gray {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	stripeH := int(float64(h) / 16)
	scale := 255.0 / float64(w * stripeH)

	sums := SumLines(img)

	out := image.NewGray(image.Rect(0, 0, 1, h))
	for y := h - (stripeH / 2) - 1; y > (stripeH / 2) + 1; y-- {
		sum := 0
		for j := 0; j < stripeH; j++ {
			sum += sums[y + (stripeH / 2) - j]
		}

		out.Set(0, y, color.Gray{uint8(float64(sum) * scale)})
	}

	return out
}

func FindVerticalLines(img *image.Gray) *image.Gray {
	w := img.Bounds().Dx()
	stripeW := int(float64(w) / 16)
	scale := 255.0 / float64(w * stripeW)

	sums := SumColumns(img)

	out := image.NewGray(image.Rect(0, 0, w, 1))
	for x := (stripeW / 2); x < w - (stripeW / 2); x++ {
		sum := 0
		for j := 0; j < stripeW; j++ {
			sum += sums[x - (stripeW / 2) + j]
		}

		out.Set(x, 0, color.Gray{uint8(float64(sum) * scale)})
	}

	return out
}
