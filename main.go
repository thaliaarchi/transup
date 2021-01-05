package main

import (
	"fmt"
	"os"

	"github.com/andrewarchi/transup/pgs"
)

func main() {
	f, err := os.Open(os.Args[1])
	try(err)
	defer f.Close()

	stream, err := pgs.NewReader(f).ReadAll()
	try(err)
	for _, ds := range stream {
		fmt.Printf("Presentation: %s Decoding:%s\n", ds.PresentationTime, ds.DecodingTime)
		fmt.Printf("Composition: %+v\n", ds.PresentationComposition)
		if ds.Windows != nil {
			fmt.Printf("Windows: %+v\n", ds.Windows)
		}
		if ds.Palette != nil {
			fmt.Printf("Palette: %+v\n", ds.Palette)
		}
		if ds.Object != nil {
			fmt.Printf("Object: %+v\n", ds.Object)
			// img, err := ds.Object.Image.Convert(ds.Palette)
			// try(err)
			// f, err := os.Create(fmt.Sprintf("obj_%s.png", ds.PresentationTime))
			// try(err)
			// try(png.Encode(f, img))
		}
		fmt.Println()
	}
}

func try(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(1)
	}
}
