package pgs

import (
	"image/color"
	"time"
)

type Segment struct {
	PresentationTime time.Duration
	DecodingTime     time.Duration
	Data             interface{}
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

type Window struct {
	WindowID      uint8
	X, Y          uint16
	Width, Height uint16
}

type Palette struct {
	ID      uint8
	Version uint8
	Entries []PaletteEntry
}

type PaletteEntry struct {
	ID uint8 // Entry number of the palette
	color.NYCbCrA
}

type Object struct {
	ID          uint16 // ID of this object
	Version     uint8  // Version of this object
	First, Last bool
	Image
}

type Image struct {
	Width, Height uint16 // Dimensions
	Data          []byte
}

type CompositionState uint8

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
	AcquisitionPoint CompositionState = 0x40
	// Normal defines a display update, and contains only functional
	// segments with elements that are different from the preceding
	// composition. It’s mostly used to stop displaying objects on the
	// screen by defining a composition with no composition objects (a
	// value of zero in the Number of Composition Objects flag) but also
	// used to define a new composition with new objects and objects
	// defined since the Epoch Start.
	Normal CompositionState = 0x00
)
