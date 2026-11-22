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

// RenderFromImage converts an image.Image to colored block art. Each output
// cell is a full block (█) colored with the average of the source pixels it
// covers. Only the foreground color is used: some terminals and multiplexers
// drop truecolor backgrounds, which made half-block rendering show dark bands.
// Rows assume terminal cells are roughly twice as tall as wide. The result has
// no trailing newline so it composes cleanly with lipgloss layouts.
func RenderFromImage(img image.Image, width int) string {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	if srcW == 0 || srcH == 0 || width <= 0 {
		return ""
	}

	height := (width * srcH) / (srcW * 2)
	if height < 1 {
		height = 1
	}

	rows := make([]string, height)
	for row := 0; row < height; row++ {
		var buf strings.Builder
		for col := 0; col < width; col++ {
			c := averageColor(img, bounds, col, row, width, height)
			fmt.Fprintf(&buf, "\033[38;2;%d;%d;%dm█", c.R, c.G, c.B)
		}
		buf.WriteString("\033[0m")
		rows[row] = buf.String()
	}

	return strings.Join(rows, "\n")
}

// averageColor returns the mean color of the source region covered by output
// cell (col, row) in an outW x outH grid.
func averageColor(img image.Image, bounds image.Rectangle, col, row, outW, outH int) color.RGBA {
	srcW, srcH := bounds.Dx(), bounds.Dy()
	x0 := bounds.Min.X + col*srcW/outW
	x1 := bounds.Min.X + (col+1)*srcW/outW
	y0 := bounds.Min.Y + row*srcH/outH
	y1 := bounds.Min.Y + (row+1)*srcH/outH
	if x1 <= x0 {
		x1 = x0 + 1
	}
	if y1 <= y0 {
		y1 = y0 + 1
	}

	var rs, gs, bs, n uint64
	for y := y0; y < y1 && y < bounds.Max.Y; y++ {
		for x := x0; x < x1 && x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			rs += uint64(r >> 8)
			gs += uint64(g >> 8)
			bs += uint64(b >> 8)
			n++
		}
	}
	if n == 0 {
		return color.RGBA{A: 255}
	}
	return color.RGBA{R: uint8(rs / n), G: uint8(gs / n), B: uint8(bs / n), A: 255}
}
