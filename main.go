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
	"runtime"
	"sync"

	"github.com/nfnt/resize"
	_ "golang.org/x/image/tiff"
)

// Instagram 1080 wide by 566 - 1350 high

var background image.Image

func main() {
	var (
		backgroundGray float64
		white          bool
		black          bool
	)

	flag.Float64Var(&backgroundGray, "background", 0.125,
		"background gray, 0.0 to 1.0")
	flag.BoolVar(&white, "white", false, "white background")
	flag.BoolVar(&black, "black", false, "black background")
	flag.Parse()

	if white {
		backgroundGray = 1
	}
	if black {
		backgroundGray = 0
	}
	background = image.NewUniform(color.Gray{Y: uint8(backgroundGray * math.MaxUint8)})

	wg := sync.WaitGroup{}
	jobs := make(chan string)
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}
	for _, filename := range flag.Args() {
		jobs <- filename
	}
	close(jobs)
	wg.Wait()
}

func worker(in chan string, done *sync.WaitGroup) {
	for fname := range in {
		doFile(fname)
	}
	done.Done()
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
	outName := fname + "_insta.jpg"
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

func doImage(original image.Image) image.Image {
	inset := resize.Thumbnail(1080, 1350,
		original, resize.Lanczos3)

	var composite *image.RGBA
	switch {
	case inset.Bounds().Dx() < 1080:
		// If inset width < 1080, it's too tall.  Add side borders to reach 1080 wide.
		composite = image.NewRGBA(image.Rect(0, 0, 1080, 1350))

	case inset.Bounds().Dy() < 566:
		// If inset height < 566, it's too wide.  Add top and bottom borders to reach 566 high.
		composite = image.NewRGBA(image.Rect(0, 0, 1080, 566))

	default:
		// Image size is great
		return inset
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
