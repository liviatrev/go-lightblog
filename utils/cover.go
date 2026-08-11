// utils/cover.go
package utils

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go-lightblog/config"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	coverWidth      = 1200
	coverHeight     = 630
	coverMaxLines   = 4
	coverTextMargin = 80 // horizontal padding in pixels
)

// GenerateCoverImage creates a cover image from the default-cover background
// with the post title overlaid on it, then uploads it through the same
// pipeline used for manual uploads (local storage or ImageKit CDN).
// The filename is derived from the post slug: cover-<slug>.jpg
func GenerateCoverImage(slug, title string) (string, error) {
	// 1. Load the default cover background
	defaultCoverPath := filepath.Join(config.PublicPath, "assets", "default-cover.jpg")
	img, err := imaging.Open(defaultCoverPath, imaging.AutoOrientation(true))
	if err != nil {
		return "", fmt.Errorf("failed to open default cover: %v", err)
	}

	// 2. Resize/crop to a consistent 1200x630 canvas
	img = imaging.Fill(img, coverWidth, coverHeight, imaging.Center, imaging.Lanczos)

	// 3. Draw the post title
	img = drawTitle(img, title)

	// 4. Encode to JPEG in memory
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 88}); err != nil {
		return "", fmt.Errorf("failed to encode cover image: %v", err)
	}

	// 5. Build the filename from the post slug
	filename := fmt.Sprintf("cover-%s.jpg", slug)
	if filename == "cover-.jpg" {
		filename = fmt.Sprintf("cover-%d.jpg", time.Now().Unix())
	}

	// 6. Upload through the shared storage pipeline
	return UploadImageFile(filename, &buf)
}

// drawTitle overlays the post title onto the cover with a semi-transparent
// dark scrim for contrast. The text is wrapped to fit the canvas and centered.
// It returns a new *image.NRGBA canvas with the title drawn on it.
func drawTitle(img image.Image, title string) *image.NRGBA {
	// Work on an NRGBA canvas so we can draw the scrim and text onto it.
	canvas := image.NewNRGBA(img.Bounds())
	draw.Draw(canvas, canvas.Bounds(), img, image.Point{}, draw.Src)

	// Semi-transparent dark scrim to keep the text readable
	scrim := image.NewNRGBA(canvas.Bounds())
	draw.Draw(scrim, scrim.Bounds(), &image.Uniform{color.NRGBA{R: 0, G: 0, B: 0, A: 130}}, image.Point{}, draw.Src)
	draw.Draw(canvas, canvas.Bounds(), scrim, image.Point{}, draw.Over)

	face, ok := loadBoldFace(72)
	if !ok {
		face = basicfont.Face7x13
	}
	defer func() {
		if c, ok := face.(interface{ Close() error }); ok {
			c.Close()
		}
	}()

	drawWrappedTitle(canvas, title, face, coverMaxLines)
	return canvas
}

// drawWrappedTitle renders the title centered on the canvas, wrapping words to
// fit the available width and truncating the last line with an ellipsis if it
// exceeds maxLines.
func drawWrappedTitle(img *image.NRGBA, title string, face font.Face, maxLines int) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.White),
		Face: face,
	}

	maxWidth := float64(img.Bounds().Dx()) - 2*coverTextMargin
	lines := wrapText(title, face, maxWidth)

	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines[maxLines-1] = truncateWithEllipsis(lines[maxLines-1], face, maxWidth)
	}

	metrics := face.Metrics()
	lineHeight := metrics.Height

	totalHeight := lineHeight * fixed.Int26_6(len(lines))
	startY := (fixed.I(img.Bounds().Dy()) - totalHeight) / 2

	for i, line := range lines {
		lineWidth := d.MeasureString(line)
		x := (fixed.I(img.Bounds().Dx()) - lineWidth) / 2
		y := startY + lineHeight*fixed.Int26_6(i) + metrics.Ascent

		d.Dot = fixed.Point26_6{X: x, Y: y}
		d.DrawString(line)
	}
}

// wrapText splits text into lines that fit within maxWidth pixels.
func wrapText(text string, face font.Face, maxWidth float64) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	d := &font.Drawer{Face: face}
	maxWidthFixed := fixed.Int26_6(maxWidth * 64)

	var lines []string
	current := words[0]

	for _, word := range words[1:] {
		testLine := current + " " + word
		if d.MeasureString(testLine) > maxWidthFixed {
			lines = append(lines, current)
			current = word
		} else {
			current = testLine
		}
	}
	lines = append(lines, current)

	return lines
}

// truncateWithEllipsis shortens a line with a trailing ellipsis so it fits
// within maxWidth pixels.
func truncateWithEllipsis(line string, face font.Face, maxWidth float64) string {
	d := &font.Drawer{Face: face}
	maxWidthFixed := fixed.Int26_6(maxWidth * 64)

	if d.MeasureString(line) <= maxWidthFixed {
		return line
	}

	const ellipsis = "…"
	runes := []rune(line)
	for i := len(runes) - 1; i > 0; i-- {
		candidate := string(runes[:i]) + ellipsis
		if d.MeasureString(candidate) <= maxWidthFixed {
			return candidate
		}
	}
	return ellipsis
}

// loadBoldFace attempts to load a bold system font face, returning nil if no
// usable system font is found (callers should fall back to basicfont).
func loadBoldFace(size float64) (font.Face, bool) {
	candidates := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Bold.ttf",
		"/System/Library/Fonts/Supplemental/Arial Bold.ttf",
		"/System/Library/Fonts/Helvetica.ttc",
		"C:/Windows/Fonts/arialbd.ttf",
		"C:/Windows/Fonts/segoeuib.ttf",
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		f, err := opentype.Parse(data)
		if err != nil {
			continue
		}
		face, err := opentype.NewFace(f, &opentype.FaceOptions{
			Size:    size,
			DPI:     72,
			Hinting: font.HintingFull,
		})
		if err != nil {
			continue
		}
		return face, true
	}

	return nil, false
}