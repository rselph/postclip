package main

import (
	"fmt"
	"image"
	"testing"
)

var tests = []image.Rectangle{
	image.Rectangle{image.Point{}, image.Point{80, 100}},
	image.Rectangle{image.Point{}, image.Point{100, 80}},
	image.Rectangle{image.Point{}, image.Point{80, 80}},
	image.Rectangle{image.Point{}, image.Point{80, 8}},
}

func TestBasicBounds(t *testing.T) {
	for _, test := range tests {
		fmt.Println(test)
		for _, service := range services {
			bounds, inset := service.bestBounds(test)
			fmt.Println("   ", service.name, bounds, inset)
		}
	}
}
