package pgs

import "time"

type header struct {
	MagicNumber      uint16    // "PG" 0x5047
	PresentationTime timestamp // When sub picture is shown on screen
	DecodingTime     timestamp // When sub picture decoding starts
	SegmentType      segmentType
	SegmentSize      uint16
}

type pcs struct {
	Width, Height     uint16 // Video dimensions in pixels
	FrameRate         uint8  // Always 0x10; can be ignored
	CompositionNumber uint16
	CompositionState  CompositionState // Type of this composition
	PaletteUpdateFlag paletteUpdateFlag
	PaletteID         uint8
	ObjectCount       uint8 // Number of composition objects defined in this segment
}

type pcsObject struct {
	ObjectID      uint16
	WindowID      uint8
	ObjectCropped objectCroppedFlag
	X, Y          uint16 // Offset from the top left pixel of the screen
}

type wds struct {
	WindowCount uint8
}

type pds struct {
	PaletteID      uint8 // ID of the palette
	PaletteVersion uint8 // Version of this palette within the Epoch
}

type ods struct {
	ObjectID         uint16       // ID of this object
	ObjectVersion    uint8        // Version of this object
	SequenceFlag     sequenceFlag // If the image is split into a series of consecutive fragments, the last fragment has this flag set. Possible values:
	ObjectDataLength uint24       // The length of the Run-length Encoding (RLE) data buffer with the compressed image data.
	Width, Height    uint16       // Dimensions of the image
}

type (
	timestamp uint32
	uint24    [3]uint8

	segmentType       uint8
	paletteUpdateFlag uint8
	objectCroppedFlag uint8
	sequenceFlag      uint8
)

const (
	pcsType segmentType = 0x16 // Presentation Composition Segment
	wdsType segmentType = 0x17 // Window Definition Segment
	pdsType segmentType = 0x14 // Palette Definition Segment
	odsType segmentType = 0x15 // Object Definition Segment
	endType segmentType = 0x80 // End of Display Set Segment

	pufFalse paletteUpdateFlag = 0x00
	pufTrue  paletteUpdateFlag = 0x80

	croppedForce objectCroppedFlag = 0x40 // Force display of the cropped image object
	croppedOff   objectCroppedFlag = 0x00 // Off

	lastInSequence  sequenceFlag = 0x40
	firstInSequence sequenceFlag = 0x80
)

// Duration converts a timestamp into a Duration. Timestamps have an
// accuracy of 90 kHz, so divide by 90 to get milliseconds.
func (ts timestamp) Duration() time.Duration {
	return time.Duration(ts) * time.Millisecond / 90
}

func (ui uint24) Uint32() uint32 {
	return uint32(ui[0])<<16 | uint32(ui[1])<<8 | uint32(ui[2])
}
