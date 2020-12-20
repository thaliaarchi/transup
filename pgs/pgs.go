package pgs

import (
	"fmt"
	"time"
)

type header struct {
	MagicNumber      uint16    // "PG" 0x5047
	PresentationTime timestamp // When sub picture is shown on screen
	DecodingTime     timestamp // When sub picture decoding starts
	SegmentType      segmentType
	SegmentSize      uint16
}

type Segment struct {
	PresentationTime time.Duration
	DecodingTime     time.Duration
	Data             interface{}
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

type PresentationComposition struct {
	Width, Height     uint16 // Video dimensions in pixels
	FrameRate         uint8  // Always 0x10; can be ignored
	CompositionNumber uint16
	CompositionState  CompositionState // Type of this composition
	PaletteUpdate     bool
	PaletteID         uint8
	Objects           []CompositionObject
}

type pcsCompositionObject struct {
	ObjectID      uint16
	WindowID      uint8
	ObjectCropped objectCroppedFlag
	X, Y          uint16 // Offset from the top left pixel of the screen
}

type CompositionObject struct {
	ObjectID uint16
	WindowID uint8
	X, Y     uint16 // Offset from the top left pixel of the screen
	Crop     *CompositionObjectCrop
}

type CompositionObjectCrop struct {
	// Offset of the cropped object from the top left pixel of the screen
	X, Y          uint16
	Width, Height uint16 // Dimensions of the cropped object
}

type wds struct {
	WindowCount uint8
}

type Window struct {
	WindowID      uint8
	X, Y          uint16
	Width, Height uint16
}

type pds struct {
	PaletteID      uint8 // ID of the palette
	PaletteVersion uint8 // Version of this palette within the Epoch
}

type Palette struct {
	ID      uint8
	Version uint8
	Entries []PaletteEntry
}

type PaletteEntry struct {
	ID uint8 // Entry number of the palette
	// Luminance, blue difference, red difference, alpha
	Y, Cr, Cb, A uint8
}

type ods struct {
	ObjectID         uint16       // ID of this object
	ObjectVersion    uint8        // Version of this object
	SequenceFlag     sequenceFlag // If the image is split into a series of consecutive fragments, the last fragment has this flag set. Possible values:
	ObjectDataLength uint24       // The length of the Run-length Encoding (RLE) data buffer with the compressed image data.
	Width, Height    uint16       // Dimensions of the image
	// ObjectData	variable	This is the image data compressed using Run-length Encoding (RLE). The size of the data is defined in the Object Data Length field.
}

type Object struct {
	ID            uint16 // ID of this object
	Version       uint8  // Version of this object
	First, Last   bool
	Width, Height uint16 // Dimensions of the image
	ObjectData    []byte // This is the image data compressed using Run-length Encoding (RLE). The size of the data is defined in the Object Data Length field.
}

type uint24 [3]uint8

func (ui uint24) Uint32() uint32 {
	return uint32(ui[0])<<16 | uint32(ui[1])<<8 | uint32(ui[2])
}

type (
	timestamp         uint32
	segmentType       uint8
	CompositionState  uint8
	paletteUpdateFlag uint8
	objectCroppedFlag uint8
	sequenceFlag      uint8
)

const (
	pcsType segmentType = 0x16 // Presentation Composition Segment
	wdsType             = 0x17 // Window Definition Segment
	pdsType             = 0x14 // Palette Definition Segment
	odsType             = 0x15 // Object Definition Segment
	endType             = 0x80 // End of Display Set Segment
)

// The composition state can be one of three values
const (
	// Epoch Start defines a new display. The Epoch Start contains all
	// functional segments needed to display a new composition on the
	// screen.
	EpochStart CompositionState = 0x80
	// Acquisition Point defines a display refresh. This is used to
	// compose in the middle of the Epoch. It includes functional segments
	// with new objects to be used in a new composition, replacing old
	// objects with the same Object ID.
	AcquisitionPoint = 0x40
	// Normal defines a display update, and contains only functional
	// segments with elements that are different from the preceding
	// composition. It’s mostly used to stop displaying objects on the
	// screen by defining a composition with no composition objects (a
	// value of zero in the Number of Composition Objects flag) but also
	// used to define a new composition with new objects and objects
	// defined since the Epoch Start.
	Normal = 0x00
)

const (
	pufFalse paletteUpdateFlag = 0x00
	pufTrue                    = 0x80
)

const (
	croppedForce objectCroppedFlag = 0x40 // Force display of the cropped image object
	croppedOff                     = 0x00 // Off
)

const (
	lastInSequence  sequenceFlag = 0x40
	firstInSequence              = 0x80
)

func (st segmentType) String() string {
	switch st {
	case pcsType:
		return "PCS"
	case wdsType:
		return "WDS"
	case pdsType:
		return "PDS"
	case odsType:
		return "ODS"
	case endType:
		return "End"
	}
	return fmt.Sprintf("SegmentType(%x)", uint8(st))
}

// Duration converts a timestamp into a Duration. Timestamps have an
// accuracy of 90 kHz, so divide by 90 to get milliseconds.
func (ts timestamp) Duration() time.Duration {
	return time.Duration(ts) * time.Millisecond / 90
}
