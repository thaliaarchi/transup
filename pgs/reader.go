package pgs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type PresentationReader struct {
	r io.Reader
}

func NewPresentationReader(r io.Reader) *PresentationReader {
	return &PresentationReader{r}
}

func (pr *PresentationReader) Read() (*Presentation, error) {
	var pc Presentation

	h, err := pr.readHeader()
	if err != nil {
		return nil, err
	}
	if h.SegmentType != PCSType {
		return nil, fmt.Errorf("segment not PCS: %s", h.SegmentType)
	}
	c, err := pr.readPresentationComposition(h.SegmentSize)
	if err != nil {
		return nil, fmt.Errorf("presentation composition segment: %w", err)
	}
	pc.PresentationTime = h.PresentationTime
	pc.DecodingTime = h.DecodingTime
	pc.PresentationComposition = *c

	for {
		h, err := pr.readHeader()
		if err != nil {
			return nil, err
		}
		if h.PresentationTime != pc.PresentationTime {
			return nil, fmt.Errorf("presentation time not consistent: PCS is %s, %s is %s", pc.PresentationTime, h.SegmentType, h.PresentationTime)
		}
		if h.DecodingTime != pc.DecodingTime {
			return nil, fmt.Errorf("decoding time not consistent: PCS is %s, %s is %s", pc.DecodingTime, h.SegmentType, h.DecodingTime)
		}

		switch h.SegmentType {
		case PCSType:
			return nil, errors.New("presentation composition not closed")
		case WDSType:
			if pc.Windows != nil {
				return nil, errors.New("multiple window definitions in presentation composition")
			}
			w, err := pr.readWindows(h.SegmentSize)
			if err != nil {
				return nil, fmt.Errorf("window definition segment: %w", err)
			}
			pc.Windows = w
		case PDSType:
			if pc.Palette != nil {
				return nil, errors.New("multiple palette definitions in presentation composition")
			}
			p, err := pr.readPalette(h.SegmentSize)
			if err != nil {
				return nil, fmt.Errorf("palette definition segment: %w", err)
			}
			pc.Palette = p
		case ODSType:
			if pc.Object != nil {
				return nil, errors.New("multiple object definitions in presentation composition")
			}
			o, err := pr.readObject(h.SegmentSize)
			if err != nil {
				return nil, fmt.Errorf("object definition segment: %w", err)
			}
			pc.Object = o
		case ENDType:
			return &pc, nil
		}
	}
}

func (pr *PresentationReader) readHeader() (*Header, error) {
	var h header
	if err := binary.Read(pr.r, binary.BigEndian, &h); err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("segment header: %w", err)
	}
	if err := h.validate(); err != nil {
		return nil, err
	}
	return &Header{
		PresentationTime: h.PresentationTime.Duration(),
		DecodingTime:     h.DecodingTime.Duration(),
		SegmentType:      h.SegmentType,
		SegmentSize:      h.SegmentSize,
	}, nil
}

func (pr *PresentationReader) readPresentationComposition(segmentSize uint16) (*PresentationComposition, error) {
	var pcs pcs
	if err := binary.Read(pr.r, binary.BigEndian, &pcs); err != nil {
		return nil, err
	}
	if err := pcs.validate(); err != nil {
		return nil, err
	}
	size := 11
	objects := make([]CompositionObject, pcs.ObjectCount)
	for i := range objects {
		var obj pcsObject
		err := binary.Read(pr.r, binary.BigEndian, &obj)
		if err == nil {
			err = obj.validate()
		}
		if err != nil {
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
			if err := binary.Read(pr.r, binary.BigEndian, &crop); err != nil {
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

func (pr *PresentationReader) readWindows(segmentSize uint16) ([]Window, error) {
	var wds wds
	if err := binary.Read(pr.r, binary.BigEndian, &wds); err != nil {
		return nil, err
	}
	if err := wds.validate(segmentSize); err != nil {
		return nil, err
	}
	windows := make([]Window, wds.WindowCount)
	for i := range windows {
		if err := binary.Read(pr.r, binary.BigEndian, &windows[i]); err != nil {
			return nil, err
		}
	}
	return windows, nil
}

func (pr *PresentationReader) readPalette(segmentSize uint16) (*Palette, error) {
	var pds pds
	if err := binary.Read(pr.r, binary.BigEndian, &pds); err != nil {
		return nil, err
	}
	n := (segmentSize - 2) / 5
	entries := make([]PaletteEntry, n)
	ids := make(map[uint8]struct{}, n)
	for i := range entries {
		if err := binary.Read(pr.r, binary.BigEndian, &entries[i]); err != nil {
			return nil, fmt.Errorf("palette entry %d/%d: %w", i, n, err)
		}
		id := entries[i].ID
		if _, ok := ids[id]; ok {
			return nil, fmt.Errorf("palette entry %d/%d: ID %d reused", i, n, id)
		}
		ids[id] = struct{}{}
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

func (pr *PresentationReader) readObject(segmentSize uint16) (*Object, error) {
	var ods ods
	if err := binary.Read(pr.r, binary.BigEndian, &ods); err != nil {
		return nil, err
	}
	if ods.SequenceFlag&^(firstInSequence|lastInSequence) != 0 {
		return nil, fmt.Errorf("unrecognized flag: 0x%x", ods.SequenceFlag)
	}
	dataLen := int(ods.ObjectDataLength.Uint32()) - 4
	if dataLen < 0 {
		return nil, fmt.Errorf("data length excludes width and height")
	}
	data := make([]byte, dataLen)
	n := 0
	for n < dataLen {
		n0, err := pr.r.Read(data[n:])
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
