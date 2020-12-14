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
	"os"
	"runtime"
	"sync"

	"github.com/nfnt/resize"

	_ "golang.org/x/image/tiff"
)

type serviceType struct {
	name  string
	sizes []image.Point
}

var services = []*serviceType{
	{"instagram", []image.Point{
		{1080, 1080},
		{1080, 608},
		{1080, 1350},
	}},
	{"facebook", []image.Point{
		{1200, 1200},
	}},
	//{"twitter", []image.Point{
	//	{1024, 512},
	//}},
	//{"linkedin", []image.Point{
	//	{1400, 800},
	//}},
	//{"pinterest", []image.Point{
	//	{1000, 1000},
	//	{1000, 1500},
	//}},
}

var backgroundColor = color.Gray{Y: 32}
var background = image.NewUniform(backgroundColor)

func main() {
	flag.Parse()

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

	for _, service := range services {
		composite := doService(service, original)
		outName := fname + "_" + service.name + ".jpg"
		o, err := os.Create(outName)
		if err != nil {
			log.Print(err)
			continue
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
}

func doService(s *serviceType, original image.Image) image.Image {
	compositeRect, _ := s.bestBounds(original.Bounds())

	composite := image.NewRGBA(compositeRect)

	inset := resize.Thumbnail(uint(compositeRect.Max.X), uint(compositeRect.Max.Y),
		original, resize.Lanczos3)

	draw.Draw(composite, composite.Bounds(), background, image.Point{}, draw.Src)
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

func (s *serviceType) bestBounds(in image.Rectangle) (container, inset image.Rectangle) {
	coverage := 0.0

	inAspect := float64(in.Dx()) / float64(in.Dy())
	inX := float64(in.Dx())
	inY := float64(in.Dy())
	for _, size := range s.sizes {
		sizeAspect := float64(size.X) / float64(size.Y)
		sizeX := float64(size.X)
		sizeY := float64(size.Y)
		var sizeCoverage, ratio float64
		switch {
		case inAspect == sizeAspect:
			sizeCoverage = 1.0
			ratio = sizeX / inX
		case inAspect < sizeAspect:
			sizeCoverage = (inX / sizeX) * (sizeY / inY)
			ratio = sizeY / inY
		case inAspect > sizeAspect:
			sizeCoverage = (inY / sizeY) * (sizeX / inX)
			ratio = sizeX / inX
		}

		if sizeCoverage > coverage {
			coverage = sizeCoverage
			container.Max = size
			inset.Max = image.Point{int(inX * ratio), int(inY * ratio)}
		}
	}

	return
}
