package pgs

import (
	"fmt"
	"image"
	"image/color"
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
		var c uint8
		var l uint16

		hd1, ld1 := d[i+1]&0xc0, d[i+1]&0x3f
		switch hd1 {
		case 0x00:
			i += 2
			// 00000000 00000000 - End of line
			if ld1 == 0 {
				if x != int(img.Width) {
					return nil, fmt.Errorf("line %d has width %d instead of %d", y, x, img.Width)
				}
				x = 0
				y++
				continue
			}
			// 00000000 00LLLLLL - L pixels in color 0
			l = uint16(ld1)
		// 00000000 01LLLLLL LLLLLLLL - L pixels in color 0
		case 0x40:
			l = uint16(ld1)<<8 | uint16(d[i+2])
			i += 3
		// 00000000 10LLLLLL CCCCCCCC - L pixels in color C
		case 0x80:
			l = uint16(ld1)
			c = idMap[d[i+2]]
			i += 3
		// 00000000 11LLLLLL LLLLLLLL CCCCCCCC - L pixels in color C
		case 0xc0:
			l = uint16(ld1)<<8 | uint16(d[i+2])
			c = idMap[d[i+3]]
			i += 4
		default:
			panic("impossible")
		}
		for i := 0; i < int(l); i++ {
			pimg.SetColorIndex(x+i, y, c)
		}
		x += int(l)
	}
	if x != 0 {
		return nil, fmt.Errorf("line %d has width %d instead of %d", y, x, img.Width)
	}
	if y != int(img.Height) {
		return nil, fmt.Errorf("image has height %d instead of %d", y, img.Height)
	}
	return pimg, nil
}
