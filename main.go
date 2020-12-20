package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/andrewarchi/transup/pgs"
)

func main() {
	f, err := os.Open("subs.sup")
	try(err)
	defer f.Close()

	sr := pgs.NewSegmentReader(f)
	for {
		s, err := sr.ReadSegment()
		if err == io.EOF {
			break
		}
		try(err)
		fmt.Printf("Presentation:%v Decoding:%v %T ", s.PresentationTime, s.DecodingTime, s.Data)
		switch d := s.Data.(type) {
		case *pgs.Palette:
			d.Entries = nil
		case *pgs.Object:
			d.ObjectData = nil
		}
		fmt.Printf("Data:%+v\n", s.Data)
	}
}

func try(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
