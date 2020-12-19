package pgs

import (
	"encoding/binary"
	"fmt"
	"io"
)

type SegmentReader struct {
	r io.Reader
}

type SizeMismatchError struct {
	ReadSize, HeaderSize uint16
}

func (err *SizeMismatchError) Error() string {
	return fmt.Sprintf("read %d bytes, %d bytes declared in header", err.ReadSize, err.HeaderSize)
}

func (sr *SegmentReader) ReadSegment() (*Segment, error) {
	var h Header
	if err := binary.Read(sr.r, binary.BigEndian, &h); err != nil {
		return nil, fmt.Errorf("segment header: %w", err)
	}
	switch h.SegmentType {
	case PCSType:
	case WDSType:
	case PDSType:
	case ODSType:
	case EndType:
	}
	panic("incomplete")
}

func (sr *SegmentReader) readPCS(segmentSize uint16) (*PCS, error) {
	var pcs PCS
	if err := binary.Read(sr.r, binary.BigEndian, &pcs); err != nil {
		return nil, fmt.Errorf("presentation composition segment: %w", err)
	}
	size := uint16(11)
	for i := uint8(0); i < pcs.ObjectCount; i++ {
		var co CompositionObject
		if err := binary.Read(sr.r, binary.BigEndian, &co); err != nil {
			return nil, fmt.Errorf("composition object %d/%d: %w", i+1, pcs.ObjectCount, err)
		}
		size += 8
		if co.ObjectCropped == CroppedForce {
			var crop CompositionObjectCrop
			if err := binary.Read(sr.r, binary.BigEndian, &crop); err != nil {
				return nil, fmt.Errorf(": %w", err)
			}
			size += 8
		}
	}
	if size != segmentSize {
		return nil, fmt.Errorf("presentation composition segment: %w",
			&SizeMismatchError{size, segmentSize})
	}
	return &pcs, nil
}
