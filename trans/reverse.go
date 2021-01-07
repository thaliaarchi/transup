package trans

import (
	"errors"
	"fmt"
	"time"

	"github.com/andrewarchi/transup/pgs"
)

func Reverse(stream []pgs.DisplaySet, d time.Duration) ([]pgs.DisplaySet, error) {
	if len(stream)%2 != 0 {
		return nil, errors.New("len not even")
	}
	rev := make([]pgs.DisplaySet, len(stream))
	j := 0
	for i := len(stream) - 2; i >= 0; i -= 2 {
		draw := &stream[i]
		clear := &stream[i+1]
		if draw.PresentationTime > d || draw.DecodingTime > d {
			return nil, fmt.Errorf("display set %d/%d: presentation %s or decoding time %s greater than duration %s",
				i, len(stream), draw.PresentationTime, draw.DecodingTime, d)
		}
		if clear.PresentationTime > d || clear.DecodingTime > d {
			return nil, fmt.Errorf("display set %d/%d: presentation %s or decoding time %s greater than duration %s",
				i+1, len(stream), draw.PresentationTime, draw.DecodingTime, d)
		}
		if draw.CompositionState != pgs.EpochStart {
			return nil, fmt.Errorf("display set %d/%d: composition state is not epoch start, got %x",
				i, len(stream), draw.CompositionState)
		}
		if clear.CompositionState != pgs.Normal ||
			clear.PaletteUpdate || clear.Palette != nil ||
			len(clear.Objects) != 0 || clear.Object != nil {
			return nil, fmt.Errorf("display set %d/%d: appears to not clear objects", i+1, len(stream))
		}
		rev[j] = *draw
		rev[j].PresentationTime = d - clear.PresentationTime
		rev[j].DecodingTime = d - clear.DecodingTime
		rev[j+1] = *clear
		rev[j+1].PresentationTime = d - draw.PresentationTime
		rev[j+1].DecodingTime = d - draw.DecodingTime
		j += 2
	}
	return rev, nil
}
