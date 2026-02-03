package sysocr

// BoundingBox represents the location of text in an image.
type BoundingBox struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// TextBlock represents a recognized text block with its location.
type TextBlock struct {
	Text        string
	BoundingBox BoundingBox
}

// Result contains the OCR recognition result.
type Result struct {
	Blocks []TextBlock
	Text   string // All text concatenated
}

// Input specifies the image source. Only one field should be set.
type Input struct {
	FilePath string // Local file path
	URL      string // Remote URL (http/https)
	Data     []byte // In-memory image data
}

// Options configures the OCR recognition.
type Options struct {
	Input     Input
	Languages []string // Optional: language hints (e.g., "zh-Hans", "en")
}
