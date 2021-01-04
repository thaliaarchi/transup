package pgs

import "fmt"

func (h *header) validate() error {
	if h.MagicNumber != 0x5047 {
		return fmt.Errorf(`magic number not "PG" 0x5047: %x`, h.MagicNumber)
	}
	switch h.SegmentType {
	case PCSType, WDSType, ODSType:
	case PDSType:
		if h.SegmentSize%5 != 2 {
			return fmt.Errorf("invalid segment size: %d bytes", h.SegmentSize)
		}
	case ENDType:
		if h.SegmentSize != 0 {
			return fmt.Errorf("nonzero segment size: %d bytes", h.SegmentSize)
		}
	default:
		return fmt.Errorf("unrecognized segment type: 0x%x", h.SegmentType)
	}
	if h.DecodingTime > h.PresentationTime {
		return fmt.Errorf("decoding time %s (0x%x) after presentation time %s (0x%x)",
			h.DecodingTime.Duration(), h.DecodingTime, h.PresentationTime.Duration(), h.PresentationTime)
	}
	return nil
}

func (pcs *pcs) validate() error {
	switch pcs.CompositionState {
	case Normal, AcquisitionPoint, EpochStart:
	default:
		return fmt.Errorf("unrecognized composition state: 0x%x", pcs.CompositionState)
	}
	if pcs.PaletteUpdateFlag&^pufTrue != 0 {
		return fmt.Errorf("unrecognized palette update flag: 0x%x", pcs.PaletteUpdateFlag)
	}
	return nil
}

func (obj *pcsObject) validate() error {
	if obj.ObjectCropped&^croppedForce != 0 {
		return fmt.Errorf("unrecognized object crop flag: 0x%x", obj.ObjectCropped)
	}
	return nil
}

func (wds *wds) validate(segmentSize uint16) error {
	if segmentSize != uint16(wds.WindowCount)*9+1 {
		return fmt.Errorf("segment size %d indicates %d windows, but %d specified",
			segmentSize, uint16(wds.WindowCount)*9+1, wds.WindowCount)
	}
	return nil
}

func (p *Palette) validate(segmentSize uint16) error {
	ids := make(map[uint8]struct{}, len(p.Entries))
	for _, e := range p.Entries {
		if _, ok := ids[e.ID]; ok {
			return fmt.Errorf("id reused: %d", e.ID)
		}
		ids[e.ID] = struct{}{}
	}
	return nil
}

func (ods *ods) validate(segmentSize uint16) error {
	if ods.SequenceFlag&^(firstInSequence|lastInSequence) != 0 {
		return fmt.Errorf("unrecognized flag: 0x%x", ods.SequenceFlag)
	}
	return nil
}
