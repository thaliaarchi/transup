package trans

import (
	"errors"
	"fmt"
	"time"

	"github.com/andrewarchi/transup/pgs"
)

func Reverse(ds []pgs.DisplaySet, d time.Duration) ([]pgs.DisplaySet, error) {
	if len(ds)%2 != 0 {
		return nil, errors.New("len not even")
	}
	dsr := make([]pgs.DisplaySet, len(ds))
	j := 0
	for i := len(ds) - 2; i >= 0; i -= 2 {
		draw := &ds[i]
		clear := &ds[i+1]
		if draw.PresentationTime > d || draw.DecodingTime > d {
			return nil, fmt.Errorf("display set %d/%d: presentation %s or decoding time %s greater than duration %s",
				i, len(ds), draw.PresentationTime, draw.DecodingTime, d)
		}
		if clear.PresentationTime > d || clear.DecodingTime > d {
			return nil, fmt.Errorf("display set %d/%d: presentation %s or decoding time %s greater than duration %s",
				i+1, len(ds), draw.PresentationTime, draw.DecodingTime, d)
		}
		if draw.CompositionState != pgs.EpochStart {
			return nil, fmt.Errorf("display set %d/%d: composition state is not epoch start, got %x",
				i, len(ds), draw.CompositionState)
		}
		if clear.CompositionState != pgs.Normal ||
			clear.PaletteUpdate || clear.Palette != nil ||
			len(clear.Objects) != 0 || clear.Object != nil {
			return nil, fmt.Errorf("display set %d/%d: appears to not clear objects", i+1, len(ds))
		}
		dsr[j] = *draw
		dsr[j].PresentationTime = d - clear.PresentationTime
		dsr[j].DecodingTime = d - clear.DecodingTime
		dsr[j+1] = *clear
		dsr[j+1].PresentationTime = d - draw.PresentationTime
		dsr[j+1].DecodingTime = d - draw.DecodingTime
		j += 2
	}
	return dsr, nil
}
