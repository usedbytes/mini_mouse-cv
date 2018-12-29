package cv

import (
	"image"
	"image/color"
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
