package pgs

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

type Header struct {
	MagicNumber      uint16    // "PG" 0x5047
	PresentationTime Timestamp // When sub picture is shown on screen
	DecodingTime     Timestamp // When sub picture decoding starts
	SegmentType      SegmentType
	SegmentSize      uint16
}

type Segment struct {
	Header
	Segment interface{}
}

type PCS struct {
	Width, Height     uint16 // Video dimensions in pixels
	FrameRate         uint8  // Always 0x10; can be ignored
	CompositionNumber uint16
	CompositionState  CompositionState // Type of this composition
	PaletteUpdateFlag PaletteUpdateFlag
	PaletteID         uint8
	ObjectCount       uint8 // Number of composition objects defined in this segment
}

type CompositionObject struct {
	ObjectID      uint16
	WindowID      uint8
	ObjectCropped ObjectCroppedFlag
	// Offset from the top left pixel of the screen
	X, Y uint16
}

type CompositionObjectCrop struct {
	// Offset of the cropped object from the top left pixel of the screen
	CropX, CropY          uint16
	CropWidth, CropHeight uint16 // Dimensions of the cropped object
}

type WDS struct {
	WindowCount uint8
}

type Window struct {
	WindowID      uint8
	X, Y          uint16
	Width, Height uint16
}

type Palette struct {
	PaletteID            uint8 // ID of the palette
	PaletteVersionNumber uint8 // Version of this palette within the Epoch
}

type PaletteEntry struct {
	PaletteEntryID uint8 // Entry number of the palette
	// Luminance, blue difference, red difference, alpha
	Y, Cr, Cb, A uint8
}

type ODS struct {
	ObjectID            uint16             // ID of this object
	ObjectVersionNumber uint8              // Version of this object
	LastInSequenceFlag  LastInSequenceFlag // If the image is split into a series of consecutive fragments, the last fragment has this flag set. Possible values:
	ObjectDataLength    uint24             // The length of the Run-length Encoding (RLE) data buffer with the compressed image data.
	Width, Height       uint16             // Dimensions of the image
	// ObjectData	variable	This is the image data compressed using Run-length Encoding (RLE). The size of the data is defined in the Object Data Length field.
}

type uint24 [3]uint8

func (ui uint24) Uint32() uint32 {
	return uint32(ui[0])<<24 | uint32(ui[1])<<16 | uint32(ui[2])
}

type (
	Timestamp          uint32
	SegmentType        uint8
	CompositionState   uint8
	PaletteUpdateFlag  uint8
	ObjectCroppedFlag  uint8
	LastInSequenceFlag uint8
)

const (
	PCSType = 0x16 // Presentation Composition Segment
	WDSType = 0x17 // Window Definition Segment
	PDSType = 0x14 // Palette Definition Segment
	ODSType = 0x15 // Object Definition Segment
	EndType = 0x80 // End of Display Set Segment
)

// The composition state can be one of three values
const (
	// Epoch Start defines a new display. The Epoch Start contains all
	// functional segments needed to display a new composition on the
	// screen.
	CSEpochStart CompositionState = 0x80
	// Acquisition Point defines a display refresh. This is used to
	// compose in the middle of the Epoch. It includes functional segments
	// with new objects to be used in a new composition, replacing old
	// objects with the same Object ID.
	CSAcquisitionPoint = 0x40
	// Normal defines a display update, and contains only functional
	// segments with elements that are different from the preceding
	// composition. It’s mostly used to stop displaying objects on the
	// screen by defining a composition with no composition objects (a
	// value of zero in the Number of Composition Objects flag) but also
	// used to define a new composition with new objects and objects
	// defined since the Epoch Start.
	CSNormal = 0x00
)

const (
	PUFFalse PaletteUpdateFlag = 0x00
	PUFTrue                    = 0x80
)

const (
	CroppedForce ObjectCroppedFlag = 0x40 // Force display of the cropped image object
	CroppedOff                     = 0x00 // Off
)

const (
	LastInSequence         LastInSequenceFlag = 0x40
	FirstInSequence                           = 0x80
	FirstAndLastInSequence                    = 0x40 | 0x80
)

func (st SegmentType) String() string {
	switch st {
	case PCSType:
		return "PCS"
	case WDSType:
		return "WDS"
	case PDSType:
		return "PDS"
	case ODSType:
		return "ODS"
	case EndType:
		return "End"
	}
	return fmt.Sprintf("SegmentType(%x)", uint8(st))
}

func (ts Timestamp) String() string {
	return ts.Duration().String()
}

func (ts Timestamp) Duration() time.Duration {
	// Timestamp has an accuracy of 90 kHz,
	// so divide by 90 to get milliseconds.
	return time.Duration(ts) * time.Millisecond / 90
}

func ReadHeader(r io.Reader) (*Header, error) {
	var h Header
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return nil, err
	}
	return &h, nil
}

func (h *Header) Write(w io.Writer) error {
	return binary.Write(w, binary.BigEndian, h)
}

func (h *Header) Validate() error {
	if h.MagicNumber != 0x5047 {
		return fmt.Errorf(`magic number not "PG" 0x5047: %x`, h.MagicNumber)
	}
	switch h.SegmentType {
	case PCSType, WDSType, PDSType, ODSType, EndType:
	default:
		return fmt.Errorf("unrecognized segment type: 0x%x", h.SegmentType)
	}
	return nil
}

func (pcs *PCS) Validate() error {
	switch pcs.CompositionState {
	case CSNormal, CSAcquisitionPoint, CSEpochStart:
	default:
		return fmt.Errorf("unrecognized composition state: 0x%x", pcs.CompositionState)
	}
	switch pcs.PaletteUpdateFlag {
	case PUFFalse, PUFTrue:
	default:
		return fmt.Errorf("unrecognized palette update flag: 0x%x", pcs.PaletteUpdateFlag)
	}
	return nil
}
