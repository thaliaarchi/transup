package main

import (
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"time"

	"github.com/andrewarchi/transup/pgs"
	"github.com/andrewarchi/transup/trans"
)

const usage = `Usage:
	transup reverse <filename> <duration> [out]
	transup dump <filename> <image-dir>`

func main() {
	if len(os.Args) < 4 || len(os.Args) > 5 ||
		!(os.Args[1] == "reverse" ||
			(os.Args[1] == "dump" && len(os.Args) == 4)) {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	cmd, filename := os.Args[1], os.Args[2]

	f, err := os.Open(filename)
	try(err)
	defer f.Close()
	stream, err := pgs.NewReader(f).ReadAll()
	try(err)

	switch cmd {
	case "reverse":
		out := os.Stdout
		if len(os.Args) == 5 {
			out, err = os.Create(os.Args[4])
			try(err)
		}
		d, err := time.ParseDuration(os.Args[3])
		try(err)
		rev, err := trans.Reverse(stream, d)
		try(err)
		try(pgs.NewWriter(out).WriteAll(rev))
	case "dump":
		dirname := os.Args[3]
		try(os.MkdirAll(dirname, 0755))
		n := 0
		for i, ds := range stream {
			if i != 0 {
				fmt.Println()
			}
			fmt.Printf("Presentation: %s Decoding:%s\n", ds.PresentationTime, ds.DecodingTime)
			fmt.Printf("Composition: %+v\n", ds.PresentationComposition)
			if ds.Windows != nil {
				fmt.Printf("Windows: %+v\n", ds.Windows)
			}
			if ds.Palette != nil {
				fmt.Printf("Palette: %+v\n", ds.Palette)
			}
			if ds.Object != nil {
				n++
				fmt.Printf("Object: %+v\n", ds.Object)
				img, err := ds.Object.Image.Convert(ds.Palette)
				try(err)
				name := fmt.Sprintf("sub_%d_%s.png", n, ds.PresentationTime)
				f, err := os.Create(filepath.Join(dirname, name))
				try(err)
				try(png.Encode(f, img))
			}
		}
	}
}

func try(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
