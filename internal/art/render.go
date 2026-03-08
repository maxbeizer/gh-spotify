package art

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
)

// RenderFromURL downloads an image from a URL and converts it to colored ASCII art.
func RenderFromURL(imageURL string, width int) (string, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("downloading image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("downloading image: HTTP %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", fmt.Errorf("decoding image: %w", err)
	}

	return RenderFromImage(img, width), nil
}

// RenderFromImage converts an image.Image to colored ASCII art using half-block
// characters. Each output character cell encodes two vertical pixels: the top
// pixel as the foreground (▄ lower-half block) and the bottom pixel as the
// background, yielding double vertical resolution.
func RenderFromImage(img image.Image, width int) string {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if srcW == 0 || srcH == 0 || width <= 0 {
		return ""
	}

	// Terminal characters are roughly twice as tall as wide, so each character
	// cell represents 2 vertical pixels via the half-block technique.
	height := (width * srcH) / (srcW * 2)
	if height < 1 {
		height = 1
	}
	// Ensure even row count so we can pair rows for half-block rendering.
	pixelRows := height * 2

	var buf strings.Builder
	for row := 0; row < pixelRows; row += 2 {
		for col := 0; col < width; col++ {
			topColor := sampleColor(img, bounds, srcW, srcH, col, row, width, pixelRows)
			botColor := sampleColor(img, bounds, srcW, srcH, col, row+1, width, pixelRows)

			// Use "▄" (lower half block): foreground = bottom pixel, background = top pixel.
			tr, tg, tb := topColor.R, topColor.G, topColor.B
			br, bg, bb := botColor.R, botColor.G, botColor.B
			buf.WriteString(fmt.Sprintf("\033[48;2;%d;%d;%dm\033[38;2;%d;%d;%dm▄",
				tr, tg, tb, br, bg, bb))
		}
		buf.WriteString("\033[0m\n")
	}

	return buf.String()
}

// sampleColor maps a cell coordinate in the output grid to the source image
// and returns the pixel color at that position.
func sampleColor(img image.Image, bounds image.Rectangle, srcW, srcH, col, row, outW, outH int) color.RGBA {
	sx := bounds.Min.X + (col*srcW)/outW
	sy := bounds.Min.Y + (row*srcH)/outH
	if sx >= bounds.Max.X {
		sx = bounds.Max.X - 1
	}
	if sy >= bounds.Max.Y {
		sy = bounds.Max.Y - 1
	}

	r, g, b, _ := img.At(sx, sy).RGBA()
	return color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: 255}
}
