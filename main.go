package main

import (
	"flag"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"sync"

	_ "golang.org/x/image/tiff"
)

type service struct {
	name  string
	sizes []image.Point
}

var services = []service{
	{"instagram", []image.Point{
		{1080, 1080},
		{1080, 608},
		{1080, 1350},
	}},
	{"facebook", []image.Point{
		{1200, 1200},
	}},
	{"twitter", []image.Point{
		{1024, 512},
	}},
	{"linkedin", []image.Point{
		{1400, 800},
	}},
	{"pinterest", []image.Point{
		{1000, 1000},
		{1000, 1500},
	}},
}

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
	defer f.Close()

	original, _, err := image.Decode(f)
	if err != nil {
		log.Print(err)
		return
	}

	services[0].bestBounds(original.Bounds())
}

func (s *service) bestBounds(in image.Rectangle) (container, inset image.Rectangle) {
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
