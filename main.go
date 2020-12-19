package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/andrewarchi/transup/pgs"
)

func main() {
	f, err := os.Open("subs.sup")
	try(err)
	defer f.Close()

Loop:
	for {
		var h pgs.Header
		try(binary.Read(f, binary.BigEndian, &h))
		fmt.Println(h)
		switch h.SegmentType {
		case pgs.PCSType:
			var pcs pgs.PCS
			try(binary.Read(f, binary.BigEndian, &pcs))
			fmt.Println(pcs)
			l := uint16(11)
			for i := uint8(0); i < pcs.ObjectCount; i++ {
				var co pgs.CompositionObject
				try(binary.Read(f, binary.BigEndian, &co))
				fmt.Print("  ", co)
				l += 8
				if co.ObjectCropped == pgs.CroppedForce {
					var crop pgs.CompositionObjectCrop
					try(binary.Read(f, binary.BigEndian, &crop))
					fmt.Print(crop)
					l += 8
				}
				fmt.Println()
			}
			if l != h.SegmentSize {
				fmt.Println("bad pcs read length:", h.SegmentSize, l)
			}

		case pgs.WDSType:
			var wds pgs.WDS
			try(binary.Read(f, binary.BigEndian, &wds))
			fmt.Println(wds)
			if h.SegmentSize != uint16(wds.WindowCount)*9+1 {
				fmt.Println("bad window read length:", h.SegmentSize, uint16(wds.WindowCount)*9+1)
			}
			for i := uint8(0); i < wds.WindowCount; i++ {
				var w pgs.Window
				try(binary.Read(f, binary.BigEndian, &w))
				fmt.Println(" ", w)
			}

		case pgs.PDSType:
			var p pgs.Palette
			try(binary.Read(f, binary.BigEndian, &p))
			fmt.Println(" ", p)
			for i := uint16(0); i < (h.SegmentSize-2)/5; i++ {
				var e pgs.PaletteEntry
				try(binary.Read(f, binary.BigEndian, &e))
				fmt.Println("   ", e)
			}

		case pgs.ODSType:
			break Loop

		case pgs.EndType:
			break Loop
		}
	}

	d := hex.Dumper(os.Stdout)
	var b [256]byte
	n, err := f.Read(b[:])
	try(err)
	_, err = d.Write(b[:n])
	try(err)
}

func try(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
