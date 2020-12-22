package main

import (
	"fmt"
	"image/png"
	"io"
	"log"
	"os"

	"github.com/andrewarchi/transup/pgs"
)

func main() {
	f, err := os.Open(os.Args[1])
	try(err)
	defer f.Close()

	sr := pgs.NewSegmentReader(f)
	var p *pgs.Palette
	for {
		s, err := sr.ReadSegment()
		if err == io.EOF {
			break
		}
		try(err)
		var typ string
		switch d := s.Data.(type) {
		case *pgs.PresentationComposition:
			typ = "PCS"
		case []pgs.Window:
			typ = "WDS"
		case *pgs.Palette:
			typ = "PDS"
			p = d
		case *pgs.Object:
			img, err := d.Image.Convert(p)
			try(err)
			f, err := os.Create(fmt.Sprintf("obj_%s.png", s.PresentationTime))
			try(err)
			try(png.Encode(f, img))
			typ = "ODS"
		case nil:
			typ = "END"
		}
		fmt.Printf("%s %v ", typ, s.PresentationTime)
		if s.DecodingTime != s.PresentationTime {
			fmt.Printf("Decoding:%v ", s.DecodingTime)
		}
		fmt.Printf("%+v\n", s.Data)
		if s.Data == nil {
			fmt.Println()
		}
	}
}

func try(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
