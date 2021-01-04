package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/andrewarchi/transup/pgs"
)

func main() {
	f, err := os.Open(os.Args[1])
	try(err)
	defer f.Close()

	pr := pgs.NewPresentationReader(f)
	for {
		pc, err := pr.Read()
		if err == io.EOF {
			break
		}
		try(err)
		fmt.Printf("Composition: %+v\n", pc.PresentationComposition)
		if pc.Windows != nil {
			fmt.Printf("Windows: %+v\n", pc.Windows)
		}
		if pc.Palette != nil {
			fmt.Printf("Palette: %+v\n", pc.Palette)
		}
		if pc.Object != nil {
			fmt.Printf("Object: %+v\n", pc.Object)
			// img, err := pc.Object.Image.Convert(pc.Palette)
			// try(err)
			// f, err := os.Create(fmt.Sprintf("obj_%s.png", pc.PresentationTime))
			// try(err)
			// try(png.Encode(f, img))
		}
		fmt.Println()
	}
}

func try(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
