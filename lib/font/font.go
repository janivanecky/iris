package font

import (
	"github.com/golang/freetype/truetype"
	"image"
	"golang.org/x/image/math/fixed"
	"golang.org/x/image/font"
)


type Font struct {
	Texture []uint8
	Glyphs map[rune]Glyph
	RowHeight int
	TopPad int
	_font truetype.Font
}

type Glyph struct {
	X            int
	Y            int
	BitmapWidth  int
	BitmapHeight int
	XOffset      int
	YOffset      int
	Advance      int
}

func GetFont(bytes []uint8, size float64) Font {
	ttfData, err := truetype.Parse(bytes)
	if err != nil {
		panic(err)
	}

	face := truetype.NewFace(ttfData, &truetype.Options{Size: size, Hinting: font.HintingFull})
	ascent := face.Metrics().Ascent.Round()
	descent := -face.Metrics().Descent.Round()
	var font Font
	font.RowHeight = face.Metrics().Height.Ceil()
	font.TopPad = font.RowHeight - (ascent - descent)
	font.Glyphs = make(map[rune]Glyph)
	texSize := 512
	font.Texture = make([]uint8, texSize*texSize)
	x, y := 0, 0
	for c := 32; c < 128; c++ {
		dr, img, imgOffset, advanceFixed, ok := face.Glyph(fixed.Point26_6{0, 0}, rune(c))
		if !ok {
			panic("Cannot load glyph")
		}

		advance := advanceFixed.Round()
		xOffset := dr.Min.X
		yOffset := ascent + dr.Min.Y

		pixels := img.(*image.Alpha).Pix
		bitmapStride := img.(*image.Alpha).Stride
		
		bitmapWidth, bitmapHeight := dr.Max.X-dr.Min.X, dr.Max.Y-dr.Min.Y
		
		if x+bitmapWidth > texSize {
			x = 0
			y += font.RowHeight
		}

		font.Glyphs[rune(c)] = Glyph{x, y, bitmapWidth, bitmapHeight, xOffset, yOffset, advance}

		bitmapX, bitmapY := imgOffset.X, imgOffset.Y

		for yLoc := 0; yLoc < bitmapHeight; yLoc++ {
			sourcePixelsPos := bitmapX + (bitmapY+yLoc)*bitmapStride
			targetPixelsPos := x + (y+yLoc)*texSize
			copy(font.Texture[targetPixelsPos:targetPixelsPos+bitmapWidth], pixels[sourcePixelsPos:sourcePixelsPos+bitmapWidth])
		}

		x += bitmapWidth
	}

	if texSize%2 == 1 {
		panic("ERROR")
	}

	for source_row, target_row := 0, texSize - 1; source_row < texSize/2; source_row, target_row = source_row+1, target_row-1 {
		temp_pixels := make([]uint8, texSize)
		copy(temp_pixels, font.Texture[source_row*texSize:(source_row+1)*texSize])
		copy(font.Texture[source_row*texSize:(source_row+1)*texSize], font.Texture[target_row*texSize:(target_row+1)*texSize])
		copy(font.Texture[target_row*texSize:(target_row+1)*texSize], temp_pixels)
	}

	return font
}

func (font *Font)GetStringWidth(text string) float64 {
	width := 0.0

	for i := 0; i < len(text); i++ {
		c := rune(text[i])
        glyph := font.Glyphs[c]
		width += float64(glyph.Advance)

		if i < len(text) - 1 {
			width += font.GetKerning(c, rune(text[i + 1]))
		}
	}

	return width
}

func (font *Font) GetKerning(c1 rune, c2 rune) float64 {
	i1, i2 := font._font.Index(c1), font._font.Index(c2)
	return float64(font._font.Kern(fixed.Int26_6(0), i1, i2).Round())
}
