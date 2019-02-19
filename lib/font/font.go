package font

import (
	"image"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"github.com/golang/freetype/truetype"
)

// Font contains data necessary for rendering of text with specific font.
type Font struct {
	Texture []uint8
	TextureWidth, TextureHeight int
	Glyphs map[rune]Glyph
	RowHeight float64
	font truetype.Font
	scale float64
}

// Glyph encapsulates size and bitmap parameters for a single character.
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

const texSize = 1024
const paddingX, paddingY = 2, 2
const bitmapSpacing = 1

// GetFont returns initialized Font object.
// bytes should contain contents of TTF file, size is size of font in pixels
// and dpiScale should reflect OS UI scaling.
func GetFont(bytes []uint8, size float64, dpiScale float64) Font {
	// Parse TTF file bytes into TrueType.Font.
	ttfData, err := truetype.Parse(bytes)
	if err != nil {
		panic(err)
	}

	// Create truetype Face, representing font at specific size.
	// 72 DPI is consider to be DPI at normal (100%) scale.
	faceOptions := truetype.Options{Size: size, Hinting: font.HintingFull, DPI: 72 * dpiScale}
	face := truetype.NewFace(ttfData, &faceOptions)

	// Get vertical measurments of the face.
	ascent := face.Metrics().Ascent.Ceil()
	descent := -face.Metrics().Descent.Ceil()
	height := face.Metrics().Height.Ceil()

	// Initialize Font structure.
	var font Font
	font.Glyphs = make(map[rune]Glyph)
	font.RowHeight = float64(ascent - descent) / dpiScale
	font.scale = dpiScale
	font.Texture = make([]uint8, texSize * texSize)
	font.TextureWidth = texSize
	font.TextureHeight = texSize

	// For each character we're going to store Glyph information and its bitmap
	// into font's texture. We start from SPACE character (ASCII code 32) and we
	// end with ~ character (ASCII code 126).
	x, y := paddingY, paddingY
	for c := 32; c < 127; c++ {
		// Get Glyph's bitmap and advance (advance is in fixed point).
		imgRect, img, imgOffset, advanceFxPt, ok := face.Glyph(fixed.Point26_6{0, 0}, rune(c))
		if !ok {
			panic("Cannot load glyph")
		}

		// Get Glyph's bitmap params.
		bitmapOffsetX, bitmapOffsetY := imgOffset.X, imgOffset.Y
		bitmapWidth, bitmapHeight := imgRect.Max.X - imgRect.Min.X, imgRect.Max.Y - imgRect.Min.Y
		bitmapStride := img.(*image.Alpha).Stride

		// Get current character's size parameters adjusted for DPI scaling.
		advance := float64(advanceFxPt.Round()) / dpiScale
		charOffsetX := float64(imgRect.Min.X) / dpiScale
		charOffsetY := float64(ascent + imgRect.Min.Y) / dpiScale
		charWidth := float64(bitmapWidth) / dpiScale
		charHeight := float64(bitmapHeight) / dpiScale
		
		// Get raw pixels of character's bitmap.
		pixels := img.(*image.Alpha).Pix
		
		// If this glyph's bitmap would go over the current texture width,
		// move it to the next row.
		if x + bitmapWidth > texSize {
			x = paddingX
			y += height
		}

		// Create and store Glyph structure for current character.
		font.Glyphs[rune(c)] = Glyph{
			x, y,
			bitmapWidth, bitmapHeight,
			charOffsetX, charOffsetY,
			charWidth, charHeight,
			advance,
		}

		// Copy character bitmap row by row into font's texture.
		for row := 0; row < bitmapHeight; row++ {
			sourcePixelsPos := bitmapOffsetX + (bitmapOffsetY + row) * bitmapStride
			targetPixelsPos := x + (y + row) * texSize
			copy(font.Texture[targetPixelsPos:targetPixelsPos + bitmapWidth], pixels[sourcePixelsPos:sourcePixelsPos + bitmapWidth])
		}

		// Move in the texture.
		x += bitmapWidth + bitmapSpacing
	}

	// Lastly we need to revert the order of rows in the font texture, because OpenGL requires bottom-up ordering.
	tempPixels := make([]uint8, texSize)
	for sourceRow, targetRow := 0, texSize - 1; sourceRow < texSize/2; sourceRow, targetRow = sourceRow+1, targetRow-1 {
		// Get start-end indices for rows being swapped.
		sourceIndexStart := sourceRow * texSize
		sourceIndexEnd := sourceIndexStart + texSize
		targetIndexStart := targetRow * texSize
		targetIndexEnd := targetIndexStart + texSize

		// Swap.
		copy(tempPixels, font.Texture[sourceIndexStart:sourceIndexEnd])
		copy(font.Texture[sourceIndexStart:sourceIndexEnd], font.Texture[targetIndexStart:targetIndexEnd])
		copy(font.Texture[targetIndexStart:targetIndexEnd], tempPixels)
	}

	return font
}

// GetStringWidth returns string's width in pixels (in DPI scaled space).
func (font *Font)GetStringWidth(text string) float64 {
	width := 0.0

	for i := 0; i < len(text); i++ {
		// Width of string is increased by how much we have to move
		// cursor after drawin the character - "advance".
		c := rune(text[i])
        glyph := font.Glyphs[c]
		width += glyph.Advance

		// Add kerning to string with.
		if i < len(text) - 1 {
			cNext := rune(text[i + 1])
			width += font.GetKerning(c, cNext)
		}
	}

	return width
}

// TODO: remove(?)
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

// TODO: remove(?)
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

// TODO: remove(?)
func (font *Font) GetStringHeight() float64 {
	return font.RowHeight
}

// GetKerning returns kerning between two characters in DPI scaled pixels.
func (font *Font) GetKerning(c1 rune, c2 rune) float64 {
	i1, i2 := font.font.Index(c1), font.font.Index(c2)
	kerningFxPt := font.font.Kern(fixed.Int26_6(100), i1, i2)
	kerningRaw := float64(kerningFxPt.Round())
	return kerningRaw / font.scale
}
