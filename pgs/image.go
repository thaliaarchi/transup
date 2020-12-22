package pgs

import (
	"fmt"
	"image"
	"image/color"
	"os"
)

func (img *Image) Convert(p *Palette) (*image.Paletted, error) {
	var cp color.Palette
	ids := make([]uint8, len(p.Entries))
	idMap := make(map[uint8]uint8)
	for i, e := range p.Entries {
		ids[i] = e.ID
		idMap[e.ID] = uint8(i)
		cp = append(cp, e.NYCbCrA)
	}
	rect := image.Rectangle{Max: image.Point{int(img.Width), int(img.Height)}}
	pimg := image.NewPaletted(rect, cp)

	d := img.Data
	x, y := 0, 0
	for i := 0; i < len(d); {
		if d[i] != 0 { // CCCCCCCC - One pixel in color C
			pimg.SetColorIndex(x, y, d[i])
			i++
			x++
			continue
		}
		i++
		var c uint8
		var l uint16

		switch d[i] & 0xc0 {
		case 0x00:
			if d[i] == 0 { // 00000000 00000000 - End of line
				if x+1 != int(img.Width) {
					fmt.Fprintf(os.Stderr, "line %d has width %d instead of %d\n", y, x+1, img.Width)
					// return nil, fmt.Errorf("line %d has width %d instead of %d", y, x+1, img.Width)
				}
				x = 0
				y++
				continue
			}
			// 00000000 00LLLLLL - L pixels in color 0 (L between 1 and 63)
			l = uint16(d[i] & 0x3f)
			i++
		case 0x40: // 00000000 01LLLLLL LLLLLLLL - L pixels in color 0 (L between 64 and 16383)
			l = uint16(d[i]&0x3f)<<8 | uint16(d[i+1])
			i += 2
		case 0x80: // 00000000 10LLLLLL CCCCCCCC - L pixels in color C (L between 3 and 63)
			l = uint16(d[i] & 0x3f)
			c = idMap[d[i+1]]
			i += 2
		case 0xc0: // 00000000 11LLLLLL LLLLLLLL CCCCCCCC - L pixels in color C (L between 64 and 16383)
			l = uint16(d[i]&0x3f)<<8 | uint16(d[i+1])
			c = idMap[d[i+2]]
			i += 3
		default:
			panic("impossible")
		}
		for i := 0; i < int(l); i++ {
			pimg.SetColorIndex(x+i, y, c)
		}
		x += int(l)
	}
	if y+1 != int(img.Height) {
		return nil, fmt.Errorf("image has height %d instead of %d", y+1, img.Height)
	}
	return pimg, nil
}
