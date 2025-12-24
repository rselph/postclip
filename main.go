package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/nfnt/resize"
	_ "golang.org/x/image/tiff"
)

const fileSuffix = "_insta.jpg"

var (
	solidBackground image.Image
	blurBackground  bool
)

func main() {
	var (
		backgroundGray float64
		white          bool
		black          bool
		gray           bool
	)

	flag.Float64Var(&backgroundGray, "background", 1,
		"background gray, 0.0 to 1.0")
	flag.BoolVar(&white, "white", false, "white background")
	flag.BoolVar(&black, "black", false, "black background")
	flag.BoolVar(&gray, "gray", false, "gray background")
	flag.BoolVar(&blurBackground, "blur", false, "blurred background")
	flag.Parse()

	switch {
	case white:
		backgroundGray = 1
	case black:
		backgroundGray = 0
	case gray:
		backgroundGray = 0.125
	default:
		backgroundGray = 1
	}

	solidBackground = image.NewUniform(color.Gray{Y: uint8(backgroundGray * math.MaxUint8)})

	wg := sync.WaitGroup{}
	jobs := make(chan string)
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for fname := range jobs {
				doFile(fname)
			}
		}()
	}
	for _, filename := range flag.Args() {
		if !strings.HasSuffix(filename, fileSuffix) {
			jobs <- filename
		}
	}
	close(jobs)
	wg.Wait()
}

func doFile(fname string) {
	f, err := os.Open(fname)
	if err != nil {
		log.Print(err)
		return
	}

	original, _, err := image.Decode(f)
	_ = f.Close()
	if err != nil {
		log.Print(err)
		return
	}

	composite := doImage(original)
	outName := strings.TrimSuffix(fname, filepath.Ext(fname)) + fileSuffix
	o, err := os.Create(outName)
	if err != nil {
		log.Print(err)
		return
	}

	err = jpeg.Encode(o, composite, &jpeg.Options{Quality: 80})
	if err != nil {
		log.Print(err)
	}

	err = o.Close()
	if err != nil {
		log.Print(err)
	}
}

const (
	maxAspectRatio = 1080.0 / 566.0
	minAspectRatio = 1080.0 / 1350.0
)

func doImage(original image.Image) image.Image {
	inset := resize.Thumbnail(1080, 1350,
		original, resize.Lanczos3)

	var width, height int
	aspectRatio := float64(inset.Bounds().Dx()) / float64(inset.Bounds().Dy())

	switch {
	case aspectRatio >= minAspectRatio && aspectRatio <= maxAspectRatio:
		// It's already in the allowed range, so no borders needed.
		return inset

	case aspectRatio < minAspectRatio:
		// If inset width < 1080, it's too tall.  Add side borders to reach 1080 wide.
		height = inset.Bounds().Dy()
		width = int(float64(height) * minAspectRatio)

	case aspectRatio > maxAspectRatio:
		// If inset height < 566, it's too wide.  Add top and bottom borders to reach 566 high.
		width = inset.Bounds().Dx()
		height = int(float64(width) / maxAspectRatio)

	default:
		panic("unreachable")
	}

	composite := image.NewRGBA(image.Rect(0, 0, width, height))
	var background image.Image
	if blurBackground {
		background = backgroundForImage(original, composite.Bounds())
	} else {
		background = solidBackground
	}

	// Start with just the background color
	draw.Draw(composite, composite.Bounds(), background, image.Point{}, draw.Src)

	// Draw the inset in the middle of the composite
	compositeMid := image.Point{
		X: composite.Bounds().Min.X + composite.Bounds().Dx()/2,
		Y: composite.Bounds().Min.Y + composite.Bounds().Dy()/2,
	}
	insetMid := image.Point{
		X: inset.Bounds().Dx() / 2,
		Y: inset.Bounds().Dy() / 2,
	}
	compositArea := image.Rect(
		compositeMid.X-insetMid.X, compositeMid.Y-insetMid.Y,
		compositeMid.X+insetMid.X, compositeMid.Y+insetMid.Y,
	)
	draw.Draw(composite, compositArea, inset,
		image.Point{},
		draw.Src)

	return composite
}
