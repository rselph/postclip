package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
)

const checkerSize = 20

var (
	important_ratios = []float64{
		4.0 / 5.0, 5.0 / 4.0,
		5.0 / 7.0, 7.0 / 5.0,
		8.5 / 11.0, 11.0 / 8.5,
		2.0 / 3.0, 3.0 / 2.0,
		9.0 / 16.0, 16.0 / 9.0,
		2.0, 0.5,
		1.0}

	important_sizes = []ImageSize{
		{"sm", image.Rect(0, 0, 600, 0)},
		{"sm", image.Rect(0, 0, 0, 315)},
		{"md", image.Rect(0, 0, 1200, 0)},
		{"md", image.Rect(0, 0, 0, 630)},
		{"lg", image.Rect(0, 0, 2400, 0)},
		{"lg", image.Rect(0, 0, 0, 1260)},
	}
)

type ImageSize struct {
	name   string
	bounds image.Rectangle
}

// Implements methods of image.Image to create a checkerboard pattern.
type CheckerBoard struct {
	checkSize int
	bounds    image.Rectangle
}

func main() {
	for _, size := range important_sizes {
		for _, ratio := range important_ratios {
			println(ratio)
			img, err := generateImage(size, ratio)
			if err != nil {
				panic(err)
			}
			err = saveImage(size.name, img)
			if err != nil {
				panic(err)
			}
		}
	}
}

// saveImage saves the generated image to disk with a filename based on the ratio.
func saveImage(name string, img image.Image) error {
	filename := fmt.Sprintf("test-%s-%04dx%04d.png", name, img.Bounds().Dx(), img.Bounds().Dy())

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	return png.Encode(out, img)
}

// generateImage creates a checkerboard image with the specified size and aspect ratio.
func generateImage(size ImageSize, ratio float64) (image.Image, error) {
	width := size.bounds.Dx()
	height := size.bounds.Dy()

	// Adjust width and height based on the desired aspect ratio.
	if width == 0.0 {
		size.bounds.Max.X = int(float64(height)*ratio) + size.bounds.Min.X
	} else if height == 0.0 {
		size.bounds.Max.Y = int(float64(width)/ratio) + size.bounds.Min.Y
	}

	return NewCheckerBoardImage(size.bounds, checkerSize), nil
}

// NewCheckerBoardImage creates a checkerboard pattern image of the specified size.
func NewCheckerBoardImage(bounds image.Rectangle, squareSize int) *CheckerBoard {
	return &CheckerBoard{checkSize: squareSize, bounds: bounds}
}

func (cb *CheckerBoard) ColorModel() color.Model {
	return color.RGBAModel
}

func (cb *CheckerBoard) Bounds() image.Rectangle {
	return cb.bounds
}

var colorA = &color.RGBA{200, 200, 200, 255} // Light gray
var colorB = &color.RGBA{100, 100, 100, 255} // Dark gray

func (cb *CheckerBoard) At(x, y int) color.Color {
	xSquare := (x / cb.checkSize) % 2
	ySquare := (y / cb.checkSize) % 2

	if xSquare == ySquare {
		return colorA
	}
	return colorB
}
