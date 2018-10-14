package font

import (
	"image"
	//"math"
	//"fmt"
	
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"github.com/golang/freetype/truetype"
)


type Font struct {
	Texture []uint8
	Glyphs map[rune]Glyph
	RowHeight float64
	TopPad float64
	font truetype.Font
	scale float64
}

type Glyph struct {
	X            int
	Y            int
	BitmapWidth  int
	BitmapHeight int
	XOffset      float64
	YOffset      float64
	Width        float64
	Height   	 float64
	Advance      float64
}

func GetFont(bytes []uint8, size float64, scale float64) Font {
	ttfData, err := truetype.Parse(bytes)
	if err != nil {
		panic(err)
	}

	face := truetype.NewFace(ttfData, &truetype.Options{Size: size, Hinting: font.HintingFull, DPI: 72 * scale})
	ascent := face.Metrics().Ascent.Ceil()
	descent := -face.Metrics().Descent.Ceil()
	//emptySpace := (ascent - descent) - face.Metrics().Height.Ceil()
	//ascent -= emptySpace / 2
	//descent -= emptySpace / 2
	var font Font
	rowHeight := face.Metrics().Height.Ceil()
	font.Glyphs = make(map[rune]Glyph)
	font.scale = scale
	texSize := 1024
	font.Texture = make([]uint8, texSize*texSize)
	x, y := 2, 2
	for c := 32; c < 128; c++ {
		dr, img, imgOffset, advanceFixed, ok := face.Glyph(fixed.Point26_6{0, 0}, rune(c))
		if !ok {
			panic("Cannot load glyph")
		}

		advance := advanceFixed.Round()
		//fmt.Println(imgOffset, dr)
		xOffset := dr.Min.X
		yOffset := ascent + dr.Min.Y
		pixels := img.(*image.Alpha).Pix
		bitmapStride := img.(*image.Alpha).Stride
		
		bitmapWidth, bitmapHeight := dr.Max.X-dr.Min.X, dr.Max.Y-dr.Min.Y
		
		if x+bitmapWidth > texSize {
			x = 2
			y += rowHeight
		}

		font.Glyphs[rune(c)] = Glyph{
			x, y,
			bitmapWidth, bitmapHeight,
			float64(xOffset) / scale, float64(yOffset) / scale,
			float64(bitmapWidth) / scale, float64(bitmapHeight) / scale,
			float64(advance) / scale,
		}

		bitmapX, bitmapY := imgOffset.X, imgOffset.Y

		for yLoc := 0; yLoc < bitmapHeight; yLoc++ {
			sourcePixelsPos := bitmapX + (bitmapY+yLoc)*bitmapStride
			targetPixelsPos := x + (y+yLoc)*texSize
			copy(font.Texture[targetPixelsPos:targetPixelsPos+bitmapWidth], pixels[sourcePixelsPos:sourcePixelsPos+bitmapWidth])
		}

		x += bitmapWidth + 1
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

	font.TopPad = 0//float64((ascent - descent) - rowHeight) / scale * 0.5
	//font.TopPad = float64((ascent - descent) - rowHeight) / scale * 0.5
	//font.RowHeight = float64(rowHeight) / scale
	font.RowHeight = float64(ascent - descent) / scale

	return font
}

func (font *Font)GetStringWidth(text string) float64 {
	width := 0.0

	for i := 0; i < len(text); i++ {
		c := rune(text[i])
        glyph := font.Glyphs[c]
		width += glyph.Advance

		if i < len(text) - 1 {
			width += font.GetKerning(c, rune(text[i + 1]))
		}
	}

	return width
}

func (font *Font)GetStringFit(text string, fitToWidth float64) int {
	width := 0.0

	i := 0
	for ; i < len(text); i++ {
		c := rune(text[i])
        glyph := font.Glyphs[c]
		width += float64(glyph.Advance)
		
		if width > fitToWidth {
			return i
		}
		if i < len(text) - 1 {
			width += float64(font.GetKerning(c, rune(text[i + 1])))
		}
	}

	return i
}

func (font *Font)GetStringFitReverse(text string, fitToWidth float64) int {
	width := 0.0

	i := len(text) - 1
	for ; i >= 0; i-- {
		c := rune(text[i])
        glyph := font.Glyphs[c]
		width += float64(glyph.Advance)
		
		if width > fitToWidth {
			return i + 1
		}
		if i > 0 {
			width += float64(font.GetKerning(c, rune(text[i - 1])))
		}
	}

	return i + 1
}


func (font *Font) GetStringHeight() float64 {
	return font.RowHeight
}

func (font *Font) GetKerning(c1 rune, c2 rune) float64 {
	i1, i2 := font.font.Index(c1), font.font.Index(c2)
	return float64(font.font.Kern(fixed.Int26_6(100), i1, i2).Round()) / font.scale
}
