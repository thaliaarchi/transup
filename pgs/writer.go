package pgs

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func (w *Writer) Write(ds *DisplaySet) error {
	h := header{
		MagicNumber:      0x5047,
		PresentationTime: fromDuration(ds.PresentationTime),
		DecodingTime:     fromDuration(ds.DecodingTime),
	}
	if err := w.writePresentationComposition(h, &ds.PresentationComposition); err != nil {
		return fmt.Errorf("presentation composition segment: %w", err)
	}
	if len(ds.Windows) != 0 {
		if err := w.writeWindows(h, ds.Windows); err != nil {
			return fmt.Errorf("window definition segment: %w", err)
		}
	}
	if ds.Palette != nil {
		if err := w.writePalette(h, ds.Palette); err != nil {
			return fmt.Errorf("palette definition segment: %w", err)
		}
	}
	if ds.Object != nil {
		if err := w.writeObject(h, ds.Object); err != nil {
			return fmt.Errorf("palette definition segment: %w", err)
		}
	}
	h.SegmentType = ENDType
	return w.writeHeader(&h)
}

func (w *Writer) writeHeader(h *header) error {
	if err := h.validate(); err != nil {
		return err
	}
	return binary.Write(w.w, binary.BigEndian, h)
}

func (w *Writer) writePresentationComposition(h header, pc *PresentationComposition) error {
	if len(pc.Objects) > 0xff {
		return fmt.Errorf("object count overflow: %d", len(pc.Objects))
	}
	size := uint16(11)
	for i := range pc.Objects {
		if pc.Objects[i].Crop != nil {
			size += 8
		}
		size += 8
	}
	h.SegmentType = PCSType
	h.SegmentSize = size

	var puf paletteUpdateFlag
	if pc.PaletteUpdate {
		puf |= pufTrue
	}
	pcs := &pcs{
		Width:             pc.Width,
		Height:            pc.Height,
		FrameRate:         pc.FrameRate,
		CompositionNumber: pc.CompositionNumber,
		CompositionState:  pc.CompositionState,
		PaletteUpdateFlag: puf,
		PaletteID:         pc.PaletteID,
		ObjectCount:       uint8(len(pc.Objects)),
	}
	if err := pcs.validate(); err != nil {
		return err
	}

	if err := w.writeHeader(&h); err != nil {
		return err
	}
	if err := binary.Write(w.w, binary.BigEndian, pcs); err != nil {
		return err
	}
	for i, obj := range pc.Objects {
		var cropped objectCroppedFlag
		if obj.Crop != nil {
			cropped |= croppedForce
		}
		o := pcsObject{
			ObjectID:      obj.ObjectID,
			WindowID:      obj.WindowID,
			ObjectCropped: cropped,
			X:             obj.X,
			Y:             obj.Y,
		}
		if err := o.validate(); err != nil {
			return fmt.Errorf("composition object %d/%d: %w", i+1, len(pc.Objects), err)
		}
		if err := binary.Write(w.w, binary.BigEndian, &o); err != nil {
			return err
		}
		if obj.Crop != nil {
			if err := binary.Write(w.w, binary.BigEndian, &obj.Crop); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *Writer) writeWindows(h header, ws []Window) error {
	if len(ws) > 0xff {
		return fmt.Errorf("window count overflow: %d", len(ws))
	}
	h.SegmentType = WDSType
	h.SegmentSize = uint16(len(ws))*9 + 1
	wds := &wds{WindowCount: uint8(len(ws))}
	if err := wds.validate(h.SegmentSize); err != nil {
		return err
	}

	if err := w.writeHeader(&h); err != nil {
		return err
	}
	if err := binary.Write(w.w, binary.BigEndian, wds); err != nil {
		return err
	}
	for i := range ws {
		if err := binary.Write(w.w, binary.BigEndian, &ws[i]); err != nil {
			return err
		}
	}
	return nil
}

func (w *Writer) writePalette(h header, p *Palette) error {
	h.SegmentType = PDSType
	h.SegmentSize = uint16(len(p.Entries)*5 + 2)
	pds := &pds{p.ID, p.Version}
	if err := p.validate(h.SegmentSize); err != nil {
		return err
	}

	if err := w.writeHeader(&h); err != nil {
		return err
	}
	if err := binary.Write(w.w, binary.BigEndian, pds); err != nil {
		return err
	}
	return binary.Write(w.w, binary.BigEndian, p.Entries)
}

func (w *Writer) writeObject(h header, obj *Object) error {
	if len(obj.Data) > 0xffffff-4 {
		return fmt.Errorf("object data length overflow: %d", len(obj.Data))
	}
	h.SegmentType = ODSType
	h.SegmentSize = uint16(len(obj.Data) + 11)

	var seq sequenceFlag
	if obj.First {
		seq |= firstInSequence
	}
	if obj.Last {
		seq |= lastInSequence
	}
	l, err := uint24FromInt(len(obj.Data) + 4)
	if err != nil {
		return err
	}
	ods := &ods{
		ObjectID:         obj.ID,
		ObjectVersion:    obj.Version,
		SequenceFlag:     seq,
		ObjectDataLength: l,
		Width:            obj.Width,
		Height:           obj.Height,
	}
	if err := ods.validate(h.SegmentSize); err != nil {
		return err
	}

	if err := w.writeHeader(&h); err != nil {
		return err
	}
	if err := binary.Write(w.w, binary.BigEndian, ods); err != nil {
		return err
	}
	return binary.Write(w.w, binary.BigEndian, obj.Data)
}
