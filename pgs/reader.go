package pgs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Reader struct {
	r io.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r}
}

func (r *Reader) ReadAll() ([]DisplaySet, error) {
	var stream []DisplaySet
	for {
		ds, err := r.Read()
		if err == io.EOF {
			return stream, nil
		}
		if err != nil {
			return nil, err
		}
		stream = append(stream, *ds)
	}
}

func (r *Reader) Read() (*DisplaySet, error) {
	var ds DisplaySet

	h0, err := r.readHeader()
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("segment header: %w", err)
	}
	if h0.SegmentType != PCSType {
		return nil, fmt.Errorf("segment not PCS: %s", h0.SegmentType)
	}
	c, err := r.readPresentationComposition(h0.SegmentSize)
	if err != nil {
		return nil, fmt.Errorf("presentation composition segment: %w", err)
	}
	ds.PresentationTime = h0.PresentationTime.Duration()
	ds.DecodingTime = h0.DecodingTime.Duration()
	ds.PresentationComposition = *c

	for {
		h, err := r.readHeader()
		if err != nil {
			return nil, fmt.Errorf("segment header: %w", err)
		}
		if h.PresentationTime != h0.PresentationTime {
			return nil, fmt.Errorf("presentation time not consistent: PCS is %s, %s is %s",
				ds.PresentationTime, h.SegmentType, h.PresentationTime.Duration())
		}
		if h.DecodingTime != h0.DecodingTime {
			return nil, fmt.Errorf("decoding time not consistent: PCS is %s, %s is %s",
				ds.DecodingTime, h.SegmentType, h.DecodingTime.Duration())
		}

		switch h.SegmentType {
		case PCSType:
			return nil, errors.New("presentation composition not ended")
		case WDSType:
			if len(ds.Windows) != 0 {
				return nil, errors.New("multiple window definitions")
			}
			w, err := r.readWindows(h.SegmentSize)
			if err != nil {
				return nil, fmt.Errorf("window definition segment: %w", err)
			}
			ds.Windows = w
		case PDSType:
			if ds.Palette != nil {
				return nil, errors.New("multiple palette definitions")
			}
			p, err := r.readPalette(h.SegmentSize)
			if err != nil {
				return nil, fmt.Errorf("palette definition segment: %w", err)
			}
			ds.Palette = p
		case ODSType:
			if ds.Object != nil {
				return nil, errors.New("multiple object definitions")
			}
			o, err := r.readObject(h.SegmentSize)
			if err != nil {
				return nil, fmt.Errorf("object definition segment: %w", err)
			}
			ds.Object = o
		case ENDType:
			return &ds, nil
		}
	}
}

func (r *Reader) readHeader() (*header, error) {
	var h header
	if err := binary.Read(r.r, binary.BigEndian, &h); err != nil {
		return nil, err
	}
	if err := h.validate(); err != nil {
		return nil, err
	}
	return &h, nil
}

func (r *Reader) readPresentationComposition(segmentSize uint16) (*PresentationComposition, error) {
	var pcs pcs
	if err := binary.Read(r.r, binary.BigEndian, &pcs); err != nil {
		return nil, err
	}
	if err := pcs.validate(); err != nil {
		return nil, err
	}
	size := 11
	objects := make([]CompositionObject, pcs.ObjectCount)
	for i := range objects {
		var obj pcsObject
		if err := binary.Read(r.r, binary.BigEndian, &obj); err != nil {
			return nil, err
		}
		if err := obj.validate(); err != nil {
			return nil, fmt.Errorf("composition object %d/%d: %w", i+1, pcs.ObjectCount, err)
		}
		objects[i] = CompositionObject{
			ObjectID: obj.ObjectID,
			WindowID: obj.WindowID,
			X:        obj.X,
			Y:        obj.Y,
		}
		if obj.ObjectCropped == croppedForce {
			var crop CompositionObjectCrop
			if err := binary.Read(r.r, binary.BigEndian, &crop); err != nil {
				return nil, err
			}
			objects[i].Crop = &crop
			size += 8
		}
		size += 8
	}
	if size != int(segmentSize) {
		return nil, fmt.Errorf("read %d bytes, %d bytes declared in header", size, segmentSize)
	}
	pc := &PresentationComposition{
		Width:             pcs.Width,
		Height:            pcs.Height,
		FrameRate:         pcs.FrameRate,
		CompositionNumber: pcs.CompositionNumber,
		CompositionState:  pcs.CompositionState,
		PaletteUpdate:     pcs.PaletteUpdateFlag&pufTrue != 0,
		PaletteID:         pcs.PaletteID,
		Objects:           objects,
	}
	return pc, nil
}

func (r *Reader) readWindows(segmentSize uint16) ([]Window, error) {
	var wds wds
	if err := binary.Read(r.r, binary.BigEndian, &wds); err != nil {
		return nil, err
	}
	if err := wds.validate(segmentSize); err != nil {
		return nil, err
	}
	windows := make([]Window, wds.WindowCount)
	for i := range windows {
		if err := binary.Read(r.r, binary.BigEndian, &windows[i]); err != nil {
			return nil, err
		}
	}
	return windows, nil
}

func (r *Reader) readPalette(segmentSize uint16) (*Palette, error) {
	var pds pds
	if err := binary.Read(r.r, binary.BigEndian, &pds); err != nil {
		return nil, err
	}
	n := (segmentSize - 2) / 5
	entries := make([]PaletteEntry, n)
	for i := range entries {
		if err := binary.Read(r.r, binary.BigEndian, &entries[i]); err != nil {
			return nil, err
		}
	}
	p := &Palette{
		ID:      pds.PaletteID,
		Version: pds.PaletteVersion,
		Entries: entries,
	}
	if err := p.validate(segmentSize); err != nil {
		return nil, err
	}
	return p, nil
}

func (r *Reader) readObject(segmentSize uint16) (*Object, error) {
	var ods ods
	if err := binary.Read(r.r, binary.BigEndian, &ods); err != nil {
		return nil, err
	}
	if err := ods.validate(segmentSize); err != nil {
		return nil, err
	}
	dataLen := ods.ObjectDataLength.Int() - 4
	data := make([]byte, dataLen)
	n := 0
	for n < dataLen {
		n0, err := r.r.Read(data[n:])
		if err != nil {
			return nil, err
		}
		n += n0
	}
	obj := &Object{
		ID:      ods.ObjectID,
		Version: ods.ObjectVersion,
		First:   ods.SequenceFlag&firstInSequence != 0,
		Last:    ods.SequenceFlag&lastInSequence != 0,
		Image: Image{
			Width:  ods.Width,
			Height: ods.Height,
			Data:   data,
		},
	}
	return obj, nil
}
